package flog

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	DateFormat = "02-01-2006" // dd-mm-yyyy
)

// cores
var cores = func() int {
	c := runtime.NumCPU()
	if c <= 1 {
		return 1
	}
	return c - 1
}()

// Levels
type level string

const (
	levelInfo  level = "INFO"
	levelError level = "ERROR"
	levelFatal level = "FATAL"
	levelDebug level = "DEBUG"
)

// string
func (l level) String() string {
	return string(l)
}

// Logger
type Logger struct {
	wg     *sync.WaitGroup
	conf   *Config
	next   *time.Timer // timer for next rotation
	logs   chan Log
	mu     *sync.Mutex
	file   *os.File
	ctx    context.Context
	cancel context.CancelFunc
}

// NewLogger
func NewLogger(cnf *Config) (*Logger, error) {
	// TODO: review.
	// days must be greater than 0
	if cnf.Rotate < 7 {
		cnf.Rotate = 7
	}

	// make sure dir exists
	if _, err := os.Stat(cnf.Dir); os.IsNotExist(err) {
		if err := os.MkdirAll(cnf.Dir, fs.ModePerm); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	fpath := logFile(now, cnf.Dir, cnf.Prefix)
	left := timeLeft(now)

	// open or create log
	log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := &Logger{
		wg:     &sync.WaitGroup{},
		conf:   cnf,
		next:   time.NewTimer(left),
		logs:   make(chan Log, cores),
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
	Level   level          `json:"level"`
	Time    string         `json:"time"`
	Message string         `json:"message"`
	Props   map[string]any `json:"properties,omitempty"`
	Line    int            `json:"line,omitempty"`
	File    string         `json:"file,omitempty"`
	Trace   string         `json:"trace,omitempty"`
}

// newLog
func (l *Logger) newLog(level level, msg string, fields ...Field) Log {
	var props = map[string]any{}

	for _, f := range fields {
		props[f.Key] = f.Val
	}

	log := Log{
		Time:    time.Now().Format(time.RFC3339),
		Level:   level,
		Message: msg,
		Props:   props,
	}

	// debug
	if log.Level == levelDebug {
		_, filename, line, _ := runtime.Caller(2)

		log.File = filename
		log.Line = line
	}

	// stack trace
	if log.Level == levelFatal {
		_, filename, line, _ := runtime.Caller(2)

		log.File = filename
		log.Line = line

		log.Trace = string(debug.Stack())
	}

	return log
}

// run
func (l *Logger) run() {
	defer func() {
		l.next.Stop()
		l.wg.Done()
	}()

	for {
		select {
		// case <-l.ctx.Done():
		// 	fmt.Println("ctx done")
		// 	return
		case now := <-l.next.C:
			l.rotate(now)
		case log, ok := <-l.logs:
			// to avoid sending `nil` log
			// on last line
			if !ok {
				return
			}

			l.write(l.ctx, log)
		}
	}
}

// rotate
func (l *Logger) rotate(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fpath := logFile(now, l.conf.Dir, l.conf.Prefix)
	left := timeLeft(now)

	log, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		l.logs <- l.newLog(levelError, "open_file", NewField("rotate", err))
		return
	}

	// close old file
	l.file.Close()

	// set new file
	l.file = log
	l.next.Reset(left)

	if err := remove(l.conf.Dir, l.conf.Prefix, l.conf.Rotate); err != nil {
		l.logs <- l.newLog(levelError, "remove", NewField("rotate", err))
		return
	}
}

// print
func (l *Logger) write(ctx context.Context, log Log) {
	if ctx.Err() != nil {
		return
	}

	row, err := json.Marshal(log)
	if err != nil {
		row = []byte(string(levelError) + ": unable to marshal log message:" + err.Error())
	}

	// Write the log entry followed by a newline.
	if _, err := l.file.Write(append(row, '\n')); err != nil {
		// TODO: review this.
		l.logs <- l.newLog(levelError, "write", NewField("file_write", err))
	}
}

// PrintInfo
func (l *Logger) Info(msg string, fields ...Field) {
	l.logs <- l.newLog(levelInfo, msg, fields...)
}

// PrintError
func (l *Logger) Error(msg string, fields ...Field) {
	l.logs <- l.newLog(levelError, msg, fields...)
}

// Fatal
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.logs <- l.newLog(levelFatal, msg, fields...)
}

// Debug
func (l *Logger) Debug(m string, fields ...Field) {
	l.logs <- l.newLog(levelDebug, m, fields...)
}

// Close
func (l *Logger) Close() {
	close(l.logs)
	l.wg.Wait()
	l.cancel()
	l.file.Close()
}

// helper func \\

// logFile - generate current file name
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
	fi, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range fi {
		info, err := f.Info()
		if err != nil {
			return err
		}

		if time.Since(info.ModTime()) > time.Duration(days)*24*time.Hour {
			if err := os.Remove(path.Join(dir, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}
