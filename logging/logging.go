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

// Logger ... simple struct for logging package, prints with color: Info, Error and Debug level
type Logger struct {
	Logger *log.Logger
}

// Info ... Info level
func (l Logger) Info(params ...string) {
	l.Logger.Printf("%s [INFO] %s %s", yellow, params, reset)
}

// Error ... Error level
func (l Logger) Error(params ...string) {
	l.Logger.Printf("%s [ERROR] %s %s", red, params, reset)
}

// Debug ... Debug level, you have to pass 0 or 1 as first param
func (l Logger) Debug(debug string, params ...string) {
	if debug == "1" {
		l.Logger.Printf("%s [DEBUG] %s %s", green, params, reset)
	}
}
