package file

import (
	"os"
	"time"
)

// File
type File struct {
	dir    string
	name   string
	prefix string
	file   *os.File
}

// NewFile
func NewFile() (*File, error) {
	return &File{
		dir:  dir,
		name: name,
	}
}

// Write
func (f *File) Write(b []byte) error {
	if f.File == nil {
		var err error
		f.File, err = os.OpenFile(f.Dir+"/"+f.Name, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
	}
	return f.File.Write(b)
}

// Rotate
func (f *File) Rotate(d time.Duration) error {
	if f.File != nil {
		f.File.Close()
		f.File = nil
	}
	return nil
}

// Close
func (f *File) Close() error {
	if f.File != nil {
		return f.File.Close()
	}
	return nil
}
