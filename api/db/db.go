package db

import (
	"agones-minecraft/config"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var db *gorm.DB

func Init() {
	var err error
	logger := zapgorm2.New(zap.L())

	user, pw, host, port, name := config.GetDB()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, pw, host, port, name)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
	if err != nil {
		zap.L().Fatal("error connecting to db", zap.Error(err))
	}

	if err = db.Debug().AutoMigrate(); err != nil {
		zap.L().Fatal("error auto-migrating db", zap.Error(err))
	}

}

func DB() *gorm.DB {
	return db
}
