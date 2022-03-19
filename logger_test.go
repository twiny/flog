package flog

import "testing"

func TestLogWrite(t *testing.T) {

}

func BenchmarkLogWrite(b *testing.B) {
	// logger, err := NewLogger("logs")
	// if err != nil {
	// 	b.Fatal(err)
	// }
	// defer logger.Close()

	// for n := 0; n < b.N; n++ {
	// 	logger.Info("benchmark", map[string]string{
	// 		"head": "202",
	// 		"body": "222",
	// 		"time": "now",
	// 	})
	// }
}
