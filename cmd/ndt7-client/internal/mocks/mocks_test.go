package mocks

import "testing"

// TestFailingWriter verifies that the FailingWriter always fails.
func TestFailingWriter(t *testing.T) {
	wr := FailingWriter{}
	n, err := wr.Write([]byte("abc"))
	if n != 0 {
		t.Fatal("Expected zero bytes here")
	}
	if err != ErrMocked {
		t.Fatal("Expected an ErrMocked here")
	}
}
