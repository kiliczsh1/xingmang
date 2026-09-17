package fanqie

// a_bogus 反爬签名 —— 从 fanqie-rank-mcp/a_bogus.py 1:1 移植。
//
// 组成：RC4 流密码 + SM3 哈希 + 自定义 Base64 + 时间戳/环境指纹 blob。
// 与 Python 原版保持逐字节一致（含 lossy UTF-8 处理），请勿"优化"。

import (
	"math/rand"
	"strings"
	"time"
)

// ─── RC4 Stream Cipher ────────────────────────────────────────────────

func rc4Encrypt(plaintext, key []byte) []byte {
	s := make([]byte, 256)
	for i := range s {
		s[i] = byte(i)
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) & 0xFF
		s[i], s[j] = s[j], s[i]
	}

	ii, jj := 0, 0
	out := make([]byte, len(plaintext))
	for n, b := range plaintext {
		ii = (ii + 1) & 0xFF
		jj = (jj + int(s[ii])) & 0xFF
		s[ii], s[jj] = s[jj], s[ii]
		t := (int(s[ii]) + int(s[jj])) & 0xFF
		out[n] = s[t] ^ b
	}
	return out
}

// ─── 自定义 Base64 编码 ───────────────────────────────────────────────

var encodeTables = map[string]string{
	"s0": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",
	"s1": "Dkdpgh4ZKsQB80/Mfvw36XI1R25+WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
	"s2": "Dkdpgh4ZKsQB80/Mfvw36XI1R25-WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
	"s3": "ckdp1h4ZKsUB80/Mfvw36XIgR25+WQAlEi7NLboqYTOPuzmFjJnryx9HVGDaStCe",
	"s4": "Dkdpgh2ZmsQB80/MfvV36XI1R45-WUAlEixNLwoqYTOPuzKFjJnry79HbGcaStCe",
}

// resultEncrypt 自定义 Base64（3 字节 → 4 字符），末尾不足 3 字节的余数忽略（与原版一致）。
func resultEncrypt(data []byte, tableName string) string {
	table := encodeTables[tableName]
	if table == "" {
		table = encodeTables["s0"]
	}
	masks := [3]int{0xFC0000, 0x03F000, 0x000FC0}

	var sb strings.Builder
	chunks := len(data) / 3
	for r := 0; r < chunks; r++ {
		i := r * 3
		longInt := int(data[i])<<16 | int(data[i+1])<<8 | int(data[i+2])
		temp := [4]int{
			(longInt & masks[0]) >> 18,
			(longInt & masks[1]) >> 12,
			(longInt & masks[2]) >> 6,
			longInt & 63,
		}
		for _, t := range temp {
			sb.WriteByte(table[t])
		}
	}
	return sb.String()
}

// ─── 随机字符串生成 ────────────────────────────────────────────────────

func generRandom(randomVal int, opt [2]int) []byte {
	return []byte{
		byte(randomVal&0xFF&0xAA | opt[0]&0x55),
		byte(randomVal&0xFF&0x55 | opt[0]&0xAA),
		byte(randomVal>>8&0xFF&0xAA | opt[1]&0x55),
		byte(randomVal>>8&0xFF&0x55 | opt[1]&0xAA),
	}
}

// generateRandomStr 生成 12 字节随机串（等价于 int(random.random()*10000)）。
func generateRandomStr() []byte {
	out := make([]byte, 0, 12)
	out = append(out, generRandom(rand.Intn(10000), [2]int{3, 45})...)
	out = append(out, generRandom(rand.Intn(10000), [2]int{1, 0})...)
	out = append(out, generRandom(rand.Intn(10000), [2]int{1, 5})...)
	return out
}

// ─── RC4 加密的二进制 blob 构造 ──────────────────────────────────────

const windowEnvStr = "1536|747|1536|834|0|30|0|0|1536|834|1536|864|1525|747|24|24|Win32"

func generateRC4BBStr(urlSearchParams, userAgent string) []byte {
	var h sm3

	startTimeMS := uint32(time.Now().UnixMilli())

	// SM3 哈希 URL 参数
	urlHashHex := h.sumHex(urlSearchParams)
	urlSearchParamsList := h.sumBytes(urlHashHex)

	// SM3 哈希后缀 "cus"
	cusHashHex := h.sumHex("cus")
	cusList := h.sumBytes(cusHashHex)

	// RC4 加密 UA + 编码 + 哈希
	rc4Key := []byte{0x01, 0x00, 0x0E}
	uaBytes := []byte(userAgent) // 原版 latin-1 编码，UA 为纯 ASCII
	rc4EncryptedUA := rc4Encrypt(uaBytes, rc4Key)
	uaBase64 := resultEncrypt(rc4EncryptedUA, "s3")
	uaList := h.sumBytesFromBytes([]byte(uaBase64))

	endTimeMS := uint32(time.Now().UnixMilli())

	// 构造 b 表（仅低 8 位被输出，按原版逻辑保留全部赋值）
	b := make([]uint32, 256)
	b[8] = 3
	b[10] = endTimeMS
	b[16] = startTimeMS
	b[18] = 44

	b[20] = (b[16] >> 24) & 0xFF
	b[21] = (b[16] >> 16) & 0xFF
	b[22] = (b[16] >> 8) & 0xFF
	b[23] = b[16] & 0xFF
	b[24] = 0
	b[25] = 0

	// args = (0, 1, 14)
	b[26] = 0
	b[27] = 0
	b[28] = 0
	b[29] = 0

	b[30] = 0
	b[31] = 1
	b[32] = 0
	b[33] = 0
	b[34] = 0
	b[35] = 0
	b[36] = 0
	b[37] = 14

	// URL 哈希的第 22/23 字节
	b[38] = uint32(urlSearchParamsList[21])
	b[39] = uint32(urlSearchParamsList[22])

	// cus 哈希的第 22/23 字节
	b[40] = uint32(cusList[21])
	b[41] = uint32(cusList[22])

	// UA 哈希的第 24/25 字节
	b[42] = uint32(uaList[23])
	b[43] = uint32(uaList[24])

	// 结束时间戳
	b[44] = (b[10] >> 24) & 0xFF
	b[45] = (b[10] >> 16) & 0xFF
	b[46] = (b[10] >> 8) & 0xFF
	b[47] = b[10] & 0xFF
	b[48] = b[8]
	b[49] = 0
	b[50] = 0

	// page_id 和 aid
	pageID := 6241
	aid := 6383

	b[51] = uint32(pageID)
	b[52] = uint32(pageID >> 24 & 0xFF)
	b[53] = uint32(pageID >> 16 & 0xFF)
	b[54] = uint32(pageID >> 8 & 0xFF)
	b[55] = uint32(pageID & 0xFF)

	b[56] = uint32(aid)
	b[57] = uint32(aid & 0xFF)
	b[58] = uint32(aid >> 8 & 0xFF)
	b[59] = uint32(aid >> 16 & 0xFF)
	b[60] = uint32(aid >> 24 & 0xFF)

	// 窗口环境字符串长度
	windowEnvBytes := []byte(windowEnvStr)
	envLen := len(windowEnvBytes)
	b[64] = uint32(envLen)
	b[65] = uint32(envLen & 0xFF)
	b[66] = uint32(envLen >> 8 & 0xFF)

	b[69] = 0
	b[70] = 0
	b[71] = 0

	// 校验和
	b[72] = b[18] ^ b[20] ^ b[26] ^ b[30] ^ b[38] ^ b[40] ^ b[42] ^
		b[21] ^ b[27] ^ b[31] ^ b[35] ^ b[39] ^ b[41] ^ b[43] ^
		b[22] ^ b[28] ^ b[32] ^ b[36] ^
		b[23] ^ b[29] ^ b[33] ^ b[37] ^
		b[44] ^ b[45] ^ b[46] ^ b[47] ^ b[48] ^ b[49] ^ b[50] ^
		b[24] ^ b[25] ^
		b[52] ^ b[53] ^ b[54] ^ b[55] ^
		b[57] ^ b[58] ^ b[59] ^ b[60] ^
		b[65] ^ b[66] ^
		b[70] ^ b[71]

	// 按顺序输出 blob
	bbOrder := [...]int{
		18, 20, 52, 26, 30, 34, 58, 38, 40, 53, 42, 21, 27, 54, 55, 31,
		35, 57, 39, 41, 43, 22, 28, 32, 60, 36, 23, 29, 33, 37, 44, 45,
		59, 46, 47, 48, 49, 50, 24, 25, 65, 66, 70, 71,
	}

	bb := make([]byte, 0, len(bbOrder)+len(windowEnvBytes)+1)
	for _, idx := range bbOrder {
		bb = append(bb, byte(b[idx]&0xFF))
	}
	bb = append(bb, windowEnvBytes...)
	bb = append(bb, byte(b[72]&0xFF))

	// 原版 Rust from_utf8_lossy 等效：无效 UTF-8 字节替换为 U+FFFD 后重新编码
	bbStr := strings.ToValidUTF8(string(bb), "\uFFFD")

	return rc4Encrypt([]byte(bbStr), []byte{121})
}

// ─── 主入口 ───────────────────────────────────────────────────────────

// generateABogus 生成番茄榜单接口的 a_bogus 反爬签名。
func generateABogus(urlSearchParams, userAgent string) string {
	randomStr := generateRandomStr()
	rc4BB := generateRC4BBStr(urlSearchParams, userAgent)

	combined := append(randomStr, rc4BB...)
	return resultEncrypt(combined, "s4") + "="
}
