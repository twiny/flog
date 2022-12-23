# flog
a simple logger API for Go program that save logs into a file.

**NOTE**: This package is provided "as is" with no guarantee. Use it at your own risk and always test it yourself before using it in a production environment. If you find any issues, please [create a new issue](https://github.com/twiny/flog/issues/new).

## Install
`go get github.com/twiny/flog`

## API
```go
Info(msg string, fields ...Field)
Error(msg string, fields ...Field)
Fatal(msg string, fields ...Field)
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/twiny/flog"
)

// main
func main() {
	// config
	conf := &flog.Config{
		Dir:    "./logs", // log directory
		Prefix: "app",  // prefix
		Rotate: 7, // how many days to store logs
	}

	logger, err := flog.NewLogger(conf)
	if err != nil {
		fmt.Println(err)
		return
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

	f := []flog.Field{
		flog.NewField("hello", "world"),
		flog.NewField("nice", 123),
		flog.NewField("person", x),
	}

	logger.Info("from_main", f...)
}
```

## Benchmark
```
goos: darwin
goarch: amd64
pkg: github.com/twiny/flog
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkLogWrite-4       163430              6599 ns/op             718 B/op         11 allocs/op
PASS
ok      github.com/twiny/flog   1.320s
```

### Run tests
```
test
go test -timeout 30s -run ^TestLogInfoWrite$ github.com/twiny/flog
go test -timeout 30s -run ^TestLogErrorWrite$ github.com/twiny/flog
go test -timeout 30s -run ^TestLogFatalWrite$ github.com/twiny/flog

benchmark
go test -benchmem -run=^$ -bench ^BenchmarkLogWrite$ github.com/twiny/flog
go test -benchmem -run=^$ -bench ^BenchmarkParallelLogWrite$ github.com/twiny/flog
```