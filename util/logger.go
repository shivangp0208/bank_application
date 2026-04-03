package util

import "log"

var logger *log.Logger

func init() {
	logger = log.Default()
}

func GetLogger() *log.Logger {
	return logger
}
