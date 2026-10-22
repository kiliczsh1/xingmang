package bookstore

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// 记忆落盘的合并写入器（docs/books-memory-storage.md §8.1）。
//
// 为什么需要它：图谱一次编辑要序列化上千个节点，如果每次写操作都同步渲染落盘，
// "删一条关系"就会变成 200ms 的接口。所以写操作只标记脏，由一个后台 goroutine
// 按固定间隔把同一批脏项合并写出去。
//
// 失败策略：保留脏标记，下个周期重试；连续失败达到上限后放弃（避免无限重试刷日志），
// 由用户在界面上点「重建索引」处理。
const (
	// DefaultFlushInterval 是合并落盘的默认间隔。
	DefaultFlushInterval = 2 * time.Second
	// defaultMaxTries 是单条记忆连续写盘失败的放弃阈值。
	defaultMaxTries = 3
)

// MemoryRenderer 把一条记录的最新状态渲染落盘；返回错误表示这次没写成功。
//
// key 是记录标识：对话/图谱传数字 id 的十进制串，Agent 会话传会话 id 本身。
// 用字符串而不是 int 是因为会话 id 本来就是字符串，折成整数会不可逆。
type MemoryRenderer func(key string) error

type flushEntry struct {
	key      string // kind + NUL + ident，用于去重
	kind     string
	ident    string // 记录标识（对话 id / 会话 id）
	attempts int    // 已连续失败次数
	dirty    bool
}

// Flusher 是记忆文件的合并写入器。零值不可用，请用 NewFlusher。
type Flusher struct {
	interval time.Duration
	maxTries int

	mu      sync.Mutex
	entries map[string]*flushEntry
	render  map[string]MemoryRenderer
	// failures 记录放弃写入的条目及原因。
	//
	// 之前是"重试三次就丢掉"，日志之外没有任何痕迹 —— 用户看到的是"待写入 0 项"，
	// 而文件其实是旧的。既然项目规则是如实报告，失败就必须留在能被读到的位置。
	failures map[string]string
	started  bool

	flushing sync.Mutex // 保证同一时刻只有一次 FlushAll 在写盘

	stopOnce sync.Once
	stopped  chan struct{}
	done     chan struct{}
}

// NewFlusher 创建写入器；interval <= 0 时用默认值。
func NewFlusher(interval time.Duration) *Flusher {
	if interval <= 0 {
		interval = DefaultFlushInterval
	}
	return &Flusher{
		interval: interval,
		maxTries: defaultMaxTries,
		entries:  map[string]*flushEntry{},
		render:   map[string]MemoryRenderer{},
		failures: map[string]string{},
	}
}

// flushKey 用 NUL 连接 kind 与记录标识：两者都可能含普通分隔符，
// 只有一个不会出现在会话 id 里的字节才能保证不歧义。
func flushKey(kind, key string) string {
	return strings.ToLower(kind) + string(rune(0)) + key
}

// Register 注册某类记忆的渲染函数（M1–M4 各自注册）。
func (f *Flusher) Register(kind string, fn MemoryRenderer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.render[strings.ToLower(kind)] = fn
}

// Mark 标记一条记录需要落盘。重复标记会被合并成一次写入。
func (f *Flusher) Mark(kind, key string) {
	kind = strings.ToLower(kind)
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	full := flushKey(kind, key)
	f.mu.Lock()
	defer f.mu.Unlock()
	if e, ok := f.entries[full]; ok {
		e.dirty = true
		return
	}
	f.entries[full] = &flushEntry{key: full, kind: kind, ident: key, dirty: true}
}

// Failures 返回放弃写入的条目与原因（kind+标识 → 原因）。
//
// 这些条目不会自动重试，需要用户在界面上点「从文件重建索引」或直接改回来。
func (f *Flusher) Failures() map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]string, len(f.failures))
	for k, v := range f.failures {
		// 键里的 NUL 不适合直接展示，换成 "kind ident"
		parts := strings.SplitN(k, string(rune(0)), 2)
		if len(parts) == 2 {
			out[parts[0]+" "+parts[1]] = v
		} else {
			out[k] = v
		}
	}
	return out
}

// Pending 返回当前待写入的条数（含等待重试的）。
func (f *Flusher) Pending() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, e := range f.entries {
		if e.dirty {
			n++
		}
	}
	return n
}

// FlushAll 立刻把当前所有脏项写出去，返回成功条数与第一个错误。
func (f *Flusher) FlushAll() (int, error) {
	f.flushing.Lock()
	defer f.flushing.Unlock()

	f.mu.Lock()
	if len(f.entries) == 0 {
		f.mu.Unlock()
		return 0, nil
	}
	var jobs []*flushEntry
	for _, e := range f.entries {
		if !e.dirty {
			continue
		}
		// 先清脏标记：渲染期间新来的 Mark 会重新置 true，不会丢。
		e.dirty = false
		jobs = append(jobs, &flushEntry{key: e.key, kind: e.kind, ident: e.ident, attempts: e.attempts})
	}
	f.mu.Unlock()

	written := 0
	var firstErr error
	for _, job := range jobs {
		f.mu.Lock()
		fn := f.render[job.kind]
		f.mu.Unlock()
		var err error
		if fn == nil {
			err = fmt.Errorf("未注册记忆渲染器: %s", job.kind)
		} else {
			err = fn(job.ident)
		}

		f.mu.Lock()
		cur, alive := f.entries[job.key]
		if err == nil {
			written++
			delete(f.failures, job.key)
			// 渲染期间没有新的标记，才真正移除；否则留给下一轮。
			if alive && !cur.dirty {
				delete(f.entries, job.key)
			}
			f.mu.Unlock()
			continue
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("%s %s: %w", job.kind, job.ident, err)
		}
		if !alive {
			// 期间被标记过又被处理掉的极端情况：重新登记，避免静默丢失。
			cur = &flushEntry{key: job.key, kind: job.kind, ident: job.ident}
			f.entries[job.key] = cur
		}
		cur.attempts = job.attempts + 1
		if cur.attempts >= f.maxTries {
			// 放弃重试，但把原因留下来：否则界面上"待写入"归零，看起来一切正常，
			// 而磁盘上那份文件其实停在了旧内容。
			f.failures[job.key] = fmt.Sprintf("%s %s: %v", job.kind, job.ident, err)
			delete(f.entries, job.key)
		} else {
			cur.dirty = true // 下个周期重试
		}
		f.mu.Unlock()
	}
	return written, firstErr
}

// Start 启动后台合并写入。重复调用无副作用。
func (f *Flusher) Start() {
	f.mu.Lock()
	if f.started {
		f.mu.Unlock()
		return
	}
	f.started = true
	f.stopped = make(chan struct{})
	f.done = make(chan struct{})
	f.mu.Unlock()
	go f.loop()
}

func (f *Flusher) loop() {
	defer close(f.done)
	t := time.NewTicker(f.interval)
	defer t.Stop()
	for {
		select {
		case <-f.stopped:
			return
		case <-t.C:
			f.FlushAll()
		}
	}
}

// Stop 停止后台写入并做最后一次 flush。
//
// timeout 是等待后台 goroutine 退出的上限；超时也仍会尝试 flush（数据比优雅退出重要）。
func (f *Flusher) Stop(timeout time.Duration) (int, error) {
	f.mu.Lock()
	started, stopped, done := f.started, f.stopped, f.done
	f.mu.Unlock()

	if started {
		f.stopOnce.Do(func() { close(stopped) })
		select {
		case <-done:
		case <-time.After(timeout):
		}
	}
	return f.FlushAll()
}

// ---------------- 包级单例 ----------------
//
// 处理函数只需要 MarkMemoryDirty(kind, id)，不必关心 Flusher 实例的生命周期。

var (
	globalFlusherMu sync.Mutex
	globalFlusher   *Flusher
)

// DefaultFlusher 返回进程级写入器（懒创建，Mark 在 Start 之前调用也安全）。
func DefaultFlusher() *Flusher {
	globalFlusherMu.Lock()
	defer globalFlusherMu.Unlock()
	if globalFlusher == nil {
		globalFlusher = NewFlusher(DefaultFlushInterval)
	}
	return globalFlusher
}

// RegisterMemoryRenderer 注册某类记忆的渲染函数。
func RegisterMemoryRenderer(kind string, fn MemoryRenderer) { DefaultFlusher().Register(kind, fn) }

// MarkMemoryDirty 标记一条记忆需要落盘（写操作后调用，不返回错误、不阻塞）。
func MarkMemoryDirty(kind, key string) { DefaultFlusher().Mark(kind, key) }

// StartMemoryFlusher 在服务启动时调用一次。
func StartMemoryFlusher() { DefaultFlusher().Start() }

// StopMemoryFlusher 在服务退出时调用，做最后一次 flush。
func StopMemoryFlusher(timeout time.Duration) (int, error) { return DefaultFlusher().Stop(timeout) }

// PendingMemoryWrites 返回待写入条数，供摘要接口展示"待同步 N 项"。
func PendingMemoryWrites() int { return DefaultFlusher().Pending() }

// MemoryWriteFailures 返回放弃写入的条目，供摘要接口如实展示。
func MemoryWriteFailures() map[string]string { return DefaultFlusher().Failures() }
