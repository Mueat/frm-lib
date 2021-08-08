package log

import (
	"os"
	"path"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rs/zerolog"
)

var configs map[string]LogConfig
var loggers map[string]zerolog.Logger

type LogConfig struct {
	LogPath         string //保存的日志目录
	LogName         string //保存的日志文件名称
	LogLevel        int8   //日志等级
	AccessLogFormat string //请求日志格式
	Default         bool   //是否是默认的日志
}

// 初始化日志
// @param map[string]LogConfig confs 日志配置
func Init(confs map[string]LogConfig) {
	loggers = make(map[string]zerolog.Logger)
	configs = confs
	for name, conf := range confs {
		logFile := path.Join(conf.LogPath, conf.LogName)
		rl, err := rotatelogs.New(logFile)
		if err != nil {
			panic(err)
		}
		loggers[name] = zerolog.New(rl).Level(zerolog.Level(conf.LogLevel)).With().Timestamp().Caller().Logger()
	}
}

func Has(name string) bool {
	_, ok := loggers[name]
	return ok
}

// 获取全部配置
func GetConfigs() map[string]LogConfig {
	return configs
}

// 根据名称获取配置
func GetConfig(name string) *LogConfig {
	conf, ok := configs[name]
	if !ok {
		return nil
	}
	return &conf
}

// 获取日志处理器
// @param string name 日志名称
func Get(name string) *zerolog.Logger {
	if name != "" {
		if l, ok := loggers[name]; ok {
			return &l
		}
	}
	for k, conf := range configs {
		if conf.Default {
			if l, ok := loggers[k]; ok {
				return &l
			}
		}
	}
	logger := zerolog.New(os.Stderr)
	return &logger
}

// 设置日志钩子
// @param string name 日志名称
// @param zerolog.Hook hook 钩子
// @param bool all 是否全部日志处理器都启用该钩子
func Hook(name string, hook zerolog.Hook, all bool) {
	for k := range loggers {
		if all || k == name {
			loggers[k].Hook(hook)
		}
	}
}

// 默认的Debug日志处理器
func Debug() *zerolog.Event {
	return Get("").Debug()
}

// 默认的Info日志处理器
func Info() *zerolog.Event {
	return Get("").Info()
}

// 默认的Warn日志处理器
func Warn() *zerolog.Event {
	return Get("").Warn()
}

// 默认的Error日志处理器
func Error() *zerolog.Event {
	return Get("").Error()
}

// 默认的Fatal日志处理器
func Fatal() *zerolog.Event {
	return Get("").Fatal()
}

// 默认的Panic日志处理器
func Panic() *zerolog.Event {
	return Get("").Panic()
}
