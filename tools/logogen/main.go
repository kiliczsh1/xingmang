// 图标生成器：把 docs/brand/logo-original.jpg 烘成 web/public/brand 下的全套尺寸。
//
// 用法（仓库根目录）：
//
//	go run ./tools/logogen docs/brand/logo-original.jpg web/public/brand
//	cp web/public/brand/favicon.ico web/public/favicon.ico
//
// 说明：原图里星体的右尖与最粗的一条拖尾是连在一起的，自动分割一定会把那条拖尾
// 带进来（见 docs/brand/logo-usage.md §1）。所以在没有矢量源之前，这里做的是
// "以星体为中心裁切 + 面积平均降采样"，而不是抠出纯星形。
// 16px 单独用更紧的裁切，否则整幅作品会糊成一团。
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// 用原图中心裁切 + 面积平均降采样生成图标。
//
// 为什么不抠出"纯星形"：原图里星体的右尖与最粗的一条拖尾是连在一起的
// （见 docs/brand 的调试图），任何自动分割都会把那条拖尾带进来。
// 在没有矢量源之前，忠实缩放整幅作品比"抠一半"更稳。
func main() {
	srcPath, outDir := os.Args[1], os.Args[2]
	f, err := os.Open(srcPath)
	if err != nil {
		panic(err)
	}
	img, err := jpeg.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	b := img.Bounds()
	W, H := b.Dx(), b.Dy()

	// 等比例缩小整幅作品：不裁切、不调色，保持原图的构图与亮度。
	// （曾经按星体中心裁切并拉过对比度，结果顶栏里只剩一颗过亮的星，拖尾全被切掉。）
	cx, cy := float64(W)/2, float64(H)/2
	sideFor := func(int) float64 { return float64(W) }

	sample := func(size int, side float64) []float64 {
		x0 := cx - side/2
		y0 := cy - side/2
		out := make([]float64, size*size*4)
		scale := side / float64(size)
		for oy := 0; oy < size; oy++ {
			for ox := 0; ox < size; ox++ {
				var sr, sg, sb float64
				var cnt int
				for sy := 0; sy < 4; sy++ {
					for sx := 0; sx < 4; sx++ {
						fx := x0 + (float64(ox)+(float64(sx)+0.5)/4)*scale
						fy := y0 + (float64(oy)+(float64(sy)+0.5)/4)*scale
						ix, iy := int(fx), int(fy)
						if ix < 0 || iy < 0 || ix >= W || iy >= H {
							cnt++
							continue
						}
						c := color.NRGBAModel.Convert(img.At(ix, iy)).(color.NRGBA)
						sr += float64(c.R)
						sg += float64(c.G)
						sb += float64(c.B)
						cnt++
					}
				}
				if cnt == 0 {
					cnt = 1
				}
				i := (oy*size + ox) * 4
				out[i] = sr / float64(cnt) / 255
				out[i+1] = sg / float64(cnt) / 255
				out[i+2] = sb / float64(cnt) / 255
				out[i+3] = 1
			}
		}
		return out
	}

	makeIcon := func(size int) *image.NRGBA {
		imgOut := image.NewNRGBA(image.Rect(0, 0, size, size))
		radius := 0.22 * float64(size)
		px := sample(size, sideFor(size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				cov := 0.0
				for sy := 0; sy < 4; sy++ {
					for sx := 0; sx < 4; sx++ {
						fx := float64(x) + (float64(sx)+0.5)/4
						fy := float64(y) + (float64(sy)+0.5)/4
						if insideRoundRect(fx, fy, float64(size), radius) {
							cov += 1.0 / 16
						}
					}
				}
				i := (y*size + x) * 4
				imgOut.SetNRGBA(x, y, color.NRGBA{
					uint8(math.Round(px[i] * 255)),
					uint8(math.Round(px[i+1] * 255)),
					uint8(math.Round(px[i+2] * 255)),
					uint8(math.Round(cov * 255)),
				})
			}
		}
		return imgOut
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	pngBySize := map[int][]byte{}
	var sizes []int
	for _, s := range []int{16, 32, 48, 64, 128, 180, 256, 512} {
		im := makeIcon(s)
		var buf bytes.Buffer
		if err := png.Encode(&buf, im); err != nil {
			panic(err)
		}
		name := filepath.Join(outDir, fmt.Sprintf("mark-%d.png", s))
		if err := os.WriteFile(name, buf.Bytes(), 0o644); err != nil {
			panic(err)
		}
		pngBySize[s] = buf.Bytes()
		sizes = append(sizes, s)
		fmt.Println("wrote", name)
	}
	_ = sizes

	// favicon：浏览器标签页，48 已经足够，多一帧只是体积。
	icoSizes := []int{16, 32, 48}
	writeICO(filepath.Join(outDir, "favicon.ico"), icoSizes, pngBySize)

	// 应用图标：Windows 的 exe / 任务栏要 256，否则大图标视图会糊。
	appIcoSizes := []int{16, 32, 48, 64, 128, 256}
	writeICO(filepath.Join(outDir, "app.ico"), appIcoSizes, pngBySize)
}

func writeICO(path string, sizes []int, pngBySize map[int][]byte) {
	frames := make([][]byte, len(sizes))
	for i, s := range sizes {
		frames[i] = pngBySize[s]
	}
	data := buildICO(sizes, frames)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("wrote", path, len(data), "bytes")
}

func insideRoundRect(x, y, size, r float64) bool {
	if x < 0 || y < 0 || x > size || y > size {
		return false
	}
	cx := math.Min(math.Max(x, r), size-r)
	cy := math.Min(math.Max(y, r), size-r)
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= r*r
}

func buildICO(sizes []int, pngs [][]byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(len(sizes)))
	offset := 6 + 16*len(sizes)
	for i, s := range sizes {
		buf.WriteByte(byte(s))
		buf.WriteByte(byte(s))
		buf.WriteByte(0)
		buf.WriteByte(0)
		binary.Write(&buf, binary.LittleEndian, uint16(1))
		binary.Write(&buf, binary.LittleEndian, uint16(32))
		binary.Write(&buf, binary.LittleEndian, uint32(len(pngs[i])))
		binary.Write(&buf, binary.LittleEndian, uint32(offset))
		offset += len(pngs[i])
	}
	for _, p := range pngs {
		buf.Write(p)
	}
	return buf.Bytes()
}
