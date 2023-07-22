# flog
a simple logger API for Go program that save logs into a file.

**NOTE**: This package is provided "as is" with no guarantee. Use it at your own risk and always test it yourself before using it in a production environment. If you find any issues, please [create a new issue](https://github.com/twiny/flog/issues/new).

## Install
`go get github.com/twiny/flog/v2`

## API
```go
Info(msg string, fields ...Field)
Error(msg string, fields ...Field)
Debug(m string, fields ...Field)
Fatal(msg string, fields ...Field)
```

## Usage

```go
package main

import (
	"github.com/twiny/flog/v2"
)

func main() {
	path := "./logs/test.log"
	maxAge := 30  // days
	maxSize := 10 // mb

	logger, err := flog.NewLogger(path, maxAge, maxSize)
	if err != nil {
		panic(err)
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
goos: linux
goarch: amd64
pkg: github.com/twiny/flog/v2
cpu: AMD Ryzen 5 PRO 5650U with Radeon Graphics     
BenchmarkLogWrite
BenchmarkLogWrite-6       275431              6931 ns/op            2406 B/op         29 allocs/op
PASS
ok      github.com/twiny/flog/v2        1.966s
```
