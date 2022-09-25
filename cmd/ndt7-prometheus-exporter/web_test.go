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

	cases := []struct{
		size int
		input []string
		want []string
	}{
		{
			size: 2,
			input: []string{},
			want: []string{},
		},
		{
			size: 2,
			input: []string{"zero"},
			want: []string{"0 zero"},
		},
		{
			size: 2,
			input: []string{"zero", "one"},
			want: []string{"1 one", "0 zero"},
		},
		{
			size: 2,
			input: []string{"zero", "one", "two"},
			want: []string{"2 two", "1 one"},
		},
	}

	for _, tc := range(cases) {
		q := newCircularQueue(tc.size)
		for i, s := range(tc.input) {
			q.push(entry{time.Unix(int64(i), 0), fakeSummary(s)})
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
