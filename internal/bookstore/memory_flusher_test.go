package bookstore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 合并写入是记忆落盘的核心：同一批抖动必须只落一次盘，失败必须能重试。

func TestFlusher_MergesWrites(t *testing.T) {
	f := NewFlusher(time.Hour) // 不启动 ticker，手动 flush
	var calls int32
	f.Register(MemoryKindChat, func(key string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	for i := 0; i < 5; i++ {
		f.Mark(MemoryKindChat, "7") // 同一条被标记 5 次
	}
	f.Mark(MemoryKindChat, "8")

	if got := f.Pending(); got != 2 {
		t.Fatalf("待写入应为 2 条（去重后），实得 %d", got)
	}
	written, err := f.FlushAll()
	if err != nil {
		t.Fatal(err)
	}
	if written != 2 || atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("同一批应只写 2 次，实得 written=%d calls=%d", written, atomic.LoadInt32(&calls))
	}
	if got := f.Pending(); got != 0 {
		t.Fatalf("flush 后不应有待写入，实得 %d", got)
	}
	if _, err := f.FlushAll(); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("没有新标记时不应重复写盘，calls=%d", atomic.LoadInt32(&calls))
	}
}

func TestFlusher_RetriesThenGivesUp(t *testing.T) {
	f := NewFlusher(time.Hour)
	var calls int32
	f.Register(MemoryKindGraph, func(key string) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("磁盘被占用")
	})
	f.Mark(MemoryKindGraph, "1")

	if _, err := f.FlushAll(); err == nil {
		t.Fatal("第一次失败应返回错误")
	}
	if f.Pending() != 1 {
		t.Fatal("失败后应保留脏标记等待重试")
	}
	if _, err := f.FlushAll(); err == nil {
		t.Fatal("第二次失败应返回错误")
	}
	if f.Pending() != 1 {
		t.Fatal("未达上限前应继续重试")
	}
	if _, err := f.FlushAll(); err == nil {
		t.Fatal("第三次失败应返回错误")
	}
	if f.Pending() != 0 {
		t.Fatal("达到重试上限后应放弃，避免无限刷盘")
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("应恰好尝试 3 次，实得 %d", atomic.LoadInt32(&calls))
	}
}

func TestFlusher_KeepsMarkArrivedDuringRender(t *testing.T) {
	f := NewFlusher(time.Hour)
	// 渲染过程中又来了一个 Mark：不能被"写成功后删除"吞掉。
	f.Register(MemoryKindAgent, func(key string) error {
		f.Mark(MemoryKindAgent, key)
		return nil
	})
	f.Mark(MemoryKindAgent, "3")
	if _, err := f.FlushAll(); err != nil {
		t.Fatal(err)
	}
	if f.Pending() != 1 {
		t.Fatalf("渲染期间到达的标记必须保留，实得 pending=%d", f.Pending())
	}
}

func TestFlusher_UnknownKind(t *testing.T) {
	f := NewFlusher(time.Hour)
	f.Mark("nope", "1")
	if _, err := f.FlushAll(); err == nil {
		t.Fatal("未注册渲染器应报错而不是静默丢弃")
	}
}

func TestFlusher_TickerAndStop(t *testing.T) {
	f := NewFlusher(20 * time.Millisecond)
	var mu sync.Mutex
	var n int
	f.Register(MemoryKindAnalysis, func(key string) error {
		mu.Lock()
		n++
		mu.Unlock()
		return nil
	})
	f.Start()
	f.Mark(MemoryKindAnalysis, "1")

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		done := n > 0
		mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("后台 ticker 没能在 2 秒内写盘")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Stop 时把剩余的脏项也写出去
	f.Mark(MemoryKindAnalysis, "2")
	written, err := f.Stop(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if written != 1 {
		t.Fatalf("Stop 应把剩余的 1 条写出去，实得 %d", written)
	}
	if f.Pending() != 0 {
		t.Fatal("Stop 之后不应还有待写入")
	}
}

// 放弃写入必须留下原因。
//
// 之前重试三次就静默丢掉：界面显示"待写入 0 项"，看起来一切正常，
// 而磁盘上的文件其实停在旧内容 —— 这正是本项目最不能接受的一种失败。
func TestFlusherReportsGivenUpWrites(t *testing.T) {
	f := NewFlusher(10 * time.Millisecond)
	f.Register(MemoryKindGraph, func(key string) error {
		return fmt.Errorf("磁盘只读")
	})
	f.Mark(MemoryKindGraph, "7")

	for i := 0; i < 3; i++ {
		if _, err := f.FlushAll(); err == nil {
			t.Fatalf("第 %d 次应返回错误", i+1)
		}
	}
	if n := f.Pending(); n != 0 {
		t.Fatalf("放弃后待写入应为 0，实得 %d", n)
	}
	fails := f.Failures()
	if len(fails) != 1 {
		t.Fatalf("放弃的条目必须被记下来，实得 %v", fails)
	}
	var reason string
	for _, v := range fails {
		reason = v
	}
	if !strings.Contains(reason, "磁盘只读") {
		t.Fatalf("失败原因应保留: %q", reason)
	}

	// 重新标脏并成功写出后，失败记录应清掉
	f.Register(MemoryKindGraph, func(key string) error { return nil })
	f.Mark(MemoryKindGraph, "7")
	if _, err := f.FlushAll(); err != nil {
		t.Fatal(err)
	}
	if len(f.Failures()) != 0 {
		t.Fatalf("成功写出后失败记录应清空，实得 %v", f.Failures())
	}
}
