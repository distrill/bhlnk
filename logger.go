package main

import (
	"github.com/sirupsen/logrus"
)

// NewLogger initializes the standard logger
func NewLogger() *logrus.Logger {
	var logger = logrus.New()

	logger.Formatter = &logrus.JSONFormatter{}

	return logger
}
