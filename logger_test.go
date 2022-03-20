package flog

import (
	"strconv"
	"testing"
)

func TestLogInfoWrite(t *testing.T) {
	t.Parallel()

	logger, err := NewLogger("logs", "test")
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
		t.Run("", func(t *testing.T) {
			logger.Info(c.message, map[string]string{})
		})
	}
}
func TestLogErrorWrite(t *testing.T) {
	t.Parallel()

	logger, err := NewLogger("logs", "test")
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
		t.Run("", func(t *testing.T) {
			logger.Error(c.message, map[string]string{})
		})
	}
}
func TestLogFetalWrite(t *testing.T) {
	t.Parallel()

	logger, err := NewLogger("logs", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	cases := []struct {
		name    string
		message string
	}{
		{
			name:    "test_fetal",
			message: "something seriously went wrong :O",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			logger.Fetal(c.message, map[string]string{})
		})
	}
}

func BenchmarkLogWrite(b *testing.B) {
	logger, err := NewLogger("logs", "test")
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("benchmark", map[string]string{
			"number": strconv.Itoa(i),
		})
	}
}
func BenchmarkParallelLogWrite(b *testing.B) {
	logger, err := NewLogger("logs", "test")
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.RunParallel(func(p *testing.PB) {
		n := 0
		for p.Next() {
			logger.Info("benchmark", map[string]string{
				"number": strconv.Itoa(n),
			})
			n++
		}
	})
}
