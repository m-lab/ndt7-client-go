package mocks

import (
	"reflect"
	"testing"
)

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

func TestSavingWriter(t *testing.T) {
	sw := &SavingWriter{}
	first := []byte("abc")
	n, err := sw.Write(first)
	if n != len(first) {
		t.Fatal("Unexpected length")
	}
	if err != nil {
		t.Fatal(err)
	}
	second := []byte("de")
	n, err = sw.Write(second)
	if n != len(second) {
		t.Fatal("Unexpected length")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 2 {
		t.Fatal("Unexpected data length")
	}
	if !reflect.DeepEqual(sw.Data[0], first) {
		t.Fatal("First write is not equal")
	}
	if !reflect.DeepEqual(sw.Data[1], second) {
		t.Fatal("Second write is not equal")
	}
}
