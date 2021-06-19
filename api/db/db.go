package db

import (
	"context"
	"net"
	"time"

	"github.com/go-pg/pg/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"

	"agones-minecraft/config"
)

const (
	DefaultThreshold time.Duration = time.Millisecond * 500
)

var db *gorm.DB

func New() *pg.DB {
	dbconfig := config.GetDBConfig()

	return pg.Connect(&pg.Options{
		Addr:     net.JoinHostPort(dbconfig.Hostname, dbconfig.Port),
		User:     dbconfig.User,
		Password: dbconfig.Password,
		Database: dbconfig.Name,
	})
}

func Init() {
	logger := zapgorm2.New(zap.L())
	logger.SlowThreshold = DefaultThreshold

	if config.GetEnv() == config.Production {
		logger.IgnoreRecordNotFoundError = true
	} else {
		logger.LogLevel = gormlogger.Info
	}

	db := New()

	if err := db.Ping(context.Background()); err != nil {
		zap.L().Fatal("error pinging database", zap.Error(err))
	}

	db.AddQueryHook(NewLogger(zap.L()))

}

func DB() *gorm.DB {
	return db
}
