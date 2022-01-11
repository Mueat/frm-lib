package http

import "github.com/gin-gonic/gin"

type Response struct {
	Ctx        *gin.Context
	StatusCode int
}

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (r *Response) Status(code int) *Response {
	r.StatusCode = code
	return r
}

func (r *Response) Send(str string) {
	r.Ctx.String(r.StatusCode, str)
}

func (r *Response) Json(v interface{}) {
	r.Ctx.PureJSON(r.StatusCode, v)
}

func (r *Response) HTML(code int, name string, obj interface{}) {
	r.Ctx.HTML(code, name, obj)
}

func (r *Response) Success(v interface{}) {
	apiResp := ApiResponse{
		Code: 0,
		Msg:  "success",
		Data: v,
	}
	r.Json(apiResp)
}

func (r *Response) Error(code int, msg string) {
	apiResp := ApiResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
	r.Json(apiResp)
}

// Abort
func (r *Response) Abort() {
	r.Ctx.Abort()
}

func (r *Response) AbortWithStatus(code int) {
	r.Ctx.AbortWithStatus(code)
}
