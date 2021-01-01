package logging

import (
	"log"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

// Logger ...
type Logger struct {
	Logger *log.Logger
}

// Info ...
func (l Logger) Info(params ...string) {
	l.Logger.Printf("%s [INFO] %s %s", yellow, params, reset)
}

// Error ...
func (l Logger) Error(params ...string) {
	l.Logger.Printf("%s [ERROR] %s %s", red, params, reset)
}

// Debug ...
func (l Logger) Debug(params ...string) {
	l.Logger.Printf("%s [DEBUG] %s %s", green, params, reset)
}
