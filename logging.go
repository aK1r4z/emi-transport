package emi_transport

import (
	"fmt"
	"strings"
	"time"
)

type (
	logLevel int
)

const (
	logLevelTrace logLevel = 0 + iota
	logLevelDebug
	logLevelInfo
	logLevelWarn
	logLevelError
	logLevelFatal
)

func (l logLevel) String() string {
	switch l {
	case logLevelTrace:
		return "TRACE"
	case logLevelDebug:
		return "DEBUG"
	case logLevelInfo:
		return "INFO"
	case logLevelWarn:
		return "WARN"
	case logLevelError:
		return "ERROR"
	case logLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type TinyLogger struct {
	name string
}

func NewTinyLogger(name string) *TinyLogger {
	return &TinyLogger{
		name: name,
	}
}

func (l *TinyLogger) logF(logLevel logLevel, format string, args ...any) {
	format = strings.TrimRight(format, "\n")

	levelString := "[" + logLevel.String() + "]"
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	nameString := "[" + l.name + "]"

	logString := fmt.Sprintf(
		"%s %7s %s: "+format+"\n",
		append([]any{timeString, levelString, nameString}, args...)...,
	)

	fmt.Print(logString)
}

func (l *TinyLogger) Tracef(format string, args ...any) {
	l.logF(logLevelTrace, format, args...)
}

func (l *TinyLogger) Debugf(format string, args ...any) {
	l.logF(logLevelDebug, format, args...)
}

func (l *TinyLogger) Infof(format string, args ...any) {
	l.logF(logLevelInfo, format, args...)
}

func (l *TinyLogger) Warnf(format string, args ...any) {
	l.logF(logLevelWarn, format, args...)
}

func (l *TinyLogger) Errorf(format string, args ...any) {
	l.logF(logLevelError, format, args...)
}

func (l *TinyLogger) Fatalf(format string, args ...any) {
	l.logF(logLevelFatal, format, args...)
}

func (l *TinyLogger) log(level logLevel, args ...any) {
	l.logF(level, "%s", args...)
}

func (l *TinyLogger) Trace(args ...any) {
	l.log(logLevelTrace, args...)
}

func (l *TinyLogger) Debug(args ...any) {
	l.log(logLevelDebug, args...)
}

func (l *TinyLogger) Info(args ...any) {
	l.log(logLevelInfo, args...)
}

func (l *TinyLogger) Warn(args ...any) {
	l.log(logLevelWarn, args...)
}

func (l *TinyLogger) Error(args ...any) {
	l.log(logLevelError, args...)
}

func (l *TinyLogger) Fatal(args ...any) {
	l.log(logLevelFatal, args...)
}
