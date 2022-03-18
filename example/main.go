package main

import (
	"github.com/twiny/flog"
	"github.com/twiny/flog/stores/file"
)

func main() {
	file, err := file.NewFile()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	log, err := flog.NewLogger(file)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	log.Info("Hello world", make(map[string]string))
}
