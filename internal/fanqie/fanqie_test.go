package fanqie

import (
	"encoding/json"
	"os"
	"testing"
)

// SM3 测试期望值取自 fanqie-rank-mcp 参考实现（a_bogus.py），
// 注意：参考实现的压缩公式与国密标准有细微偏离（TT2 用 SS2+rotl(A,12) 替代 SS1），
// 服务端不做 SM3 校验，签名只要求与参考实现逐字节一致，故以参考实现输出为基准。
func TestSM3ReferenceVectors(t *testing.T) {
	var h sm3
	if got := h.sumHex("abc"); got != "5f99119d900874d41965297e112f4220cb48bce3e6dd58c99ecad05ca824f09d" {
		t.Fatalf("SM3(\"abc\") = %s, want 参考实现向量", got)
	}
	if got := h.sumHex(""); got != "e5e663b75ae3e9373e95244d596a442311cc2647719510cc4bd0b0f9c8de77cb" {
		t.Fatalf("SM3(\"\") = %s, want 参考实现向量", got)
	}
}

// 多块输入（>64 字节）确保压缩循环正确
func TestSM3LongInput(t *testing.T) {
	var h sm3
	long := "app_id=2503&rank_list_type=3&offset=0&limit=30&category_id=1141&rank_version=&gender=1&rankMold=2"
	got := h.sumHex(long)
	want := "92a1dc5623edd3be99194a18f5fdf6164fceb2e9cdfd731b20ebfce38c9e7e73" // 参考实现实测值
	if got != want {
		t.Fatalf("SM3(long) = %s, want %s", got, want)
	}
}

func TestDecryptPUA(t *testing.T) {
	// 58670→'0'、58611→'的'；未收录的 PUA 字符保留原样
	if got := decryptPUA(string([]rune{58670, 'x', 58611})); got != "0x的" {
		t.Fatalf("decryptPUA = %q, want %q", got, "0x的")
	}
	if got := decryptPUA("普通文本"); got != "普通文本" {
		t.Fatalf("decryptPUA(普通) = %q", got)
	}
	// 0xF8FF 之外的 PUA 区间不处理
	if got := decryptPUA(string([]rune{0xE000})); got != string([]rune{0xE000}) {
		t.Fatalf("decryptPUA(0xE000) = %q, want 原样", got)
	}
}

func TestNormalizeCoverURL(t *testing.T) {
	cases := map[string]string{
		"":                           "",
		"//p3-novel.byteimg.com/a":   "https://p3-novel.byteimg.com/a",
		"https://example.com/x.jpg":  "https://example.com/x.jpg",
		"http://example.com/x.jpg":   "http://example.com/x.jpg",
		"/novel-images/cover/123":    "https://p3-novel.byteimg.com/novel-images/cover/123",
	}
	for in, want := range cases {
		if got := normalizeCoverURL(in); got != want {
			t.Fatalf("normalizeCoverURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResultEncrypt(t *testing.T) {
	// 固定输入验证自定义 Base64（与 Python result_encrypt 一致）：
	// {1,2,3} = 0x010203 → 索引 (0,16,8,3) → s4 表 "Dfmp"
	if got := resultEncrypt([]byte{1, 2, 3}, "s4"); got != "Dfmp" {
		t.Fatalf("resultEncrypt = %q, want %q", got, "Dfmp")
	}
}

// TestLiveFanqie 联网验证：真实请求番茄榜单接口（默认跳过）。
// 运行：FANQIE_LIVE=1 go test ./internal/fanqie/ -run TestLive -v
func TestLiveFanqie(t *testing.T) {
	if os.Getenv("FANQIE_LIVE") != "1" {
		t.Skip("设置 FANQIE_LIVE=1 运行联网验证")
	}

	groupsBody, err := RankingsJSON()
	if err != nil {
		t.Fatalf("RankingsJSON: %v", err)
	}
	var groupsPayload struct {
		Groups []groupJSON `json:"groups"`
	}
	if err := json.Unmarshal(groupsBody, &groupsPayload); err != nil {
		t.Fatalf("解析 groups: %v", err)
	}
	if len(groupsPayload.Groups) == 0 {
		t.Fatal("groups 为空（可能走了兜底数据）")
	}
	t.Logf("榜单分组: %d 组，首个: %s", len(groupsPayload.Groups), groupsPayload.Groups[0].Title)

	booksBody, err := BooksJSON("1", "2", "1141", 0, 10)
	if err != nil {
		t.Fatalf("BooksJSON: %v", err)
	}
	var booksPayload struct {
		Books []bookJSON `json:"books"`
	}
	if err := json.Unmarshal(booksBody, &booksPayload); err != nil {
		t.Fatalf("解析 books: %v", err)
	}
	if len(booksPayload.Books) == 0 {
		t.Fatal("books 为空")
	}
	first := booksPayload.Books[0]
	if first.Title == "" || first.Author == "" || first.BookID == "" {
		t.Fatalf("首本书字段缺失: %+v", first)
	}
	t.Logf("首本书: #%d《%s》%s (ID=%s)", first.Rank, first.Title, first.Author, first.BookID)

	// 回归：gender=0（女频）是合法值，不得被默认值改写（Atoi("0") 陷阱）
	femaleBody, err := BooksJSON("0", "2", "1139", 0, 3)
	if err != nil {
		t.Fatalf("BooksJSON(女频): %v", err)
	}
	var femalePayload struct {
		Books []bookJSON `json:"books"`
	}
	if err := json.Unmarshal(femaleBody, &femalePayload); err != nil {
		t.Fatalf("解析女频 books: %v", err)
	}
	if len(femalePayload.Books) == 0 {
		t.Fatal("女频 gender=0 返回空（gender 被错误改写成男频？）")
	}
	t.Logf("女频首本: 《%s》%s", femalePayload.Books[0].Title, femalePayload.Books[0].Author)
}
