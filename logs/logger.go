package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var once sync.Once

func InitLogger() {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.MessageKey = "msg"
		config.EncoderConfig.LevelKey = "level"

		var err error
		log, err = config.Build()
		if err != nil {
			panic(err)
		}
	})
}

func GetLogger() *zap.Logger {
	if log == nil {
		InitLogger()
	}
	return log
}
