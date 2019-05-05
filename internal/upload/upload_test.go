package upload

import (
	"context"
	"testing"
	"time"

	"github.com/m-lab/ndt7-client-go/internal/mockable"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestNormal is the normal test case
func TestNormal(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mockable.Conn{}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}
