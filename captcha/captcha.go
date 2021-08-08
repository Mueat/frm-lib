package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"time"

	"gitee.com/Rainkropy/frm-lib/util"
	"github.com/afocus/captcha"
)

var cap *captcha.Captcha

// 初始化
func InitCaptcha(fontFile string) {
	cap = captcha.New()
	cap.SetFont(fontFile)
	cap.SetDisturbance(captcha.MEDIUM)
	cap.SetFrontColor(color.RGBA{255, 255, 255, 255})
	cap.SetBkgColor(color.RGBA{0, 0, 0, 255}, color.RGBA{0, 0, 0, 255}, color.RGBA{0, 0, 0, 255})
}

type Captcha struct {
	ID    string `json:"captcha_id"`
	Code  string `json:"code"`
	Image string `json:"image"`
}

// 获取验证码信息
func GetCaptcha(w int, h int) Captcha {
	cap.SetSize(w, h)
	img, str := cap.Create(4, captcha.CLEAR)
	emptyBuff := bytes.NewBuffer(nil) //开辟一个新的空buff
	png.Encode(emptyBuff, img)        //开辟存储空间

	captchaID := fmt.Sprintf("%x-%x-%x", time.Now().UnixNano(), util.GetGID(), rand.Intn(1000000))

	c := Captcha{
		ID:    captchaID,
		Code:  str,
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(emptyBuff.Bytes()),
	}
	return c
}
