package log

import (
	"agones-minecraft/config"
	"log"

	"go.uber.org/zap"
)

// Sets global zap logger
func SetLog() {
	var logger *zap.Logger
	var err error

	if config.GetEnv() == config.Production {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatal(err)
	}

	zap.ReplaceGlobals(logger)
}
