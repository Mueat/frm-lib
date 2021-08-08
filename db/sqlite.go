package db

import (
	"time"

	"gitee.com/Rainkropy/frm-lib/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sqliteConfigs map[string]SqliteConfig
var sqliteConnections map[string]*gorm.DB

type SqliteConfig struct {
	DBPath          string
	MaxOpen         int64
	MaxIdle         int64
	ConnMaxLifetime int64
	SlowThreshold   int64
	LogLevel        int64
	Default         bool
}

// Connect 连接到sqlite
func ConnectSqlite(confs map[string]SqliteConfig) {
	sqliteConfigs = confs
	sqliteConnections = make(map[string]*gorm.DB)
	connectOnce.Do(func() {
		for k, conf := range sqliteConfigs {
			newLogger := logger.New(
				DBLogger{}, // io writer
				logger.Config{
					SlowThreshold: time.Duration(conf.SlowThreshold) * time.Millisecond, // 慢 SQL 阈值
					LogLevel:      logger.LogLevel(conf.LogLevel),                       // Log level
				},
			)
			db, err := gorm.Open(sqlite.Open(conf.DBPath), &gorm.Config{
				Logger: newLogger,
			})
			if err != nil {
				log.Fatal().Msgf("connect sqlite err : %s", err.Error())
				panic(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				panic(err)
			}
			sqlDB.SetMaxIdleConns(int(conf.MaxIdle))
			sqlDB.SetMaxOpenConns(int(conf.MaxOpen))
			sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)

			sqliteConnections[k] = db
		}
	})
}

// 获取连接
func GetSqlite(name string) *gorm.DB {
	if name != "" {
		if conn, ok := sqliteConnections[name]; ok {
			return conn
		}
	}

	for k, conf := range sqliteConfigs {
		if conf.Default {
			if conn, ok := sqliteConnections[k]; ok {
				return conn
			}
		}
	}
	return nil
}
