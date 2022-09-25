package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/m-lab/ndt7-client-go/internal/emitter"
)

func TestCircularQueue(t * testing.T) {
	fakeSummary := func(fqdn string) emitter.Summary {
		return emitter.Summary{
			ServerFQDN: fqdn,
		}
	}

	tests := []struct{
		size int
		input []emitter.Summary
		want []string
	}{
		{
			2,
			[]emitter.Summary{},
			[]string{},
		},
		{
			2,
			[]emitter.Summary{fakeSummary("zero")},
			[]string{"0 zero"},
		},
		{
			2,
			[]emitter.Summary{fakeSummary("zero"), fakeSummary("one")},
			[]string{"1 one", "0 zero"},
		},
		{
			2,
			[]emitter.Summary{fakeSummary("zero"), fakeSummary("one"), fakeSummary("two")},
			[]string{"2 two", "1 one"},
		},
	}

	for _, tc := range(tests) {
		q := newCircularQueue(tc.size)
		for i, s := range(tc.input) {
			q.push(entry{time.Unix(int64(i), 0), s})
		}

		got := make([]string, 0)
		q.forEachReversed(func(e entry) {
			got = append(got, fmt.Sprintf("%d %s", e.completion.Unix(), e.summary.ServerFQDN))
		})

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("want %v; got %v", tc.want, got)
		}
	}
}
