package http

import (
	"encoding/json"
	"errors"
	"strings"

	"gitee.com/Rainkropy/frm-lib/log"
	"github.com/gin-gonic/gin"
)

type Request struct {
	Ctx  *gin.Context
	Body map[string]interface{}
}

// 获取body数据
func (r *Request) GetBody() []byte {
	bodyInter, _ := r.Ctx.Get("body")
	bodyBytes := bodyInter.([]byte)
	return bodyBytes
}

// 绑定
func (r *Request) Bind(v interface{}) error {
	err := json.Unmarshal(r.GetBody(), v)
	if err != nil {
		log.Error().Err(err).Str("type", ErrPack).Str("name", "request").Str("method", "Bind").Str("body", string(r.GetBody())).Send()
	}
	return err
}

func (r *Request) bindMap() error {
	var bm map[string]interface{}
	err := r.Bind(bm)
	if err == nil {
		r.Body = bm
	}
	return err
}

// 获取body中的string
func (r *Request) GetBodyStr(k string) string {
	if r.Body == nil {
		if err := r.bindMap(); err != nil {
			return ""
		}
	}
	if res, ok := r.Body[k].(string); ok {
		return res
	}
	return ""
}

// 获取body中的int64
func (r *Request) GetBodyInt64(k string) int64 {
	if r.Body == nil {
		if err := r.bindMap(); err != nil {
			return 0
		}
	}
	if res, ok := r.Body[k].(int64); ok {
		return res
	}
	return 0
}

// 获取body中的bool值
func (r *Request) GetBodyBool(k string) bool {
	if r.Body == nil {
		if err := r.bindMap(); err != nil {
			return false
		}
	}
	if res, ok := r.Body[k].(bool); ok {
		return res
	}
	return false
}

// 绑定body中特定的key值
func (r *Request) BodyBind(k string, v interface{}) error {
	if r.Body == nil {
		if err := r.bindMap(); err != nil {
			return err
		}
	}
	res, ok := r.Body[k]
	if !ok {
		return errors.New("bodayKeyNotFound")
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, v)
	return err
}

//从url的query中获取指定key的内容，如果key不存在，则返回def内容
func (r *Request) GetQuery(key string, def string) string {
	return r.Ctx.DefaultQuery(key, def)
}

//从url的params中获取指定key内容
func (r *Request) GetParam(key string) string {
	return r.Ctx.Param(key)
}

//从form中获取值
func (r *Request) GetForm(key, def string) string {
	return r.Ctx.DefaultPostForm(key, def)
}

//获取客户端IP
func (r *Request) GetIP() string {
	if r.Ctx.Request.Header.Get("X-Real-IP") != "" {
		return strings.Split(r.Ctx.Request.Header.Get("X-Real-IP"), ",")[0]
	}
	if r.Ctx.Request.Header.Get("X-Forwarded-For") != "" {
		return strings.Split(r.Ctx.Request.Header.Get("X-Forwarded-For"), ",")[0]
	}
	return r.Ctx.ClientIP()
}

//获取user-agent
func (r *Request) GetUserAgent() string {
	return r.Ctx.Request.UserAgent()
}

//设置值
func (r *Request) Set(key string, v interface{}) {
	r.Ctx.Set(key, v)
}

//获取值
func (r *Request) Get(key string) (value interface{}, exists bool) {
	return r.Ctx.Get(key)
}

func (r *Request) GetString(key string) string {
	return r.Ctx.GetString(key)
}

func (r *Request) GetStringMap(key string) map[string]interface{} {
	return r.Ctx.GetStringMap(key)
}

func (r *Request) GetStringMapString(key string) map[string]string {
	return r.Ctx.GetStringMapString(key)
}

func (r *Request) GetStringSlice(key string) []string {
	return r.Ctx.GetStringSlice(key)
}

func (r *Request) GetStringMapStringSlice(key string) map[string][]string {
	return r.Ctx.GetStringMapStringSlice(key)
}

func (r *Request) GetBool(key string) bool {
	return r.Ctx.GetBool(key)
}

func (r *Request) GetInt(key string) int {
	return r.Ctx.GetInt(key)
}

func (r *Request) GetInt64(key string) int64 {
	return r.Ctx.GetInt64(key)
}

func (r *Request) GetUint(key string) uint {
	return r.Ctx.GetUint(key)
}

func (r *Request) GetFloat64(key string) float64 {
	return r.Ctx.GetFloat64(key)
}

// Abort
func (r *Request) Abort() {
	r.Ctx.Abort()
}
