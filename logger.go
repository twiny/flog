package flog

import (
	"context"
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

// threads
var cores = func() int {
	c := runtime.NumCPU()
	if c <= 1 {
		return 1
	}
	return c - 1
}()

const (
	DateFormat = "02-01-2006" // dd-mm-yyyy
)

// Levels
type level string

//
const (
	levelInfo  level = "INFO"
	levelError level = "ERROR"
	levelFatal level = "FATAL"
)

// string
func (l level) String() string {
	return string(l)
}

// Logger
type Logger struct {
	wg     *sync.WaitGroup
	mu     *sync.Mutex
	dir    string
	prefix string
	days   int         // days to keep logs
	next   *time.Timer // timer for next rotation
	logs   chan *Log
	file   *os.File
	ctx    context.Context
	cancel context.CancelFunc
}

// NewLogger
func NewLogger(dir, prefix string, days int) (*Logger, error) {
	// days must be greater than 0
	if days < 7 {
		days = 7
	}

	// make sure dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	fpath := logFile(now, dir, prefix)
	left := timeLeft(now)

	// open or create log
	log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := &Logger{
		wg:     &sync.WaitGroup{},
		dir:    dir,
		prefix: prefix,
		next:   time.NewTimer(left),
		logs:   make(chan *Log, cores),
		days:   days,
		mu:     &sync.Mutex{},
		file:   log,
		ctx:    ctx,
		cancel: cancel,
	}

	logger.wg.Add(1)
	go logger.run()

	return logger, nil
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

// newLog
func newLog(level level, message string, props map[string]string) *Log {
	_, filename, line, _ := runtime.Caller(2)

	log := Log{
		Time:       time.Now().Format(time.RFC3339),
		Level:      level.String(),
		Message:    message,
		Line:       line,
		File:       filename,
		Properties: props,
	}

	// stack trace
	if log.Level == levelFatal.String() {
		log.Trace = string(debug.Stack())
	}

	return &log
}

// run
func (l *Logger) run() {
	defer func() {
		l.next.Stop()
		l.wg.Done()
	}()

	for {
		select {
		case <-l.ctx.Done():
			return
		case now := <-l.next.C:
			l.rotate(now)
		case log, ok := <-l.logs:
			if !ok {
				return
			}
			l.write(log)
		}
	}
}

// rotate
func (l *Logger) rotate(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fpath := logFile(now, l.dir, l.prefix)
	left := timeLeft(now)

	log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		l.logs <- newLog(levelError, "unable to open log file: "+err.Error(), nil)
		return
	}

	// close old file
	l.file.Close()

	// set new file
	l.file = log
	l.next.Reset(left)

	if err := remove(l.dir, l.prefix, l.days); err != nil {
		l.logs <- newLog(levelError, "unable to remove log file: "+err.Error(), nil)
		return
	}
}

// print
func (l *Logger) write(log *Log) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var row []byte
	row, err := json.Marshal(log)
	if err != nil {
		row = []byte(string(levelError) + ": unable to marshal log message:" + err.Error())
	}

	// Write the log entry followed by a newline.
	l.file.Write(append(row, '\n'))
}

// PrintInfo
func (l *Logger) Info(message string, props map[string]string) {
	l.logs <- newLog(levelInfo, message, props)
}

// PrintError
func (l *Logger) Error(message string, props map[string]string) {
	l.logs <- newLog(levelError, message, props)
}

// Fatal
func (l *Logger) Fatal(message string, props map[string]string) {
	l.logs <- newLog(levelFatal, message, props)
}

// Close
func (l *Logger) Close() {
	l.cancel()    // first exit run() loop
	close(l.logs) // then close channel
	l.wg.Wait()
	l.file.Close()
}

// helper func \\

// logFile - generate current file name and time left for tomorrow
func logFile(now time.Time, dir, prefix string) string {
	filename := strings.Join([]string{
		prefix,
		now.Format(DateFormat),
	}, "_")
	filename = filename + ".log"

	return path.Join(dir, filename)
}

// timeLeft - time left for tomorrow
func timeLeft(now time.Time) time.Duration {
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tomorrow.Sub(now)
}

// remove - removes log files older then x days
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
