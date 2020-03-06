package qrcode

import (
	"bytes"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"
)

type RGB struct {
	R, G, B uint8
}

type DrawText struct {
	JPG      draw.Image    // 背景图
	Merged   *os.File      // 存放容器
	Buffer   *bytes.Buffer // 缓存存放容器
	Title    string
	RGBA     RGB // 颜色RGB值 默认为黑色
	X0       int
	Y0       int
	Size0    int
	MidX     bool
	MidY     bool
	MoreText []Poster
}
type Poster struct {
	Title string
	RGBA  RGB // 颜色RGB值 默认为黑色
	X     int
	Y     int
	Size0 int
}

func (d *DrawText) DrawPoster(fontName string, b *image.Rectangle) error {
	if d.X0 == 0 {
		d.SetMidX()
	}

	if d.Y0 == 0 {
		d.SetMidY()
	}
	fontSourceBytes, err := ioutil.ReadFile("./" + fontName)
	if err != nil {
		return err
	}

	trueTypeFont, err := freetype.ParseFont(fontSourceBytes)
	if err != nil {
		return err
	}

	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(trueTypeFont)
	fc.SetFontSize(float64(d.Size0))
	fc.SetClip(d.JPG.Bounds())
	fc.SetDst(d.JPG)
	if d.RGBA.R == 0 && d.RGBA.G == 0 && d.RGBA.B == 0 {
		fc.SetSrc(image.Black) // 默认黑
	} else {
		fc.SetSrc(image.NewUniform(color.RGBA{
			R: d.RGBA.R,
			G: d.RGBA.G,
			B: d.RGBA.B,
			A: d.RGBA.R, // 与R保持相同
		}))
	}

	pt := freetype.Pt(d.X0, d.Y0)
	if d.MidX {
		pt = freetype.Pt(b.Max.X/2-d.Size0/2, d.Y0)
	}

	if d.MidY {
		pt = freetype.Pt(d.X0, b.Max.Y/2-d.Size0/2)
	}

	if d.MidX && d.MidY {
		pt = freetype.Pt(b.Max.X/2-d.Size0/2, b.Max.Y/2-d.Size0/2)
	}

	_, err = fc.DrawString(d.Title, pt)
	if err != nil {
		return err
	}

	for _, v := range d.MoreText {
		fc.SetFontSize(float64(v.Size0))
		if v.RGBA.R == 0 && v.RGBA.G == 0 && v.RGBA.B == 0 {
			fc.SetSrc(image.Black) // 默认黑
		} else {
			fc.SetSrc(image.NewUniform(color.RGBA{
				R: v.RGBA.R,
				G: v.RGBA.G,
				B: v.RGBA.B,
				A: v.RGBA.R, // 与R保持相同
			}))
		}

		_, err = fc.DrawString(v.Title, freetype.Pt(v.X, v.Y))
		if err != nil {
			return err
		}
	}

	if d.Buffer != nil {
		err = jpeg.Encode(d.Buffer, d.JPG, nil)
		if err != nil {
			return err
		}
		return nil
	}

	err = jpeg.Encode(d.Merged, d.JPG, nil)
	if err != nil {
		return err
	}
	return nil
}

func (d *DrawText) SetMidX() {
	d.MidY = true
}

func (d *DrawText) SetMidY() {
	d.MidY = true
}
