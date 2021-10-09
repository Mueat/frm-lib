package curl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gitee.com/Rainkropy/frm-lib/util"
	"github.com/ddliu/go-httpclient"
)

const (
	DefaultConnectTimeout = 3
	DefaultTimeout        = 60
)

type Opts struct {
	// 链接超时时间，单位：秒，默认：3秒
	ConnectTimeout int64
	// 请求超时时间，单位：秒，默认：60秒
	Timeout int64
	// User-Agent
	UserAgent string
	// 请求发起前执行的方法
	BeforeRequestFun func(*http.Client, *http.Request)
	// 重试次数
	Retry uint
	// 是否开启Debug
	Debug bool
}

type Client struct {
	Options    Opts
	Tried      uint
	HttpClient *httpclient.HttpClient
}

// New curl client
func New(options Opts) *Client {
	hmap := httpclient.Map{}
	if options.ConnectTimeout > 0 {
		hmap[httpclient.OPT_CONNECTTIMEOUT] = options.ConnectTimeout
	} else {
		hmap[httpclient.OPT_CONNECTTIMEOUT] = DefaultConnectTimeout
	}

	if options.Timeout > 0 {
		hmap[httpclient.OPT_TIMEOUT] = options.Timeout
	} else {
		hmap[httpclient.OPT_TIMEOUT] = DefaultTimeout
	}

	if options.UserAgent != "" {
		hmap[httpclient.OPT_USERAGENT] = options.UserAgent
	}

	if options.BeforeRequestFun != nil {
		hmap[httpclient.OPT_BEFORE_REQUEST_FUNC] = options.BeforeRequestFun
	}

	if options.Debug {
		hmap[httpclient.OPT_DEBUG] = options.Debug
	}

	client := Client{}
	client.HttpClient = httpclient.NewHttpClient().Defaults(hmap)
	client.Options = options
	client.Tried = 0
	return &client
}

// 发起http请求
func (c *Client) Do(method string, url string, data interface{}, headers map[string]string) (*httpclient.Response, error) {
	c.Tried++
	isJson := false
	for k, v := range headers {
		if util.Strtoupper(k) == "CONTENT-TYPE" && util.Stripos(v, "application/json", 0) > -1 {
			isJson = true
		}
	}
	hc := c.HttpClient.WithHeaders(headers)

	var resp *httpclient.Response
	var err error
	if util.Strtoupper(method) == http.MethodPost {
		if isJson {
			resp, err = hc.PostJson(url, data)
		} else {
			resp, err = hc.Post(url, data)
		}
	}
	if util.Strtoupper(method) == http.MethodPut {
		if isJson {
			resp, err = hc.PutJson(url, data)
		} else {
			var body []byte
			switch t := data.(type) {
			case []byte:
				body = t
			case string:
				body = []byte(t)
			default:
				var err error
				body, err = json.Marshal(data)
				if err != nil {
					return nil, err
				}
			}

			resp, err = hc.Do(http.MethodPut, url, headers, bytes.NewReader(body))
		}
	} else if util.Strtoupper(method) == http.MethodGet {
		resp, err = hc.Get(url, data)
	} else if util.Strtoupper(method) == http.MethodDelete {
		resp, err = hc.Delete(url, data)
	}

	if err != nil {
		if c.Options.Retry > c.Tried {
			return c.Do(method, url, data, headers)
		}
	} else if resp.StatusCode >= 500 {
		if c.Options.Retry > c.Tried {
			return c.Do(method, url, data, headers)
		}
	}
	return resp, err
}

func BindResponse(resp *httpclient.Response, bindData interface{}) error {
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, bindData)
}
