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

// go test -count=1 -v -timeout 30s -run ^TestLogRotate$ github.com/twiny/flog -tags=integration,unit
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

// go test -benchmem -v -count=1 -run=^$ -bench ^BenchmarkLogWrite$ github.com/twiny/flog -tags=integration,unit
func BenchmarkLogWrite(b *testing.B) {
	logger, err := NewLogger("./logs/test.log", 30, 10)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			logger.Info("hello univers :)", NewField("test", "test"))
			wg.Done()
		}()
	}
	wg.Wait()
}
