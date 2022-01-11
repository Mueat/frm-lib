package db

import "github.com/Mueat/frm-lib/log"

// 日志记录
type DBLogger struct {
}

func (l DBLogger) Printf(format string, args ...interface{}) {
	log.Info().Str("type", "SQL").Msgf(format, args...)
}
