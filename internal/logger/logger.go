package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

// Init initializes the global logger based on environment.
func Init(env string) {
	var err error

	if env == "production" {
		Log, err = zap.NewProduction()
	} else {
		Log, err = zap.NewDevelopment()
	}

	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}
