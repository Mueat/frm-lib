package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	elog "gitee.com/Rainkropy/frm-lib/log"
	"gitee.com/Rainkropy/frm-lib/util"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

const (
	DEVELOPMENT = "DEVELOPMENT"
	PRODUCTION  = "PRODUCTION"
	TESTING     = "TESTING"
	UAT         = "UAT"
	CRASH       = "CRASH"
)

const (
	ErrPack = "HTTP"
)

// 服务配置
type ServerConfig struct {
	// 服务名称
	Name string
	// 环境 DEVELOPMENT PRODUCTION TESTING UAT CRASH
	Environment string
	// 域名地址
	URL string
	// 接口前缀
	ApiURLPrefix string
	// 时区
	TimeZone string
	// 监听地址
	ListenAddr string
}

// 服务
type GinServer struct {
	Engine *gin.Engine
}

var config ServerConfig

// 初始化
func Init(conf ServerConfig) *GinServer {
	config = conf

	var engine *gin.Engine
	if IsDevelopment() {
		engine = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	}
	engine.Use(gin.Recovery())

	// 设置body
	engine.Use(setBody)

	ser := GinServer{
		Engine: engine,
	}
	return &ser
}

// 设置body
func setBody(c *gin.Context) {
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete {
		ct := c.ContentType()
		if util.Stripos(ct, "application/json", 0) > -1 {
			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				return
			}
			_ = c.Request.Body.Close()
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			c.Set("body", body)
			return
		}
	}
	c.Set("body", []byte{})
}

// 设置日志
// @param string confName 日志名称
func (s *GinServer) SetLogger(confName string) {
	if !elog.Has(confName) {
		elog.Error().Str("type", ErrPack).Str("name", "server").Str("method", "SetLogger").Msgf("log config %s not found", confName)
		return
	}

	s.Engine.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		now := time.Now()
		l := elog.Get(confName).Info().Str("type", "AccessLog")
		bodyInter, _ := c.Get("body")
		bodyBytes := bodyInter.([]byte)
		mp := map[string]interface{}{
			"$client_ip":            c.ClientIP(),
			"$timestamp":            now.Format(time.RFC3339Nano),
			"$timestamp_unix":       strconv.Itoa(int(now.UnixNano() / 1e6)),
			"$server_addr":          util.GetServerIP(),
			"$method":               c.Request.Method,
			"$body_size":            c.Writer.Size(),
			"$request_time":         now.Sub(start),
			"$http_host":            c.Request.Host,
			"$http_user_agent":      c.Request.UserAgent(),
			"$status":               c.Writer.Status(),
			"$http_referer":         c.Request.Referer(),
			"$request_uri":          c.Request.URL.String(),
			"$args":                 c.Request.URL.RawQuery,
			"$http_x_forwarded_for": c.Request.Header.Get("X-Forwarded-For"),
			"$error":                c.Errors.ByType(gin.ErrorTypePrivate).String(),
			"$body":                 string(bodyBytes),
		}
		conf := elog.GetConfig(confName)
		logSeted := false
		if conf.AccessLogFormat != "" {
			var fmap map[string]string
			err := json.Unmarshal([]byte(conf.AccessLogFormat), &fmap)
			if err != nil {
				logSeted = true
				for k, v := range fmap {
					if val, ok := mp[v]; ok {
						l = l.Interface(k, val)
					}
				}
			}
		}

		if !logSeted {
			for k, v := range mp {
				l = l.Interface(util.Substr(k, 1, -1), v)
			}
		}

		l.Send()
	})
}

// 启动
func (s *GinServer) Start() error {
	return gracehttp.Serve(
		&http.Server{Addr: config.ListenAddr, Handler: s.Engine},
	)
}

// 绑定路由
func (s *GinServer) Handle(method string, url string, handlers ...RouterFun) {
	if config.ApiURLPrefix != "" {
		url = config.ApiURLPrefix + url
		if util.Substr(url, 0, 1) != "/" {
			url = "/" + url
		}
	}
	ginHandlers := make([]gin.HandlerFunc, 0)
	for _, fun := range handlers {
		ginHandlers = append(ginHandlers, func(c *gin.Context) {
			app := InitApp(c)
			fun(&app)
		})
	}
	s.Engine.Handle(method, url, ginHandlers...)
}

// 静态文件绑定
func (s *GinServer) Static(url string, dir string) {
	s.Engine.Static(url, dir)
}

// 绑定POST请求
func (s *GinServer) Post(url string, handlers ...RouterFun) {
	s.Handle(http.MethodPost, url, handlers...)
}

// 绑定GET请求
func (s *GinServer) Get(url string, handlers ...RouterFun) {
	s.Handle(http.MethodGet, url, handlers...)
}

// 设置中间件
func (s *GinServer) Use(funs ...RouterFun) {
	for _, fun := range funs {
		s.Engine.Use(func(c *gin.Context) {
			app := InitApp(c)
			fun(&app)
		})
	}
}

// 设置组
func (s *GinServer) Group(groupURL string, routers []Router, middlewares ...RouterFun) {
	gp := s.Engine.Group(groupURL)
	midds := make([]gin.HandlerFunc, 0)
	for _, fun := range middlewares {
		midds = append(midds, func(c *gin.Context) {
			app := InitApp(c)
			fun(&app)
		})
	}
	gp.Use(midds...)
	{
		for _, r := range routers {
			handler := r.Handler
			gp.Handle(r.Method, r.URL, func(c *gin.Context) {
				app := InitApp(c)
				handler(&app)
			})
		}
	}

}

// 是否是开发环境
func IsDevelopment() bool {
	return config.Environment == DEVELOPMENT
}

// 是否是测试环境
func IsTesting() bool {
	return config.Environment == TESTING
}

// 是否是正式环境
func IsProdcution() bool {
	return config.Environment == PRODUCTION
}

// 是否是预发环境
func IsUat() bool {
	return config.Environment == UAT
}

// 是否是crash环境
func IsCrash() bool {
	return config.Environment == CRASH
}
