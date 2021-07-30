package logger

import (
	"log"
	"os"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

type Logger interface {
	Info(message string, fields ...string)
	Error(message string, err error)
	WithValues(fields ...string) Logger
}

type zapLogger struct {
	*zap.Logger
}

func New() Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return &zapLogger{logger}
}

func (z *zapLogger) Info(message string, fields ...string) {
	if len(fields) == 0 {
		z.Logger.Info(message)
		return
	}
	f := make([]zap.Field, len(fields))
	for i := 0; i < len(fields); i++ {
		f[i].Type = zapcore.StringType
		f[i].String = fields[i]
	}
	z.Logger.Info(message, f...)
}

func (z *zapLogger) Error(message string, err error) {
	z.Logger.Error(message, zap.Error(err))
}

func (z *zapLogger) WithValues(fields ...string) Logger {
	if len(fields) == 0 {
		return z
	}
	f := make([]zap.Field, len(fields))
	for i := 0; i < len(fields); i++ {
		f[i].Type = zapcore.StringType
		f[i].String = fields[i]
	}
	with := z.Logger.With(f...)
	return &zapLogger{with}
}
