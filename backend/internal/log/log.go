package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) *zap.Logger {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}
	logger.Info("logger initialized", zap.String("env", env), zap.String("pid", procID()))
	return logger
}

func procID() string {
	if v := os.Getenv("HOSTNAME"); v != "" {
		return v
	}
	return "local"
}
