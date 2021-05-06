package logger

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

func Init(level logrus.Level) {
	logger = logrus.New()
	logger.SetLevel(level)
}

func GetLogger() *logrus.Logger {
	return logger
}
