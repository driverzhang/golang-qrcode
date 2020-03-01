package qrcode

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

var (
	bgImg     image.Image
	qrCodeImg image.Image
	offset    image.Point
)

type QrCode struct {
	Debug        bool   `json:"debug"`         // 调试模式本地生成海报 不走OSS
	Id           string `json:"id"`            // 海报的唯一标识
	Size         int    `json:"size"`          // 二维码大小 150 == x150 y150
	Type         int    `json:"type"`          // 生成图片类型，默认1.jpg  2.png
	Content      string `json:"content"`       // 二维码识别出的内容
	BackendImage string `json:"backend_image"` // 背景图片名称 3.png
	MidX         bool   `json:"mid_x"`         // 二维码X坐标是否居中
	MidY         bool   `json:"mid_y"`         // 二维码y坐标是否居中
	X            int    `json:"x"`             // 二维码相对图片坐标
	Y            int    `json:"y"`
}

/*
 * content： 二维码扫码读取的内容
 * BackendImage: debug模式只需要传入背景图片名称，否知传入URL
 * id：业务UID
 * 默认二维码位于背景图位置 X, Y 都是居中位置
 */
func NewQrCodeImage(content, BackendImage, id string) (q *QrCode) {
	if id == "" {
		return
	}
	q = &QrCode{
		Content:      content,
		Size:         150, // 默认150. X:150 y:150
		BackendImage: BackendImage,
		X:            0,
		Y:            0,
		Id:           id,
	}
	q.MiddleX()
	q.MiddleY()
	return
}

func (q *QrCode) SetQrCodeSize(size int) {
	q.Size = size
}

func (q *QrCode) SetX(x int) {
	q.X = x
	q.MidX = false
}

func (q *QrCode) SetY(y int) {
	q.Y = y
	q.MidY = false
}

// 设置debug模式，图片都应该存在本地该路径下 便于直接观察调试图片效果
func (q *QrCode) DebugCode() {
	q.Debug = true
}

func (q *QrCode) MiddleX() {
	q.MidX = true
}

func (q *QrCode) MiddleY() {
	q.MidY = true
}

// content: 二维码识别信息内容输入
func (q *QrCode) createQrCode() (img image.Image, err error) {
	qrCode, err := qrcode.New(q.Content, qrcode.Highest)
	if err != nil {
		err = errors.New("创建二维码失败")
		return
	}

	qrCode.DisableBorder = true
	img = qrCode.Image(q.Size)
	return
}

func readImgData(url string) (pix []byte, file io.ReadCloser, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	// defer resp.Body.Close()
	file = resp.Body
	return
}

func (q *QrCode) QrCode4ImageDebug() (err error) {
	nameList := strings.Split(q.BackendImage, ".")
	imageType := nameList[len(nameList)-1]
	qrCodeImg, err = q.createQrCode()
	if err != nil {
		fmt.Println("生成二维码失败:", err)
		return
	}

	i, err := os.Open(path.Base("./" + q.BackendImage))
	if err != nil {
		return
	}
	defer i.Close()
	switch imageType {
	case "png":
		bgImg, err = png.Decode(i)
		if err != nil {
			return
		}
	case "jpg", "jpeg":
		bgImg, err = jpeg.Decode(i)
		if err != nil {
			return
		}
	default:
		err = errors.New("图片格式只支持png/jpg/jpeg")
		return
	}

	b := bgImg.Bounds()
	offset = image.Pt(q.X, q.Y)
	if q.MidX {
		offset = image.Pt(b.Max.X/2-q.Size/2, q.Y)
	}

	if q.MidY {
		offset = image.Pt(q.X, b.Max.Y/2-q.Size/2)
	}

	if q.MidX && q.MidY {
		offset = image.Pt(b.Max.X/2-q.Size/2, b.Max.Y/2-q.Size/2)
	}
	m := image.NewRGBA(b)
	draw.Draw(m, b, bgImg, image.Point{X: 0, Y: 0}, draw.Src)
	draw.Draw(m, qrCodeImg.Bounds().Add(offset), qrCodeImg, image.Point{X: 0, Y: 0}, draw.Over)

	// 本地生成海报图
	nowName := fmt.Sprintf("%s_backend_%s.%s", nameList[0], q.Id, imageType)
	j, err := os.Create(path.Base(nowName))
	if err != nil {
		return
	}
	defer j.Close()
	if nameList[1] == "png" {
		_ = png.Encode(j, m)
	} else {
		_ = jpeg.Encode(j, m, nil)
	}
	return
}

func (q *QrCode) QrCode4Image() (addr string, err error) {
	if q.Debug {
		err = q.QrCode4ImageDebug()
		if err != nil {
			return
		}
		return
	}

	nameList := strings.Split(q.BackendImage, ".")
	imageType := nameList[len(nameList)-1]
	imageHostList := strings.Split(nameList[len(nameList)-2], "/")
	imageHost := imageHostList[len(imageHostList)-1]
	nowName := fmt.Sprintf("%s_backend_%s.%s", imageHost, q.Id, imageType)
	// if !oss.IsFilePostfix(nowName) { // OSS格式限制判断
	// 	err = errors.New("上传文件格式不符合规范，请重新上传~")
	// 	return
	// }

	_, file, err := readImgData(q.BackendImage) // 读取背景图 URL
	if err != nil {
		return
	}
	defer file.Close()

	qrCodeImg, err = q.createQrCode()
	if err != nil {
		fmt.Println("生成二维码失败:", err)
		return
	}
	switch imageType {
	case "png":
		bgImg, err = png.Decode(file)
		if err != nil {
			return
		}
	case "jpg", "jpeg":
		bgImg, err = jpeg.Decode(file)
		if err != nil {
			return
		}
	default:
		err = errors.New("图片格式只支持png/jpg/jpeg")
		return
	}

	b := bgImg.Bounds()
	offset = image.Pt(q.X, q.Y)
	if q.MidX {
		offset = image.Pt(b.Max.X/2-q.Size/2, q.Y)
	}

	if q.MidY {
		offset = image.Pt(q.X, b.Max.Y/2-q.Size/2)
	}

	if q.MidX && q.MidY {
		offset = image.Pt(b.Max.X/2-q.Size/2, b.Max.Y/2-q.Size/2)
	}
	m := image.NewRGBA(b)
	draw.Draw(m, b, bgImg, image.Point{X: 0, Y: 0}, draw.Src)
	draw.Draw(m, qrCodeImg.Bounds().Add(offset), qrCodeImg, image.Point{X: 0, Y: 0}, draw.Over)

	imgBuff := bytes.NewBuffer(nil)
	if nameList[1] == "png" {
		_ = png.Encode(imgBuff, m)
	} else {
		_ = jpeg.Encode(imgBuff, m, nil)
	}

	// 上传至 你的 oss
	// ossClient := oss.GetClient()             // 初始化你的oss
	// exit, _ := ossClient.FileIsExit(nowName) // 判断OSS上该文件是否已存在
	// addr = nowName                           // 返回文件名
	// if exit {
	// 	return
	// }
	// ossName, err := ossClient.Upload(nowName, imgBuff.Bytes()) // 上传到你的OSS
	fmt.Print(nowName)
	return
}
