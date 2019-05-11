// Package mocks contains mocks
package mocks

import "errors"

// ErrMocked is a mocked error
var ErrMocked = errors.New("mocked error")

// FailingWriter is a writer that always fails
type FailingWriter struct{}

// Write always returns a mocked error
func (FailingWriter) Write([]byte) (int, error) {
	return 0, ErrMocked
}
