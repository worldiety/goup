package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var logger Logger = &defaultLogger{}

// The LogLevel determines what to log
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// SetLogger sets the module/package level logger
func SetLogger(log Logger) {
	logger = log
}

// Fields is just a type alias to avoid verbosity while using the logger
type Fields = map[string]interface{}

// A Logger is just a simple interface which can be easily satisfied by any implementor
type Logger interface {
	Debug(fields Fields)
	Info(fields Fields)
	Warn(fields Fields)
	Error(fields Fields)
}

// the default logger just prints as json into stdout
type defaultLogger struct {
	LogLevel LogLevel
}

func (l *defaultLogger) Debug(fields map[string]interface{}) {
	if l.LogLevel > Debug {
		return
	}
	fields["level"] = "DEBUG"
	l.log(fields)
}

func (l *defaultLogger) Info(fields map[string]interface{}) {
	if l.LogLevel > Info {
		return
	}
	fields["level"] = "INFO"
	l.log(fields)
}

func (l *defaultLogger) Warn(fields map[string]interface{}) {
	if l.LogLevel > Warn {
		return
	}
	fields["level"] = "WARN"
	l.log(fields)
}

func (l *defaultLogger) Error(fields map[string]interface{}) {
	if l.LogLevel > Error {
		return
	}
	fields["level"] = "ERROR"
	l.log(fields)
}

func (l *defaultLogger) log(fields map[string]interface{}) {
	time := time.Now()
	sb := &strings.Builder{}
	// 2017-04-20 20:25:42
	sb.WriteString(fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%03d", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond()/1000/1000))
	sb.WriteString(" [")
	sb.WriteString(fields["level"].(string))
	delete(fields, "level")
	sb.WriteString("] - ")

	tmp := make([]string, 0)
	for k, v := range fields {
		tmp = append(tmp, fmt.Sprintf("%s: %v, ", k, v))
	}
	sort.Strings(tmp)
	for _, s := range tmp {
		sb.WriteString(s)
	}
	fmt.Println(sb.String())
}
