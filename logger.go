package flog

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
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
	mu     *sync.Mutex
	dir    string
	prefix string
	file   *os.File
	next   time.Time
}

// NewLogger
func NewLogger(dir, prefix string) (*Logger, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
			return nil, err
		}
	}
	fpath, tomorrow := generate(dir, prefix)
	// open or create log
	log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		mu:     &sync.Mutex{},
		dir:    dir,
		prefix: prefix,
		file:   log,
		next:   tomorrow,
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
func (l *Logger) print(level level, message string, props map[string]string) {
	// Lock the mutex so that no two writes to the output destination cannot happen // concurrently. If we don't do this, it's possible that the text for two or more // log entries will be intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()

	//
	if time.Now() == l.next {
		// generate new file
		fpath, tomorrow := generate(l.dir, l.prefix)

		l.next = tomorrow

		log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			l.file.Write([]byte(err.Error() + "\n"))
			return
		}
		// close current file
		l.file.Close()

		l.file = log

		// remove old files
		if err := remove(l.dir, l.prefix, 30); err != nil {
			l.file.Write([]byte(err.Error() + "\n"))
			return
		}
	}

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

	// Write the log entry followed by a newline.
	l.file.Write(append(row, '\n'))
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

// generate - generate current file name and time left for tomorrow
func generate(dir, prefix string) (string, time.Time) {
	now := time.Now()
	filename := strings.Join([]string{
		prefix,
		now.Format(DateFormat),
	}, "_")
	filename = filename + ".log"

	// tomorrow
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	return path.Join(dir, filename), tomorrow
}

// remove - removes log files older then 30 days
func remove(dir, prefix string, days int) error {
	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range fi {
		if time.Since(f.ModTime()) > time.Duration(days)*24*time.Hour {
			if err := os.Remove(path.Join(dir, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}
