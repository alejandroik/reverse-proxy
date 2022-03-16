package logger

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(message string) {
	logger.Sugar().Info(message)
}

func Infof(message string, args ...interface{}) {
	logger.Sugar().Infof(message, args...)
}

func Error(message string) {
	logger.Sugar().Error(message)
}

func Fatal(args ...interface{}) {
	logger.Sugar().Fatal(args...)
}