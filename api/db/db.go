package db

import (
	"context"
	"net"
	"time"

	"github.com/go-pg/pg/v10"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"

	"agones-minecraft/config"
)

const (
	DefaultThreshold time.Duration = time.Millisecond * 500
)

var db *pg.DB

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

	pg := New()

	if err := pg.Ping(context.Background()); err != nil {
		zap.L().Fatal("error pinging database", zap.Error(err))
	}

	pg.AddQueryHook(NewLogger(zap.L()))

	db = pg
}

func DB() *pg.DB {
	return db
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return db.Ping(ctx)
}
