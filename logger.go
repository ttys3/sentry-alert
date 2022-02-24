package slaxy

// Logger defines simple methods for info and debugging messages
// that could be of interest for the outside
type Logger interface {
	Debug(args ...interface{})
	Debugf(string, ...interface{})
	Info(args ...interface{})
	Infof(string, ...interface{})
	Warn(args ...interface{})
	Warnf(string, ...interface{})
	Error(args ...interface{})
	Errorf(string, ...interface{})
}

// nullLogger is a logger that does nothing
type nullLogger struct{}

// NewNullLogger returns a new instance of a logger that does nothing
func NewNullLogger() Logger {
	return nullLogger{}
}

// Debug logs debug messages
func (l nullLogger) Debug(args ...interface{}) {
}

// Debugf logs debug messages
func (l nullLogger) Debugf(string, ...interface{}) {
}

// Info logs info messages
func (l nullLogger) Info(args ...interface{}) {
}

// Infof logs info messages
func (l nullLogger) Infof(string, ...interface{}) {
}

// Warn logs info messages
func (l nullLogger) Warn(args ...interface{}) {
}

// Warnf logs info messages
func (l nullLogger) Warnf(string, ...interface{}) {
}

// Error logs info messages
func (l nullLogger) Error(args ...interface{}) {
}

// Errorf logs info messages
func (l nullLogger) Errorf(string, ...interface{}) {
}
