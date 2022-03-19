package main

import (
	"fmt"
	"path"
	"strings"
	"time"
)

const (
	DateFormat = "02-01-2006"            // dd-mm-yyyy
	TimeFormat = "02-Jan-2006 15h04m05s" // dd-mm-yyyy hhmmss
)

func main() {
	p, t := GenerateNext("/tmp", "test")
	fmt.Println(p, t)
}

// GenerateNext
func GenerateNext(dir, prefix string) (string, time.Duration) {
	now := time.Now()
	filename := strings.Join([]string{
		prefix,
		now.Format(DateFormat),
	}, "_")
	filename = filename + ".log"

	// tomorrow
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	left := time.Until(tomorrow).Round(time.Minute)

	return path.Join(dir, filename), left
}
