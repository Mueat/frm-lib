package util

import (
	"errors"
	"time"
)

// JWT
type JWT struct {
	Aud   string      `json:"aud"`   //接收对象
	Exp   int64       `json:"exp"`   //到期时间
	Nbf   int64       `json:"nbf"`   //生效时间
	Iat   int64       `json:"iat"`   //创建时间
	Jti   string      `json:"jti"`   //token的唯一标识
	IP    string      `json:"ip"`    //生成的IP地址
	UID   uint        `json:"uid"`   // 用户ID
	Extra interface{} `json:"extra"` // 其他数据
}

// 生成JWT请求对象
type JWTReq struct {
	Secret string      // 加密秘钥
	Expire int64       // 多少秒后过期
	UID    uint        // 用户ID
	Aud    string      // 接收对象
	IP     string      // IP地址
	Nbf    *time.Time  // 生效时间
	Extra  interface{} // 额外数据
}

// 常用AUD
const (
	JWT_AUD_USER    = "USER"    // 用户 aud
	JWT_AUD_ADMIN   = "ADMIN"   // 管理员 aud
	JWT_AUD_AGENT   = "AGENT"   // 代理商 aud
	JWT_AUD_STORE   = "STORE"   // 店铺 aud
	JWT_AUD_TEACHER = "TEACHER" // 老师 aud
	JWT_AUD_STUDENT = "STUDENT" // 学生 aud
)

// 生成token
func GetJWT(req JWTReq) string {
	if req.Nbf == nil {
		timeNow := time.Now()
		req.Nbf = &timeNow
	}
	now := time.Now().Unix()
	jti := Uniqid("tk")
	token := JWT{
		Aud:   req.Aud,
		Exp:   now + req.Expire,
		Nbf:   req.Nbf.Unix(),
		Iat:   now,
		Jti:   jti,
		IP:    req.Aud,
		UID:   req.UID,
		Extra: req.Extra,
	}

	tokenStr := EncryptJWT(token, req.Secret)
	return tokenStr
}

// 校验token
// @param string encryptSecret 加密秘钥
// @param string token 要校验的token
// @param string aud token的接收对象
// @param string ip token使用的IP地址
func VerifyJWT(encryptSecret string, token string, aud string, ip string) (*JWT, error) {
	if token == "" {
		return nil, errors.New("token not set")
	}
	tk := JWT{}
	er := DecryptJWT(token, encryptSecret, &tk)
	if er != nil {
		return nil, er
	}
	if aud != "" && tk.Aud != aud {
		return nil, errors.New("token aud error")
	}
	if tk.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	if tk.Nbf > time.Now().Unix() {
		return nil, errors.New("token not effective")
	}
	if tk.UID < 1 {
		return nil, errors.New("token uid error")
	}
	if tk.IP != ip {
		return nil, errors.New("token ip error")
	}

	return &tk, nil
}
