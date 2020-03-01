package qrcode

import (
	"testing"
)

func TestQrCode_QrCode4Image(t *testing.T) {
	qDebug := NewQrCodeImage("http://h5.ffff.com/share/erfefe.html?id=ife12z4", "share_red.png", "4444")
	qDebug.DebugCode()
	qDebug.SetY(760)
	addr, err := qDebug.QrCode4Image()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
}
