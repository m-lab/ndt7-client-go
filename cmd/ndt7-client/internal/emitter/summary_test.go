package emitter

import (
	"testing"
)

func TestNewSummary(t *testing.T) {
	s := NewSummary("test")
	if s == nil {
		t.Fatal("NewSummary() did not return a Summary")
	}
	if s.Server != "test" {
		t.Fatal("NewSummary(): unexpected Server field")
	}
}
