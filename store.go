package flog

// Store
type Store interface {
	Write([]byte) error
}
