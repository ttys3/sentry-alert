package main

import (
	"github.com/innogames/slaxy"
	"github.com/sirupsen/logrus"
)

// logrusLogger wraps a logrus logger for compatibility with the slaxy library
type logrusLogger struct {
	slaxy.Logger
	l *logrus.Logger
}

// Debug logs debug messages
func (l *logrusLogger) Debug(msg string) {
	l.l.Debug(msg)
}

// Debugf logs debug messages
func (l *logrusLogger) Debugf(msg string, args ...interface{}) {
	l.l.Debugf(msg, args...)
}

// Info logs debug messages
func (l *logrusLogger) Info(msg string) {
	l.l.Info(msg)
}

// Infof logs debug messages
func (l *logrusLogger) Infof(msg string, args ...interface{}) {
	l.l.Infof(msg, args...)
}

// Warn logs debug messages
func (l *logrusLogger) Warn(msg string) {
	l.l.Warn(msg)
}

// Warnf logs debug messages
func (l *logrusLogger) Warnf(msg string, args ...interface{}) {
	l.l.Warnf(msg, args...)
}

// Error logs debug messages
func (l *logrusLogger) Error(msg string) {
	l.l.Error(msg)
}

// Errorf logs debug messages
func (l *logrusLogger) Errorf(msg string, args ...interface{}) {
	l.l.Errorf(msg, args...)
}
