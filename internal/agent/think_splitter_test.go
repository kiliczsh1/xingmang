package agent

import "testing"

// 模拟一次完整的流式拼接：把 chunks 依次喂给 splitter，返回（回答, 思考）。
func runSplitter(chunks []string) (string, string) {
	var s ThinkSplitter
	var answer, think string
	for _, c := range chunks {
		a, th := s.Feed(c)
		answer += a
		think += th
	}
	a, th := s.Flush()
	return answer + a, think + th
}

func TestThinkSplitter(t *testing.T) {
	cases := []struct {
		name    string
		chunks  []string
		answer  string
		think   string
	}{
		{
			name:   "纯回答没有思考块",
			chunks: []string{"你好，", "我是星芒。"},
			answer: "你好，我是星芒。",
		},
		{
			name:   "单个分片内含完整思考块",
			chunks: []string{"<think>先分析一下</think>正文回答"},
			answer: "正文回答",
			think:  "先分析一下",
		},
		{
			name:   "标签被切在分片之间（开标签）",
			chunks: []string{"<thi", "nk>思考", "</think>答案"},
			answer: "答案",
			think:  "思考",
		},
		{
			name:   "标签被切在分片之间（闭标签）",
			chunks: []string{"<think>思考</thi", "nk>答案"},
			answer: "答案",
			think:  "思考",
		},
		{
			name:   "未闭合的思考块全部算思考",
			chunks: []string{"<think>想了一半就断了"},
			answer: "",
			think:  "想了一半就断了",
		},
		{
			name:   "思考与回答交替出现",
			chunks: []string{"<think>A</think>回答1<thinking>B</thinking>回答2"},
			answer: "回答1回答2",
			think:  "AB",
		},
		{
			name:   "不完整的标签名不算标签",
			chunks: []string{"if (a<b) { echo \"<think\" }"},
			answer: "if (a<b) { echo \"<think\" }",
			think:  "",
		},
	}
	for _, c := range cases {
		gotAnswer, gotThink := runSplitter(c.chunks)
		if gotAnswer != c.answer || gotThink != c.think {
			t.Errorf("%s:\n  answer got=%q want=%q\n  think  got=%q want=%q",
				c.name, gotAnswer, c.answer, gotThink, c.think)
		}
	}
}
