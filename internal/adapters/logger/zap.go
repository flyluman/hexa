package logger

import (
	"portfolio/internal/core/ports"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	log *zap.SugaredLogger
}

func NewZapLogger() (ports.Logger, error) {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			MessageKey:    "msg",
			CallerKey:     "file",
			StacktraceKey: "stacktrace",
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	base, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &zapLogger{log: base.Sugar()}, nil
}

func (l *zapLogger) Debug(msg string, kv ...any) {
	l.log.Debugw(msg, kv...)
}

func (l *zapLogger) Info(msg string, kv ...any) {
	l.log.Infow(msg, kv...)
}

func (l *zapLogger) Warn(msg string, kv ...any) {
	l.log.Warnw(msg, kv...)
}

func (l *zapLogger) Error(msg string, kv ...any) {
	l.log.Errorw(msg, kv...)
}
