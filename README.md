# flog
a simple logger API for Go program that save logs into a file.

**NOTE**: This package is provided "as is" with no guarantee. Use it at your own risk and always test it yourself before using it in a production environment. If you find any issues, please [create a new issue](https://github.com/twiny/flog/issues/new).

## Install
`go get github.com/twiny/flog`

## API
```go
Info(message string, props map[string]string)
Error(message string, props map[string]string)
Fatal(message string, props map[string]string)
```

## Usage

```go
package main

import "github.com/twiny/flog"

func main() {
	logger, err := flog.NewLogger("./tmp/logs/", "test")
	if err != nil {
		// handler error
		return
	}
	defer logger.Close()

	logger.Info("info", map[string]string{
		// add other info
	})

	logger.Error("error", map[string]string{
		// add other info
	})

	logger.Fatal("fatal", map[string]string{
		// add other info
	})
}
```

## Benchmark
```
goos: darwin
goarch: amd64
pkg: github.com/twiny/flog
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkLogWrite-4   	  185366	      6995 ns/op	    1120 B/op	      12 allocs/op
PASS
ok  	github.com/twiny/flog	1.719s
```

test
go test -timeout 30s -run ^TestLogInfoWrite$ github.com/twiny/flog
go test -timeout 30s -run ^TestLogErrorWrite$ github.com/twiny/flog
go test -timeout 30s -run ^TestLogFatalWrite$ github.com/twiny/flog

bench
go test -benchmem -run=^$ -bench ^BenchmarkLogWrite$ github.com/twiny/flog