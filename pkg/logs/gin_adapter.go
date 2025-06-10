package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type GinLoggerAdapter struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	level  zapcore.Level
}

func NewGinLoggerAdapter(level zapcore.Level) *GinLoggerAdapter {
	return &GinLoggerAdapter{
		logger: Logger,
		sugar:  Sugar,
		level:  level,
	}
}
func (w GinLoggerAdapter) Write(p []byte) (n int, err error) {
	switch w.level {
	case zapcore.InfoLevel:
		w.logger.Info(string(p[:len(p)-1])) // 去除末尾的换行符
	case zapcore.ErrorLevel:
		w.logger.Error(string(p[:len(p)-1]))
	default:
		w.logger.Info(string(p[:len(p)-1]))
	}
	return len(p), nil
}
