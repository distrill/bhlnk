package main

import (
	"github.com/sebest/logrusly"
	"github.com/sirupsen/logrus"
)

var logglyToken string = "b2e96194-9e09-4e5c-baf2-fd45aa02f928"
var hook *logrusly.LogglyHook

func init() {
	hook = logrusly.NewLogglyHook(logglyToken, "bhlnk.com", logrus.InfoLevel)
}

// NewLogger initializes the standard logger
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Formatter = &logrus.JSONFormatter{}

	return logger
}

func FlushLogs() {
	hook.Flush()
}
