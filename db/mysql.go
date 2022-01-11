package db

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/Mueat/frm-lib/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var configs map[string]MysqlConfig
var mysqlConnections map[string]*gorm.DB
var connectOnce = sync.Once{}

type MysqlConfig struct {
	Host            string
	Username        string
	Password        string
	DBName          string
	Charset         string
	Location        string
	MaxOpen         int64
	MaxIdle         int64
	ConnMaxLifetime int64
	SlowThreshold   int64
	LogLevel        int64
	Default         bool
}

// Connect 连接到mysql
func ConnectMysql(confs map[string]MysqlConfig) {
	configs = confs
	mysqlConnections = make(map[string]*gorm.DB)
	connectOnce.Do(func() {
		for k, conf := range configs {
			newLogger := logger.New(
				DBLogger{}, // io writer
				logger.Config{
					SlowThreshold: time.Duration(conf.SlowThreshold) * time.Millisecond, // 慢 SQL 阈值
					LogLevel:      logger.LogLevel(conf.LogLevel),                       // Log level
				},
			)
			dns := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=%s&loc=%s", conf.Username, conf.Password, conf.Host, conf.DBName, conf.Charset, url.QueryEscape(conf.Location))
			db, err := gorm.Open(mysql.Open(dns), &gorm.Config{
				Logger: newLogger,
			})
			if err != nil {
				log.Fatal().Msgf("connect mysql err : %s", err.Error())
				panic(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				panic(err)
			}
			sqlDB.SetMaxIdleConns(int(conf.MaxIdle))
			sqlDB.SetMaxOpenConns(int(conf.MaxOpen))
			sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)

			mysqlConnections[k] = db
		}
	})
}

// 获取连接
func GetMySql(name string) *gorm.DB {
	if name != "" {
		if conn, ok := mysqlConnections[name]; ok {
			return conn
		}
	}

	for k, conf := range configs {
		if conf.Default {
			if conn, ok := mysqlConnections[k]; ok {
				return conn
			}
		}
	}
	return nil
}

// 获取默认连接
func GetDefaultMysql() *gorm.DB {
	return GetMySql("")
}

// 获取configs
func GetMysqlConfigs() map[string]MysqlConfig {
	return configs
}

// 获取config
func GetMysqlConfig(name string) MysqlConfig {
	return configs[name]
}
