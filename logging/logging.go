package logging

import (
	"log"
)

// Logger ...
type Logger struct {
	Logger *log.Logger
}

// Info ...
func (l Logger) Info(params ...string) {
	l.Logger.Println(params)
}
