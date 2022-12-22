package flog

import (
	"strconv"
	"testing"
)

func TestLogInfoWrite(t *testing.T) {
	t.Parallel()

	// "logs", "test", 30
	logger, err := NewLogger(&Config{
		Dir:    "test_logs",
		Prefix: "test",
		Rotate: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	cases := []struct {
		name    string
		message string
	}{
		{
			name:    "test_info",
			message: "hello univers :)",
		},
	}

	for _, c := range cases {
		t.Run("info", func(t *testing.T) {
			logger.Info(c.message, NewField("test", "test"))
		})
	}
}
func TestLogErrorWrite(t *testing.T) {
	t.Parallel()

	// "logs", "test", 30
	logger, err := NewLogger(&Config{
		Dir:    "test_logs",
		Prefix: "test",
		Rotate: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	cases := []struct {
		name    string
		message string
	}{
		{
			name:    "test_error",
			message: "something went wrong :(",
		},
	}

	for _, c := range cases {
		t.Run("error", func(t *testing.T) {
			logger.Info(c.message, NewField("test", "test"))
		})
	}
}
func TestLogFatalWrite(t *testing.T) {
	t.Parallel()

	// "logs", "test", 30
	logger, err := NewLogger(&Config{
		Dir:    "test_logs",
		Prefix: "test",
		Rotate: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	cases := []struct {
		name    string
		message string
	}{
		{
			name:    "test_fatal",
			message: "something seriously went wrong :O",
		},
	}

	for _, c := range cases {
		t.Run("fatal", func(t *testing.T) {
			logger.Info(c.message, NewField("test", "test"))
		})
	}
}

func BenchmarkLogWrite(b *testing.B) {
	// "logs", "test", 30
	logger, err := NewLogger(&Config{
		Dir:    "test_logs",
		Prefix: "test",
		Rotate: 30,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	n := 0
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark", NewField("number", strconv.Itoa(n)))
		n++
	}
}
func BenchmarkParallelLogWrite(b *testing.B) {
	// "logs", "test", 30
	logger, err := NewLogger(&Config{
		Dir:    "test_logs",
		Prefix: "test",
		Rotate: 30,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.RunParallel(func(p *testing.PB) {
		n := 0
		for p.Next() {
			logger.Info("benchmark", NewField("number", strconv.Itoa(n)))
			n++
		}
	})
}
