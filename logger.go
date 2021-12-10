package flog

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	DateFormat = "02-01-2006"            // dd-mm-yyyy
	TimeFormat = "02-Jan-2006 15h04m05s" // dd-mm-yyyy hhmmss
)

// Levels
type level string

//
const (
	levelInfo  level = "INFO"
	levelError level = "ERROR"
	levelFetal level = "FETAL"
)

// Logger
type Logger struct {
	mu   *sync.Mutex
	file *os.File
}

// NewLogger
func NewLogger(logDir string) (*Logger, error) {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, fs.ModePerm); err != nil {
			return nil, err
		}
	}

	// rotate log everyday
	todaysdate := time.Now().Format(DateFormat)
	filename := "app_" + todaysdate + ".log"
	logPath := path.Join(logDir, filename)

	// open or create log
	log, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		mu:   &sync.Mutex{},
		file: log,
	}, nil
}

// Log
type Log struct {
	Time       string            `json:"time"`
	Level      string            `json:"level"`
	Message    string            `json:"message"`
	Line       int               `json:"line"`
	File       string            `json:"file"`
	Properties map[string]string `json:"properties,omitempty"`
	Trace      string            `json:"trace,omitempty"`
}

// print
func (l *Logger) print(level level, message string, props map[string]string) (int, error) {
	_, filename, line, _ := runtime.Caller(2)
	// If the severity level of the log entry is below the minimum severity for the
	// log row
	log := Log{
		Time:       time.Now().UTC().Format(TimeFormat),
		Level:      string(level),
		Message:    message,
		Line:       line,
		File:       filename,
		Properties: props,
	}
	// stack trace
	if level == levelFetal {
		log.Trace = string(debug.Stack())
	}

	// Declare a line variable for holding the actual log entry text.
	var row []byte
	// Marshal the anonymous struct to JSON and store it in the line variable. If there // was a problem creating the JSON, set the contents of the log entry to be that
	// plain-text error message instead.
	row, err := json.Marshal(log)
	if err != nil {
		row = []byte(string(levelError) + ": unable to marshal log message:" + err.Error())
	}
	// Lock the mutex so that no two writes to the output destination cannot happen // concurrently. If we don't do this, it's possible that the text for two or more // log entries will be intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write the log entry followed by a newline.
	return l.file.Write(append(row, '\n'))
}

// PrintInfo
func (l *Logger) Info(message string, props map[string]string) {
	l.print(levelInfo, message, props)
}

// PrintError
func (l *Logger) Error(message string, props map[string]string) {
	l.print(levelError, message, props)
}

// Fetal
func (l *Logger) Fetal(message string, props map[string]string) {
	l.print(levelFetal, message, props)
}

// Close
func (l *Logger) Close() {
	l.file.Close()
}
