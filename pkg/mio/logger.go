package mio

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Logger interface {
	Info(msg string, pairs ...interface{})
	Warn(msg string, pairs ...interface{})
	Error(msg string, pairs ...interface{})
	Debug(msg string, pairs ...interface{})

	Named(name string) Logger
}

type warnLevel int

const (
	warnLevelInfo warnLevel = iota
	warnLevelWarn
	warnLevelError
	warnLevelDebug
)

const (
	ansiReset   = "\u001B[0m"
	ansiRed     = "\u001B[31m"
	ansiGreen   = "\u001B[32m"
	ansiBlue    = "\u001B[34m"
	ansiMagenta = "\u001B[35m"
)

var (
	infoText  = fmt.Sprintf("%sINFO%s\t", ansiBlue, ansiReset)
	warnText  = fmt.Sprintf("%sWARN%s\t", ansiMagenta, ansiReset)
	errorText = fmt.Sprintf("%sERROR%s\t", ansiRed, ansiReset)
	debugText = fmt.Sprintf("%sDEBUG%s\t", ansiGreen, ansiReset)
)

type logger struct {
	mutex sync.Mutex
	name  string
	Out   io.Writer
}

func NewLogger(out io.Writer) Logger {
	return &logger{
		name: "",
		Out:  out,
	}
}

func NewDefaultLogger() Logger {
	return NewLogger(os.Stderr)
}

func NewDiscardLogger() Logger {
	return NewLogger(io.Discard)
}

func (l *logger) log(level warnLevel, msg string, pairs ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	fields := make(map[string]interface{})
	for i := 0; i < len(pairs); i += 2 {
		if i+1 < len(pairs) {
			fields[pairs[i].(string)] = pairs[i+1]
		}
	}

	b, err := json.Marshal(fields)
	if err != nil {
		fmt.Println("Error marshalling fields: ", err)
	}

	text := fmt.Sprintf("%s\t%s\n", textFromWarnLevel(level), msg)
	if len(pairs) > 0 {
		text = fmt.Sprintf("%s\t%s\t%s\n", textFromWarnLevel(level), msg, string(b))
	}

	_, err = l.Out.Write([]byte(text))
	if err != nil {
		fmt.Println("Error writing to log: ", err)
	}
}

func textFromWarnLevel(level warnLevel) string {
	switch level {
	case warnLevelInfo:
		return infoText
	case warnLevelWarn:
		return warnText
	case warnLevelError:
		return errorText
	case warnLevelDebug:
		return debugText
	default:
		return ""
	}
}

func (l *logger) Info(msg string, pairs ...interface{}) {
	l.log(warnLevelInfo, msg, pairs...)
}

func (l *logger) Warn(msg string, pairs ...interface{}) {
	l.log(warnLevelWarn, msg, pairs...)
}

func (l *logger) Error(msg string, pairs ...interface{}) {
	l.log(warnLevelError, msg, pairs...)
}

func (l *logger) Debug(msg string, pairs ...interface{}) {
	l.log(warnLevelDebug, msg, pairs...)
}

func (l *logger) Named(name string) Logger {
	if name == "" {
		return l
	}

	return &logger{
		name: strings.Join([]string{l.name, name}, "."),
		Out:  l.Out,
	}
}
