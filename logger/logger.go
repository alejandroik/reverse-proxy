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
	logger.Info(message)
}

func Error(message string) {
	logger.Error(message)
}

func Fatal(message string) {
	logger.Fatal(message)
}