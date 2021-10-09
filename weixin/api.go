package weixin

import (
	"gitee.com/Rainkropy/frm-lib/curl"
	"gitee.com/Rainkropy/frm-lib/errors"
	"gitee.com/Rainkropy/frm-lib/log"
)

type WeixinResponseInterface interface {
	GetCode() int64
	GetMsg() string
}

type WeixinResponse struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (r *WeixinResponse) GetCode() int64 {
	return r.ErrCode
}

func (r *WeixinResponse) GetMsg() string {
	return r.ErrMsg
}

// 微信API地址
const WEIXIN_BASE_URL = "https://api.weixin.qq.com"

// 发送请求至微信
func Api(method string, action string, params interface{}, resp WeixinResponseInterface) error {
	client := curl.New(curl.Opts{})

	url := WEIXIN_BASE_URL + action
	headers := make(map[string]string)
	if method == "POST" {
		headers["content-type"] = "application/json"
	}
	res, err := client.Do(method, url, params, nil)
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "api.curl.do").Err(err).Send()
		return err
	}

	err = curl.BindResponse(res, resp)
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "api.curl.BindResponse").Err(err).Send()
		return err
	}

	if resp.GetCode() != 0 {
		log.Error().Str("lib", "weixin").Str("method", "api").Str("url", url).Interface("params", params).Int64("code", resp.GetCode()).Str("msg", resp.GetMsg()).Send()
		return errors.Msg(resp.GetMsg())
	}
	return nil
}
