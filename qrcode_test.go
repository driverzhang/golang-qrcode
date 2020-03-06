package qrcode

import (
	"testing"
)

func TestQrCode_QrCode4Image(t *testing.T) {
	qDebug := NewQrCodeImage("http://h5.ffff.com/share/erfefe.html?id=ife12z4", "sharelist.png", "4444")
	qDebug.DebugCode()
	qDebug.SetY(1133)
	poster := Poster{
		Title: "来呀！",
		X:     285,
		Y:     700,
		Size0: 50,
	}
	drawText1 := DrawText{
		Title:    "200",
		Size0:    100,
		Y0:       490,
		X0:       285,
		RGBA:     struct{ R, G, B uint8 }{R: 255, G: 53, B: 53},
		MoreText: []Poster{poster},
	}

	qDebug.SetDrawText(drawText1)
	// qDebug.SetHeadImage("zxx.jpeg")
	// qDebug.HeadImage.SetHeadSize(550)
	addr, err := qDebug.QrCode4Image()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
}
