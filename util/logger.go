package util

import (
	"log"
)

var logger *log.Logger

func init() {
	logger = log.Default()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func GetLogger() *log.Logger {
	return logger
}
