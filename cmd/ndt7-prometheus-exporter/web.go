package main

import (
	"fmt"
	"net/http"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/m-lab/ndt7-client-go/internal/emitter"
	"github.com/m-lab/ndt7-client-go/spec"
)

type entry struct {
	completion time.Time
	summary emitter.Summary
}

type circularQueue struct {
	sync.Mutex

	// Underlying storage for queue elements
	entries []entry

	// Index into current start of the queue
	start int

	// Number of elements in the queue
	count int
}

func newCircularQueue(maxSize int) *circularQueue {
	return &circularQueue{
		entries: make([]entry, maxSize),
		start: 0,
		count: 0,
	}
}

func (q *circularQueue) internalPop() {
	if q.count == 0 {
		return
	}

	// Assumes mutex is locked
	q.start = (q.start + 1) % len(q.entries)
	q.count--
}

func (q *circularQueue) push(s entry) {
	q.Lock()
	defer q.Unlock()

	if q.count >= len(q.entries) {
		q.internalPop()
	}

	i := (q.start + q.count) % len(q.entries)
	q.entries[i] = s
	q.count++
}

func (q *circularQueue) forEachReversed(f func(entry)) {
	var copy []entry
	func() {
		q.Lock()
		defer q.Unlock()

		copy = make([]entry, q.count)
		for i := 0; i < q.count; i++ {
			j := (i + q.start) % len(q.entries)
			copy[q.count - i - 1] = q.entries[j]
		}
	}()

	for _, entry := range copy {
		f(entry)
	}
}

// statusHandler implements both emitter.Emitter and http.Handler interfaces
type statusHandler struct {
	// Chained emitter.Emitter
	emitter emitter.Emitter

	// A cache of recent test results
	results *circularQueue
}

func newStatusHandler(e emitter.Emitter, maxSize int) *statusHandler {
	return &statusHandler{
		emitter: e,
		results: newCircularQueue(maxSize),
	}
}

// OnStarting emits the starting event
func (h *statusHandler) OnStarting(test spec.TestKind) error {
	return h.emitter.OnStarting(test)
}

// OnError emits the error event
func (h *statusHandler) OnError(test spec.TestKind, err error) error {
	return h.emitter.OnError(test, err)
}

// OnConnected emits the connected event
func (h *statusHandler) OnConnected(test spec.TestKind, fqdn string) error {
	return h.emitter.OnConnected(test, fqdn)
}

// OnDownloadEvent handles an event emitted during the download
func (h *statusHandler) OnDownloadEvent(m *spec.Measurement) error {
	return h.emitter.OnDownloadEvent(m)
}

// OnUploadEvent handles an event emitted during the upload
func (h *statusHandler) OnUploadEvent(m *spec.Measurement) error {
	return h.emitter.OnUploadEvent(m)
}

// OnComplete is the event signalling the end of the test
func (h *statusHandler) OnComplete(test spec.TestKind) error {
	return h.emitter.OnComplete(test)
}

// OnSummary handles the summary event, emitted after the test is over.
func (h *statusHandler) OnSummary(s *emitter.Summary) error {
	h.results.push(entry{time.Now(), *s})

	return h.emitter.OnSummary(s)
}

// Part of http.Handler interface
func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `<!DOCTYPE html>
<html>
<head><title>NDT7 Prometheus Exporter</title></head>
<body>
<h1>A non-interactive NDT7 client</h1>
<p><a href="/metrics">Metrics</a></p>
<table border='1'>
`)

	h.results.forEachReversed(func(entry entry) {
		b := &strings.Builder{}
		io.WriteString(b, "<tr>\n")
		fmt.Fprintf(b, "<td>%s</td><td><pre>\n", entry.completion.Format(time.RFC3339))
		e := emitter.NewHumanReadableWithWriter(b)
		e.OnSummary(&entry.summary)
		io.WriteString(b, "\n</pre></td>\n")
		io.WriteString(b, "</tr>\n")
		io.WriteString(w, b.String())
	})

	io.WriteString(w, `</table>
</body>
</html>
`)
}

func (h *statusHandler) handler() http.Handler {
	return h
}
