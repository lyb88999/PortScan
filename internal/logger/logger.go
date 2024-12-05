package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct{}

func (l *Logger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))

}
func (l *Logger) Debug(args ...interface{}) {
	l.Print(zerolog.DebugLevel, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Print(zerolog.InfoLevel, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Print(zerolog.WarnLevel, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Print(zerolog.ErrorLevel, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Print(zerolog.FatalLevel, args...)
}
