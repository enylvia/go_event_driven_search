package util

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{log.New(os.Stdout, "[PUBLISHER] ", log.Ldate|log.Ltime|log.Lshortfile)}
}
