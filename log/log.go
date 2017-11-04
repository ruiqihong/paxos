package log

import (
	"log"
)

type F map[string]interface{}

type Logger interface {
	With(fields F) Logger
	Debug(string)
	Info(string)
	Warn(string)
	Err(string)
}

type DFLogger struct {
	fields F
}

func (l *DFLogger) With(fields F) Logger {
	return &DFLogger{fields: fields}
}

func (l *DFLogger) Debug(s string) {
	if l.fields != nil {
		log.Printf("DEBUG: %s: %+v", s, l.fields)
	} else {
		log.Printf("DEBUG: %s", s)
	}
}

func (l *DFLogger) Info(s string) {
	if l.fields != nil {
		log.Printf("INFO: %s: %+v", s, l.fields)
	} else {
		log.Printf("INFO: %s", s)
	}
}

func (l *DFLogger) Warn(s string) {
	if l.fields != nil {
		log.Printf("WARN: %s: %+v", s, l.fields)
	} else {
		log.Printf("WARN: %s", s)
	}
}

func (l *DFLogger) Err(s string) {
	if l.fields != nil {
		log.Printf("ERR: %s: %+v", s, l.fields)
	} else {
		log.Printf("ERR: %s", s)
	}
}

var gLogger Logger = &DFLogger{}

func SetLogger(logger Logger) {
	gLogger = logger
}

func With(fields F) Logger {
	return gLogger.With(fields)
}

func Debug(s string) {
	gLogger.Debug(s)
}

func Info(s string) {
	gLogger.Info(s)
}

func Warn(s string) {
	gLogger.Warn(s)
}

func Err(s string) {
	gLogger.Err(s)
}
