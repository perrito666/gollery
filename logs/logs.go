package logs

import (
	"fmt"
	"log"
	"os"
)

// Level represents a level of logging
type Level int

const (
	// LvlError represets log level Error
	LvlError Level = iota
	// LvlWarning represets log level Warning
	LvlWarning
	// LvlInfo represets log level Info
	LvlInfo
	// LvlDebug represets log level Debug
	LvlDebug
	// LvlTrace represets log level Trace
	LvlTrace
)

// Logger wraps a bunh of standard loggers named after the
// expected log levels.
type Logger struct {
	level   Level
	trace   *log.Logger
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	lerror  *log.Logger
}

// SetLevel sets the logging to the passed level
func (l *Logger) SetLevel(lvl Level) {
	l.level = lvl
}

// Tracef outputs formatted log message in trace level
func (l *Logger) Tracef(message string, args ...interface{}) {
	if l.level < LvlTrace {
		return
	}
	l.trace.Printf(message, args...)
}

// Trace outputs log message in trace level
func (l *Logger) Trace(message string) {
	if l.level < LvlTrace {
		return
	}
	l.trace.Print(message)
}

// Debugf outputs formatted log message in debug level
func (l *Logger) Debugf(message string, args ...interface{}) {
	if l.level < LvlDebug {
		return
	}
	l.trace.Printf(message, args...)
}

// Debug outputs log message in debug level
func (l *Logger) Debug(message string) {
	if l.level < LvlDebug {
		return
	}
	l.trace.Print(message)
}

// Infof outputs formatted log message in info level
func (l *Logger) Infof(message string, args ...interface{}) {
	if l.level < LvlInfo {
		return
	}
	l.trace.Printf(message, args...)
}

// Info outputs log message in info level
func (l *Logger) Info(message string) {
	if l.level < LvlInfo {
		return
	}
	l.trace.Print(message)
}

// Warningf outputs formatted log warning in trace level
func (l *Logger) Warningf(message string, args ...interface{}) {
	if l.level < LvlWarning {
		return
	}
	l.trace.Printf(message, args...)
}

// Warning outputs log message in warning level
func (l *Logger) Warning(message string) {
	if l.level < LvlWarning {
		return
	}
	l.trace.Print(message)
}

// Errorf outputs formatted log message in error level
func (l *Logger) Errorf(message string, args ...interface{}) {
	if l.level < LvlError {
		return
	}
	l.trace.Printf(message, args...)
}

// Error outputs log message in error level
func (l *Logger) Error(message string) {
	if l.level < LvlError {
		return
	}
	l.trace.Print(message)
}

// New return a new initialized logger.
func New(prefix string) *Logger {
	Trace := log.New(os.Stdout,
		fmt.Sprintf("[%s] TRACE: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)
	Debug := log.New(os.Stdout,
		fmt.Sprintf("[%s] DEBUG: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	Info := log.New(os.Stdout,
		fmt.Sprintf("[%s] INFO: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning := log.New(os.Stdout,
		fmt.Sprintf("[%s] WARNING: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	Error := log.New(os.Stderr,
		fmt.Sprintf("[%s] ERROR: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	return &Logger{
		trace:   Trace,
		debug:   Debug,
		info:    Info,
		warning: Warning,
		lerror:  Error,
	}
}
