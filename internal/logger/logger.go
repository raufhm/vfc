package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(development bool) (*zap.Logger, error) {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	return config.Build()
}
