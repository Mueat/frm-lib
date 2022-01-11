package weixin

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Mueat/frm-lib/cache"
	"github.com/Mueat/frm-lib/util"
)

type WeApp struct {
	// 微信小程序appid
	AppID string

	// 微信小程序secret
	Secret string

	// redis缓存名称
	CacheName string
}

type AuthResp struct {
	WeixinResponse
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

type AccessTokenResp struct {
	WeixinResponse
	AccessToken string `json:"access_token"`
	ExpriesIn   int64  `json:"expires_in"`
}

type QrCodeResp struct {
	WeixinResponse
	ContentType string `json:"contentType"`
	Buffer      []byte `json:"buffer"`
}

type UserInfo struct {
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
	Gender    int64  `json:"gender"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Language  string `json:"language"`
	OpenID    string `json:"openId"`
	UnionID   string `json:"unionId"`
}

type UserPhone struct {
	PhoneNumber     string                 `json:"phoneNumber"`
	PurePhoneNumber string                 `json:"purePhoneNumber"`
	CountryCode     string                 `json:"countryCode"`
	Watermark       map[string]interface{} `json:"watermark"`
}

const (
	CODE_2_SESSION_URL = "/sns/jscode2session"
	ACCESSTOKEN_URL    = "/cgi-bin/token"
	GET_QR_CODE        = "/wxa/getwxacodeunlimit"

	ACCESSTOKEN_CACHE_KEY = "WEAPP:ACCESSTOKEN:"
)

// 通过code获取用户openid
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
func (mp *WeApp) Code2Session(code string) (*AuthResp, error) {
	params := make(map[string]string)
	params["appid"] = mp.AppID
	params["secret"] = mp.Secret
	params["js_code"] = code
	params["grant_type"] = "authorization_code"

	resp := AuthResp{}
	err := Api("GET", CODE_2_SESSION_URL, params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// 解密微信小程序用户信息
// https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/signature.html
func (mp *WeApp) DecodeData(encryptedData string, iv string, sessionKey string, v interface{}) error {
	if len(sessionKey) != 24 {
		return errors.New("SessionKeyErr")
	}
	if len(iv) != 24 {
		return errors.New("IvErr")
	}
	aesKey, err := util.Base64Decode(sessionKey)
	if err != nil {
		return errors.New("SessionKeyErr")
	}

	iv, err = util.Base64Decode(iv)
	if err != nil {
		return errors.New("IvErr")
	}

	res, err := util.DeAesCode2Base64(encryptedData, []byte(aesKey), []byte(iv))
	if err != nil {
		return errors.New("EncryptedDataErr")
	}

	//uinfo := UserInfo{}
	err = json.Unmarshal(res, &v)
	if err != nil {
		return errors.New("EncryptedDataErr")
	}

	return nil
}

// 获取access token
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
func (mp *WeApp) GetAccessToken() (string, error) {
	var redis *cache.Pools
	key := ACCESSTOKEN_CACHE_KEY + mp.AppID
	if mp.CacheName != "" {
		redis = cache.GetRedis(mp.CacheName)
		// 从缓存中读取
		if redis != nil {
			res := redis.GetString(key)
			if res != "" {
				return res, nil
			}
		}
	}

	// 通过API获取
	params := make(map[string]string)
	params["grant_type"] = "client_credential"
	params["appid"] = mp.AppID
	params["secret"] = mp.Secret

	resp := AccessTokenResp{}
	err := Api("GET", ACCESSTOKEN_URL, params, &resp)
	if err != nil {
		return "", err
	}

	// 设置缓存，30分钟过期
	if mp.CacheName != "" {
		redis.Set(key, resp.AccessToken, time.Minute*30)
	}

	return resp.AccessToken, nil
}

// 生成小程序二维码
// @param string sence 二维码参数
func (mp *WeApp) GetQrCode(page, sence string, width int64) (*QrCodeResp, error) {
	ak, err := mp.GetAccessToken()
	if err != nil {
		return nil, err
	}

	// 通过API获取
	params := make(map[string]interface{})
	params["scene"] = util.URLEncode(sence)
	params["page"] = page
	params["width"] = width

	resp := QrCodeResp{}
	err = Api("POST", GET_QR_CODE+"?access_token="+ak, params, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
