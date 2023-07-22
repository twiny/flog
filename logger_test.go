package flog

import (
	"sync"
	"testing"
)

func TestLogWrite(t *testing.T) {
	logger, err := NewLogger("./logs/test.log", 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	logger.Info("hello univers :)", NewField("test", "test"))
	logger.Error("something went wrong :(", NewField("test", "test"))
}
func TestLogRotate(t *testing.T) {
	logger, err := NewLogger("./logs/test.log", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	for i := 0; i < 1024; i++ {
		logger.Info("hello univers :)", NewField("test", "test"))
	}
}
func BenchmarkLogWrite(b *testing.B) {
	logger, err := NewLogger("./logs/test.log", 30, 10)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	var x = struct {
		Name string
		Age  int
		Jobs map[string]any
	}{
		Name: "John Doe",
		Age:  22,
		Jobs: map[string]any{
			"digital marketer": 2019,
			"golang developer": "github",
		},
	}

	f := []Field{
		NewField("hello", "world"),
		NewField("nice", 123),
		NewField("person", x),
	}

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			logger.Info("from_main", f...)
			wg.Done()
		}()
	}
	wg.Wait()
}
