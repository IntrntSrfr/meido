package test

import (
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewTestConfig() *utils.Config {
	conf := utils.NewConfig()
	conf.Set("shards", 1)
	conf.Set("token", "asdf")
	return conf
}

func NewTestLogger() *zap.Logger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.OutputPaths = []string{}
	loggerConfig.ErrorOutputPaths = []string{}
	logger, _ := loggerConfig.Build()
	return logger.Named("test")
}
