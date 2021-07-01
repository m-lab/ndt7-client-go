package emitter

import (
	"testing"
)

func TestNewSummary(t *testing.T) {
	s := NewSummary("test")
	if s == nil {
		t.Fatal("NewSummary() did not return a Summary")
	}
	if s.ServerFQDN != "test" {
		t.Fatal("NewSummary(): unexpected Server field")
	}
}
