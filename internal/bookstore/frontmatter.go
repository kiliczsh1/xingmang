package bookstore

import (
	"strings"
)

// FrontMatter 是章节/人物/条目的最小元数据块。
//
// 只做"行级"处理，不做 YAML 序列化 —— 用户在 Obsidian 里加的自定义字段、
// 注释、缩进必须原样保留，任何 round-trip 都可能把格式改坏。
type FrontMatter struct {
	Has   bool
	Lines []string // "---" 之间的原始行（不含围栏）
	Body  string   // 围栏之后的正文
}

// ParseFrontMatter 解析文本开头的 front-matter。
//
// 容错：没有围栏、只有开头没有结尾、或首行不是 "---" 时，一律当纯正文处理，
// 绝不报错 —— 用户的文件可能来自任何编辑器。
func ParseFrontMatter(text string) FrontMatter {
	text = NormalizeText(text)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return FrontMatter{Body: text}
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			body := lines[i+1:]
			// 围栏后通常空一行，跳过它，避免正文永远以 "\n" 开头。
			if len(body) > 0 && strings.TrimSpace(body[0]) == "" {
				body = body[1:]
			}
			return FrontMatter{
				Has:   true,
				Lines: append([]string(nil), lines[1:i]...),
				Body:  strings.Join(body, "\n"),
			}
		}
	}
	// 没有闭合围栏：当正文，避免"吃文件"
	return FrontMatter{Body: text}
}

// Get 读取一个标量字段。找不到返回空串。
func (f FrontMatter) Get(key string) string {
	for _, line := range f.Lines {
		if !strings.HasPrefix(line, key+":") {
			continue
		}
		v := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
		v = strings.Trim(v, `"'`)
		return v
	}
	return ""
}

// WithField 返回修改后的副本：value 为空则删除该键，否则替换/追加。
func (f FrontMatter) WithField(key, value string) FrontMatter {
	out := FrontMatter{Has: f.Has, Lines: append([]string(nil), f.Lines...), Body: f.Body}
	idx := -1
	for i, line := range out.Lines {
		if strings.HasPrefix(line, key+":") {
			idx = i
			break
		}
	}

	switch {
	case value == "":
		if idx >= 0 {
			out.Lines = append(out.Lines[:idx], out.Lines[idx+1:]...)
		}
	default:
		newLine := key + ": " + value
		if idx >= 0 {
			out.Lines[idx] = newLine
		} else {
			out.Lines = append(out.Lines, newLine)
		}
		// 原本没有 front-matter 时，写入第一个字段就要把围栏打开。
		out.Has = true
	}

	if len(out.Lines) == 0 {
		out.Has = false
	}
	return out
}

// Render 还原成完整文本。没有字段时不写围栏，保持"纯 Markdown 笔记"。
func (f FrontMatter) Render() string {
	body := f.Body
	if !f.Has || len(f.Lines) == 0 {
		return body
	}
	var b strings.Builder
	b.WriteString("---\n")
	for _, line := range f.Lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	if body != "" {
		if !strings.HasPrefix(body, "\n") {
			b.WriteString("\n")
		}
		b.WriteString(body)
	}
	return b.String()
}

// SetFrontMatterField 是"读-改-写"的便捷入口：只动一个键，其余原样保留。
func SetFrontMatterField(text, key, value string) string {
	return ParseFrontMatter(text).WithField(key, value).Render()
}

// BodyOf 返回去掉 front-matter 后的正文。
func BodyOf(text string) string {
	return ParseFrontMatter(text).Body
}
