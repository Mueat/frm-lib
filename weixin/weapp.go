package weixin

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitee.com/Rainkropy/frm-lib/cache"
	"gitee.com/Rainkropy/frm-lib/util"
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

const (
	CODE_2_SESSION_URL = "/sns/jscode2session"
	ACCESSTOKEN_URL    = "/cgi-bin/token"

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
func (mp *WeApp) DecodeData(encryptedData string, iv string, sessionKey string) (*UserInfo, error) {
	if len(sessionKey) != 24 {
		return nil, errors.New("SessionKeyErr")
	}
	if len(iv) != 24 {
		return nil, errors.New("IvErr")
	}
	aesKey, err := util.Base64Decode(sessionKey)
	if err != nil {
		return nil, errors.New("SessionKeyErr")
	}

	iv, err = util.Base64Decode(iv)
	if err != nil {
		return nil, errors.New("IvErr")
	}

	res, err := util.DeAesCode2Base64(encryptedData, []byte(aesKey), []byte(iv))

	if err != nil {
		return nil, errors.New("EncryptedDataErr")
	}

	fmt.Println(string(res))

	uinfo := UserInfo{}
	err = json.Unmarshal(res, &uinfo)
	if err != nil {
		return nil, errors.New("EncryptedDataErr")
	}

	return &uinfo, nil
}

// 获取access token
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
func (mp *WeApp) GetAccessToken() (string, error) {
	cache := cache.GetRedis(mp.CacheName)
	key := ACCESSTOKEN_CACHE_KEY + mp.AppID

	// 从缓存中读取
	if cache != nil {
		res := cache.GetString(key)
		if res != "" {
			return res, nil
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
	cache.Set(key, resp.AccessToken, time.Minute*30)

	return resp.AccessToken, nil
}
