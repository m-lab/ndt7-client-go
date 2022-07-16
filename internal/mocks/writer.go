// Package mocks contains mocks
package mocks

import (
	"errors"
)

// ErrMocked is a mocked error
var ErrMocked = errors.New("mocked error")

// FailingWriter is a writer that always fails
type FailingWriter struct{}

// Write always returns a mocked error
func (FailingWriter) Write([]byte) (int, error) {
	return 0, ErrMocked
}

// SavingWriter is a writer that saves what it's passed
type SavingWriter struct {
	Data [][]byte
}

// Write appends data to sw.Data. It never fails.
func (sw *SavingWriter) Write(data []byte) (int, error) {
	sw.Data = append(sw.Data, data)
	return len(data), nil
}
