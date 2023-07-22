package flog

import (
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

func BenchmarkLogWrite(b *testing.B) {}
