package http

import (
	"mime/multipart"
	"strconv"

	"github.com/Mueat/frm-lib/cache"
	"github.com/Mueat/frm-lib/db"
	"github.com/Mueat/frm-lib/errors"
	"github.com/Mueat/frm-lib/log"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type App struct {
	Request  Request
	Response Response
}

func InitApp(c *gin.Context) App {
	req := Request{
		Ctx: c,
	}
	resp := Response{
		Ctx: c,
	}
	return App{Request: req, Response: resp}
}

func (a *App) GetContext() *gin.Context {
	return a.Request.Ctx
}

// 获取body数据
func (a *App) GetBody() []byte {
	return a.Request.GetBody()
}

// 绑定
func (a *App) Bind(v interface{}) error {
	return a.Request.Bind(v)
}

// 获取body中的string
func (a *App) GetBodyStr(k string) string {
	return a.Request.GetBodyStr(k)
}

// 获取body中的int64
func (a *App) GetBodyInt64(k string) int64 {
	return a.Request.GetBodyInt64(k)
}

// 获取body中的bool值
func (a *App) GetBodyBool(k string) bool {
	return a.Request.GetBodyBool(k)
}

// 绑定body中特定的key值
func (a *App) BodyBind(k string, v interface{}) error {
	return a.Request.BodyBind(k, v)
}

//从url的query中获取指定key的内容，如果key不存在，则返回def内容
func (a *App) GetQuery(key string, def string) string {
	return a.Request.GetQuery(key, def)
}

//从url的query中获取指定key的int64值，如果key不存在，则返回def内容
func (a *App) GetQueryInt64(key string, def int64) int64 {
	res := a.Request.GetQuery(key, "")
	if res != "" {
		ret, err := strconv.Atoi(res)
		if err != nil {
			return def
		}
		return int64(ret)
	}
	return def
}

//从url的params中获取指定key内容
func (a *App) GetParam(key string) string {
	return a.Request.GetParam(key)
}

//从form中获取值
func (a *App) GetForm(key, def string) string {
	return a.Request.GetForm(key, def)
}

//获取上传文件
func (a *App) GetFile(field string) (*multipart.FileHeader, error) {
	return a.Request.GetFile(field)
}

//获取客户端IP
func (a *App) GetIP() string {
	return a.Request.GetIP()
}

// 获取请求头信息
func (a *App) GetHeader(key string) string {
	return a.Request.GetHeader(key)
}

//获取user-agent
func (a *App) GetUserAgent() string {
	return a.Request.GetUserAgent()
}

//设置值
func (a *App) Set(key string, v interface{}) {
	a.Request.Set(key, v)
}

//获取值
func (a *App) Get(key string) (value interface{}, exists bool) {
	return a.Request.Get(key)
}

func (a *App) GetString(key string) string {
	return a.Request.GetString(key)
}

func (a *App) GetStringMap(key string) map[string]interface{} {
	return a.Request.GetStringMap(key)
}

func (a *App) GetStringMapString(key string) map[string]string {
	return a.Request.GetStringMapString(key)
}

func (a *App) GetStringSlice(key string) []string {
	return a.Request.GetStringSlice(key)
}

func (a *App) GetStringMapStringSlice(key string) map[string][]string {
	return a.Request.GetStringMapStringSlice(key)
}

func (a *App) GetBool(key string) bool {
	return a.Request.GetBool(key)
}

func (a *App) GetInt(key string) int {
	return a.Request.GetInt(key)
}

func (a *App) GetInt64(key string) int64 {
	return a.Request.GetInt64(key)
}

func (a *App) GetUint(key string) uint {
	return a.Request.GetUint(key)
}

func (a *App) GetFloat64(key string) float64 {
	return a.Request.GetFloat64(key)
}

// Abort
func (a *App) Abort() {
	a.Response.Abort()
}

func (a *App) AbortWithStatus(code int) {
	a.Response.AbortWithStatus(code)
}

func (a *App) Next() {
	a.Request.Ctx.Next()
}

func (a *App) Status(code int) *App {
	a.Response.StatusCode = code
	return a
}

func (a *App) Send(str string) {
	a.Response.Send(str)
}

func (a *App) Json(v interface{}) {
	a.Response.Json(v)
}

func (a *App) HTML(name string, obj interface{}) {
	a.Response.HTML(200, name, obj)
}

func (a *App) Success(v interface{}) {
	a.Response.Success(v)
}

func (a *App) Error(code int) {
	msg, ok := errors.Errors[code]
	if !ok {
		msg = "Unkonw Error"
	}
	a.Response.Error(code, msg)
}

func (a *App) ErrorMsg(msg string) {
	a.Response.Error(-1, msg)
}

// 数据库
func (a *App) DB(name string) *gorm.DB {
	return db.GetMySql(name)
}

func (a *App) DefaultDB() *gorm.DB {
	return db.GetMySql("")
}

// redis
func (a *App) Redis(name string) *cache.Pools {
	return cache.GetRedis(name)
}

func (a *App) DefaultRedis() *cache.Pools {
	return cache.GetRedis("")
}

// 日志
func (a *App) Log(name string) *zerolog.Logger {
	return log.Get(name)
}

func (a *App) LogDebug() *zerolog.Event {
	return log.Debug()
}

func (a *App) LogInfo() *zerolog.Event {
	return log.Info()
}

func (a *App) LogError() *zerolog.Event {
	return log.Error()
}

func (a *App) LogFatal() *zerolog.Event {
	return log.Fatal()
}

func (a *App) LogPanic() *zerolog.Event {
	return log.Panic()
}
