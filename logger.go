package flog

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	TimeFormat = "2006-01-02T15-04-05"

	levelInfo  level = "INFO"
	levelError level = "ERROR"
	levelFatal level = "FATAL"
	levelDebug level = "DEBUG"
)

type (
	level string

	Logger struct {
		wg    *sync.WaitGroup
		param struct {
			MaxSize  int64
			MaxAge   int
			Compress bool
		}
		logs chan *Log
		mu   *sync.Mutex
		path string
		file *os.File
		done chan struct{}
	}

	Log struct {
		Level   level          `json:"level"`
		Time    string         `json:"time"`
		Message string         `json:"message"`
		Props   map[string]any `json:"properties,omitempty"`
		Line    int            `json:"line,omitempty"`
		File    string         `json:"file,omitempty"`
		Trace   string         `json:"trace,omitempty"`
	}

	Field struct {
		Name string `json:"name"`
		Val  any    `json:"value"`
	}
)

func (l level) String() string {
	return string(l)
}

func NewField(name string, v any) Field {
	return Field{
		Name: name,
		Val:  v,
	}
}

func NewLogger(path string, maxAge, maxSize int) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), fs.ModePerm); err != nil {
		return nil, err
	}

	// open or create log
	log, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file: %w", err)
	}

	logger := &Logger{
		wg: &sync.WaitGroup{},
		param: struct {
			MaxSize  int64
			MaxAge   int
			Compress bool
		}{
			MaxSize:  int64(maxSize),
			MaxAge:   maxAge,
			Compress: false,
		},
		logs: make(chan *Log, 2048),
		mu:   new(sync.Mutex),
		path: path,
		file: log,
		done: make(chan struct{}, 1),
	}

	logger.wg.Add(1)
	go logger.loop()

	return logger, nil
}

func (l *Logger) loop() {
	defer l.wg.Done()

	tomorrow := time.Now().AddDate(0, 0, 1)

	for {
		select {
		case <-l.done:
			return
		case <-time.After(time.Until(tomorrow)):
			l.mu.Lock()
			if err := l.rotate(); err != nil {
				l.logs <- l.newLog(levelError, err.Error())
			}
			l.mu.Unlock()
			tomorrow = time.Now().AddDate(0, 0, 1)

		case log := <-l.logs:
			l.mu.Lock()
			if err := l.write(log); err != nil {
				l.logs <- l.newLog(levelError, err.Error())
			}
			l.mu.Unlock()
		}
	}
}
func (l *Logger) rotate() error {
	l.file.Close()

	base := filepath.Base(l.path)

	parts := strings.Split(base, ".")
	backupName := fmt.Sprintf("%s_%s.%s", parts[0], time.Now().Format(TimeFormat), parts[1])
	backupName = filepath.Join(filepath.Dir(l.path), backupName)
	if err := os.Rename(l.path, backupName); err != nil {
		return fmt.Errorf("unable to rename log file: %w", err)
	}

	if l.param.Compress {
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			compressLog(backupName)
		}()
	}

	var err error
	l.file, err = os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("unable to open log file: %w", err)
	}

	// Cleanup old logs
	files, err := filepath.Glob(fmt.Sprintf("%s.*", l.path))
	if err != nil {
		return fmt.Errorf("unable to glob log files: %w", err)
	}

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("unable to get file info: %w", err)
		}
		if time.Since(fi.ModTime()).Hours() > float64(24*l.param.MaxAge) {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("unable to remove log file: %w", err)
			}
		}
	}

	return nil
}
func (l *Logger) write(log *Log) error {
	fi, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file info: %w", err)
	}

	// Check file size
	if fi.Size() > l.param.MaxSize*1024*1024 {
		l.rotate()
	}

	row, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("unable to marshal log: %w", err)
	}

	if _, err := l.file.Write(append(row, '\n')); err != nil {
		return fmt.Errorf("unable to write log: %w", err)
	}

	return nil
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.logs <- l.newLog(levelInfo, msg, fields...)
}
func (l *Logger) Error(msg string, fields ...Field) {
	l.logs <- l.newLog(levelError, msg, fields...)
}
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.logs <- l.newLog(levelFatal, msg, fields...)
}
func (l *Logger) Debug(m string, fields ...Field) {
	l.logs <- l.newLog(levelDebug, m, fields...)
}

func (l *Logger) Close() {
	close(l.done)
	close(l.logs)
	l.wg.Wait()
	l.file.Close()
}

func (l *Logger) newLog(level level, msg string, fields ...Field) *Log {
	var props = map[string]any{}

	for _, f := range fields {
		props[f.Name] = f.Val
	}

	log := &Log{
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

func compressLog(filename string) {
	inputFile, err := os.Open(filename)
	if err != nil {
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(filename + ".gz")
	if err != nil {
		return
	}
	defer outputFile.Close()

	writer := gzip.NewWriter(outputFile)
	defer writer.Close()

	io.Copy(writer, inputFile)

	os.Remove(filename)
}
