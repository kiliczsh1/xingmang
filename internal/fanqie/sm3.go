package fanqie

// SM3 国密哈希 —— 从 fanqie-rank-mcp/a_bogus.py 移植（仅用于生成番茄榜单 a_bogus 签名）。
//
// 注意：本实现刻意复刻 Python 原版的全部细节（包括其填充逻辑与
// bytes().decode('utf-8', errors='replace') 的 lossy 行为），
// 以保持生成的签名与线上验证通过的原版逐字节一致，不要"修正"任何行为。

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"strings"
)

func rotl32(x uint32, n int) uint32 { return bits.RotateLeft32(x, n) }

func sm3FF(j int, x, y, z uint32) uint32 {
	if j < 16 {
		return x ^ y ^ z
	}
	return (x & y) | (x & z) | (y & z)
}

func sm3GG(j int, x, y, z uint32) uint32 {
	if j < 16 {
		return x ^ y ^ z
	}
	return (x & y) | (^x & z)
}

func sm3TJ(j int) uint32 {
	if j < 16 {
		return 0x79CC4519
	}
	return 0x7A879D8A
}

func sm3P0(x uint32) uint32 { return x ^ rotl32(x, 9) ^ rotl32(x, 17) }

func sm3P1(x uint32) uint32 { return x ^ rotl32(x, 15) ^ rotl32(x, 23) }

func sm3CompressBlock(reg [8]uint32, block []byte) [8]uint32 {
	var w [132]uint32

	// 消息扩展：W0 ~ W15
	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(block[i*4:])
	}

	// 消息扩展：W16 ~ W67
	for n := 16; n < 68; n++ {
		a := w[n-16] ^ w[n-9] ^ rotl32(w[n-3], 15)
		a = a ^ rotl32(a, 15) ^ rotl32(a, 23)
		w[n] = a ^ rotl32(w[n-13], 7) ^ w[n-6]
	}

	// 消息扩展：W'0 ~ W'63
	for n := 0; n < 64; n++ {
		w[n+68] = w[n] ^ w[n+4]
	}

	// 64 轮压缩
	v := reg
	for c := 0; c < 64; c++ {
		ss1 := rotl32(v[0], 12) + v[4] + rotl32(sm3TJ(c), c)
		ss1 = rotl32(ss1, 7) ^ rotl32(v[0], 12)
		tt1 := sm3FF(c, v[0], v[1], v[2]) + v[3] + ss1 + w[c+68]
		tt2 := sm3GG(c, v[4], v[5], v[6]) + v[7] + (ss1 + rotl32(v[0], 12)) + w[c]
		v[3] = v[2]
		v[2] = rotl32(v[1], 9)
		v[1] = v[0]
		v[0] = tt1
		v[7] = v[6]
		v[6] = rotl32(v[5], 19)
		v[5] = v[4]
		v[4] = sm3P0(tt2)
	}

	var result [8]uint32
	for i := 0; i < 8; i++ {
		result[i] = reg[i] ^ v[i]
	}
	return result
}

// sm3 是仅本项目需要的极简 SM3 流式实现（接口对齐 a_bogus.py 的 SM3 类）。
type sm3 struct {
	reg   [8]uint32
	chunk []byte
	size  int
}

func (h *sm3) reset() {
	h.reg = [8]uint32{
		0x7380166F, 0x4914B2B9, 0x172442D7, 0xDA8A0600,
		0xA96F30BC, 0x163138AA, 0xE38DEE4D, 0xB0FB0E4E,
	}
	h.chunk = h.chunk[:0]
	h.size = 0
}

// urlEncodeBytes 与 Python 原版 _url_encode_bytes 一致：
// 保留字母数字与 " -_.~"，其余字节百分号编码（大写十六进制）。
func urlEncodeBytes(s string) []byte {
	const hexChars = "0123456789ABCDEF"
	var out []byte
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= '0' && ch <= '9', ch >= 'A' && ch <= 'Z', ch >= 'a' && ch <= 'z',
			ch == ' ', ch == '-', ch == '_', ch == '.', ch == '~':
			out = append(out, ch)
		default:
			out = append(out, '%', hexChars[ch>>4], hexChars[ch&0x0F])
		}
	}
	return out
}

func (h *sm3) writeStr(s string) { h.writeBytes(urlEncodeBytes(s)) }

func (h *sm3) writeBytes(data []byte) {
	h.size += len(data)
	if len(h.chunk)+len(data) < 64 {
		h.chunk = append(h.chunk, data...)
		return
	}

	// 填满当前块
	fill := 64 - len(h.chunk)
	h.chunk = append(h.chunk, data[:fill]...)
	h.reg = sm3CompressBlock(h.reg, h.chunk)
	h.chunk = h.chunk[:0]

	offset := fill
	for offset+64 <= len(data) {
		h.reg = sm3CompressBlock(h.reg, data[offset:offset+64])
		offset += 64
	}
	if offset < len(data) {
		h.chunk = append(h.chunk, data[offset:]...)
	}
}

// fill 与 Python 原版 _fill 完全一致（含其不规范的边界行为，勿改）。
func (h *sm3) fill() {
	bitLen := uint64(h.size) * 8
	h.chunk = append(h.chunk, 0x80)
	for len(h.chunk)%64 < 56 {
		h.chunk = append(h.chunk, 0)
	}
	var lenBytes [8]byte
	binary.BigEndian.PutUint64(lenBytes[:], bitLen)
	h.chunk = append(h.chunk, lenBytes[:]...)
}

func (h *sm3) digest() [8]uint32 {
	reg := h.reg
	for i := 0; i+64 <= len(h.chunk); i += 64 {
		reg = sm3CompressBlock(reg, h.chunk[i:i+64])
	}
	return reg
}

// sumHex 计算字符串的 SM3 哈希（小写十六进制）。
func (h *sm3) sumHex(s string) string {
	h.reset()
	h.writeStr(s)
	h.fill()
	reg := h.digest()
	h.reset()
	var out strings.Builder
	for _, r := range reg {
		fmt.Fprintf(&out, "%08x", r)
	}
	return out.String()
}

// sumBytes 计算字符串的 SM3 哈希（32 字节）。
func (h *sm3) sumBytes(s string) []byte {
	h.reset()
	h.writeStr(s)
	h.fill()
	reg := h.digest()
	h.reset()
	return regToBytes(reg)
}

// sumBytesFromBytes 计算原始字节的 SM3 哈希（32 字节）。
func (h *sm3) sumBytesFromBytes(data []byte) []byte {
	h.reset()
	h.writeBytes(data)
	h.fill()
	reg := h.digest()
	h.reset()
	return regToBytes(reg)
}

func regToBytes(reg [8]uint32) []byte {
	var out [32]byte
	for i, r := range reg {
		binary.BigEndian.PutUint32(out[i*4:], r)
	}
	return out[:]
}
