package flog

import (
	"strconv"
	"testing"
)

func TestLogWrite(t *testing.T) {
	// todo: implement
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

	b.RunParallel(func(p *testing.PB) {
		defer logger.Close()

		n := 0
		for p.Next() {
			logger.Info("benchmark", map[string]string{
				"number": strconv.Itoa(n),
			})
			n++
		}
	})
}
