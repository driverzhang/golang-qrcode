package qrcode

import (
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
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
	Debug        bool      `json:"debug"`         // 调试模式本地生成海报 不走OSS
	Id           string    `json:"id"`            // 海报的唯一标识
	Size         int       `json:"size"`          // 二维码大小 150 == x150 y150
	Type         int       `json:"type"`          // 生成图片类型，默认1.jpg  2.png
	Content      string    `json:"content"`       // 二维码识别出的内容
	BackendImage string    `json:"backend_image"` // 背景图片名称 3.png
	HeadImage    HeadImage `json:"head_image"`    // 嵌入微信头像
	DrawText     DrawText  `json:"draw_text"`     // 嵌入文字+坐标
	MidX         bool      `json:"mid_x"`         // 二维码X坐标是否居中
	MidY         bool      `json:"mid_y"`         // 二维码y坐标是否居中
	X            int       `json:"x"`             // 二维码相对图片坐标
	Y            int       `json:"y"`
}

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
		HeadImage: HeadImage{
			HeadName: "",
			X:        0,
			Y:        0,
			Size:     250,
		},
		DrawText: DrawText{},
	}
	q.MiddleX()
	q.MiddleY()
	return
}

func (q *QrCode) SetDrawText(test DrawText) {
	q.DrawText = test
}

func (q *QrCode) SetHeadImage(head string) {
	q.HeadImage.HeadName = head
	q.HeadImage.MiddleHeadX()
	q.HeadImage.MiddleHeadY()
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
	var qrCode *qrcode.QRCode
	qrCode, err = qrcode.New(q.Content, qrcode.Highest)
	if err != nil {
		err = errors.New("创建二维码失败")
		return
	}

	qrCode.DisableBorder = true
	img = qrCode.Image(q.Size)
	return img, nil
}

func readImgData(url string) (pix []byte, file io.ReadCloser, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	//defer resp.Body.Close()
	//b, err := ioutil.ReadAll(resp.Body)
	//log.Error(string(b))
	file = resp.Body
	return
}

type HeadImage struct {
	HeadName string `json:"head_name"` // 头像名称 or URL
	X        int    `json:"x"`
	Y        int    `json:"y"`
	MidHeadX bool   `json:"mid_head_x"`
	MidHeadY bool   `json:"mid_head_y"`
	Size     int    `json:"size"`
}
type CircleMask struct {
	image    image.Image
	point    image.Point
	diameter int
}

func (h *HeadImage) SetHeadX(x int) {
	h.X = x
	h.MidHeadX = false
}

func (h *HeadImage) SetHeadSize(size int) {
	h.Size = size
}

func (h *HeadImage) SetHeadY(y int) {
	h.Y = y
	h.MidHeadY = false
}

func (h *HeadImage) MiddleHeadX() {
	h.MidHeadX = true
}

func (h *HeadImage) MiddleHeadY() {
	h.MidHeadY = true
}

func (ci CircleMask) ColorModel() color.Model {
	return ci.image.ColorModel()
}

func (ci CircleMask) Bounds() image.Rectangle {
	return image.Rect(0, 0, ci.diameter, ci.diameter)
}

func (ci CircleMask) At(x, y int) color.Color {
	d := ci.diameter
	dis := math.Sqrt(math.Pow(float64(x-d/2), 2) + math.Pow(float64(y-d/2), 2))
	if dis > float64(d)/2 {
		return ci.image.ColorModel().Convert(color.RGBA{255, 255, 255, 255})
	} else {
		return ci.image.At(ci.point.X+x, ci.point.Y+y)
	}
}

func NewCircleMask(img image.Image, p image.Point, d int) CircleMask {
	return CircleMask{img, p, d}
}

func (q *QrCode) DecodeHeadImg(b image.Rectangle) (headImg image.Image, offset image.Point, err error) {
	if q.HeadImage.HeadName == "" {
		return
	}

	offset = image.Point{}
	// 读取背景图URL
	var file io.ReadCloser
	if q.Debug {
		headUrl, err := os.Open(path.Base("./" + q.HeadImage.HeadName)) // 本地打开头像图
		if err != nil {
			return nil, offset, err
		}
		file = headUrl
	} else {
		_, file, err = readImgData(q.HeadImage.HeadName)
		if err != nil {
			return
		}
	}
	defer file.Close()

	nameList := strings.Split(q.HeadImage.HeadName, ".")
	headImageType := nameList[len(nameList)-1]
	switch headImageType {
	case "png":
		headImg, err = png.Decode(file)
		if err != nil {
			return
		}
	case "jpg", "jpeg":
		headImg, err = jpeg.Decode(file)
		if err != nil {
			return
		}
	default:
		err = errors.New("图片格式只支持png/jpg/jpeg")
		return
	}
	offset = image.Pt(q.HeadImage.X, q.HeadImage.Y)
	if q.HeadImage.MidHeadX {
		offset = image.Pt(b.Max.X/2-q.HeadImage.Size/2, q.HeadImage.Y)
	}

	if q.HeadImage.MidHeadY {
		offset = image.Pt(q.HeadImage.X, b.Max.Y/2-q.HeadImage.Size/2)
	}

	if q.HeadImage.MidHeadX && q.HeadImage.MidHeadY {
		offset = image.Pt(b.Max.X/2-q.HeadImage.Size/2, b.Max.Y/2-q.HeadImage.Size/2)
	}

	w := headImg.Bounds().Max.X - headImg.Bounds().Min.X
	h := headImg.Bounds().Max.Y - headImg.Bounds().Min.Y
	d := w
	if w > h {
		d = h
	}

	dstImg := NewCircleMask(headImg, image.Point{d / 4, d / 4}, d/2)
	headImg = dstImg
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

	i, err := os.Open(path.Base("./" + q.BackendImage)) // 本地打开背景图
	if err != nil {
		return
	}

	headUrl, err := os.Open(path.Base("./" + q.HeadImage.HeadName)) // 本地打开头像图
	if err != nil {
		return
	}

	defer i.Close()
	defer headUrl.Close()
	headNameList := strings.Split(q.HeadImage.HeadName, ".")
	headImageType := headNameList[len(nameList)-1]
	var headImg image.Image
	switch headImageType {
	case "png":
		headImg, err = png.Decode(headUrl)
		if err != nil {
			return
		}
	case "jpg", "jpeg":
		headImg, err = jpeg.Decode(headUrl)
		if err != nil {
			return
		}
	default:
		err = errors.New("图片格式只支持png/jpg/jpeg")
		return
	}

	nameAll := ""
	switch imageType {
	case "png":
		nameAll = "png"
		bgImg, err = png.Decode(i)
		if err != nil {
			return
		}
	case "jpg", "jpeg":
		nameAll = "jpg"
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
	draw.Draw(m, headImg.Bounds().Add(offset), headImg, image.Point{X: 0, Y: 0}, draw.Over)

	// 本地生成海报图
	nowName := fmt.Sprintf("%s_backend_%s.%s", nameList[0], q.Id, nameAll)
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

func (q *QrCode) DecodeQrCode(b image.Rectangle) (qrCodeImg image.Image, offset image.Point, err error) {
	qrCodeImg, err = q.createQrCode()
	if err != nil {
		fmt.Println("生成二维码失败:", err)
		return
	}

	// 二维码对象
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
	return
}

func (q *QrCode) QrCode4Image() (addr string, err error) {
	nowName := ""
	var file io.ReadCloser
	// var ossClient oss.OssCdn
	nameList := strings.Split(q.BackendImage, ".")
	imageType := nameList[len(nameList)-1]

	if q.Debug {
		nowName = fmt.Sprintf("%s_backend_%s.%s", nameList[0], q.Id, imageType)
		file, err = os.Open(path.Base("./" + q.BackendImage)) // 本地打开背景图
		if err != nil {
			return
		}
	} else {
		// oss 业务
		// imageHostList := strings.Split(nameList[len(nameList)-2], "/") // http://cdn.ijianqu.com/common/cash_bg.png
		// imageHost := imageHostList[len(imageHostList)-1]
		// nowName = fmt.Sprintf("%s_backend_%s.%s", imageHost, q.Id, imageType)
		// if !IsFilePostfix(nowName) {
		// 	err = errors.New("上传文件格式不符合规范，请重新上传~")
		// 	return
		// }
		//
		// ossClient = oss.GetClient()
		// exit, _ := ossClient.FileIsExit(nowName)
		// addr = nowName
		// if exit {
		// 	return
		// }
		//
		// // 读取背景图URL
		// log.Print(q.BackendImage)
		// _, file, err = readImgData(q.BackendImage)
		// if err != nil {
		// 	return "", err
		// }
		// defer file.Close()
	}

	log.Print(imageType)
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
	qrCodeImg, offset, err = q.DecodeQrCode(b)
	if err != nil {
		return
	}

	headImg, offsetHead, err := q.DecodeHeadImg(b)
	if err != nil {
		return
	}

	m := image.NewRGBA(b)
	draw.Draw(m, b, bgImg, image.Point{X: 0, Y: 0}, draw.Src)                                   // 背景图布局
	draw.Draw(m, qrCodeImg.Bounds().Add(offset), qrCodeImg, image.Point{X: 0, Y: 0}, draw.Over) // 二维码布局
	if q.HeadImage.HeadName != "" {
		draw.Draw(m, headImg.Bounds().Add(offsetHead), headImg, image.Point{X: 0, Y: 0}, draw.Over) // 头像布局
	}
	if q.Debug {
		// 本地生成海报图
		j, err := os.Create(path.Base(nowName))
		if err != nil {
			return "", err
		}
		defer j.Close()

		q.DrawText.JPG = m
		q.DrawText.Merged = j
		err = q.DrawText.DrawPoster("msyhbd.ttc", &b)
		if err != nil {
			return "", err
		}

		if nameList[1] == "png" {
			_ = png.Encode(j, m)
		} else {
			_ = jpeg.Encode(j, m, nil)
		}
		addr = nowName
		return addr, nil
	}

	// 上传至 oss
	// `imgBuff := bytes.NewBuffer(nil)
	// if q.DrawText.Title != "" {
	// 	q.DrawText.JPG = m
	// 	q.DrawText.Buffer = imgBuff
	// 	err = q.DrawText.DrawPoster("msyhbd.ttc", &b)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// }
	//
	// if nameList[1] == "png" {
	// 	_ = png.Encode(imgBuff, m)
	// } else {
	// 	_ = jpeg.Encode(imgBuff, m, nil)
	// }
	//
	// ossName, err := ossClient.Upload(nowName, imgBuff.Bytes())
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }`
	return
}

// 判断是否满足规定文件后缀格式
func IsFilePostfix(file string) bool {
	// 图片
	if strings.Contains(file, ".jpg") ||
		strings.Contains(file, ".JPG") ||
		strings.Contains(file, ".Jpg") ||
		strings.Contains(file, ".png") ||
		strings.Contains(file, ".Png") ||
		strings.Contains(file, ".PNG") ||
		strings.Contains(file, ".jpeg") ||
		strings.Contains(file, ".JPEG") ||
		strings.Contains(file, ".Jpeg") {
		return true
	}

	// 视屏
	if strings.Contains(file, ".avi") ||
		strings.Contains(file, ".Avi") ||
		strings.Contains(file, ".AVI") ||
		strings.Contains(file, ".mp4") ||
		strings.Contains(file, ".Mp4") ||
		strings.Contains(file, ".MP4") ||
		strings.Contains(file, ".rmvb") ||
		strings.Contains(file, ".Rmvb") ||
		strings.Contains(file, ".RMVB") {
		return true
	}
	return false
}
