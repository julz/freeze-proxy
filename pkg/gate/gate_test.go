package gate_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/julz/freeze-proxy/pkg/gate"
	"go.uber.org/atomic"
)

func TestGate(t *testing.T) {
	tests := []struct {
		name            string
		pauses, resumes int64
		events          map[time.Duration]time.Duration // start time => req length
	}{{
		name:    "single request",
		pauses:  1,
		resumes: 1,
		events: map[time.Duration]time.Duration{
			1 * time.Second: 2 * time.Second,
		},
	}, {
		name:    "overlapping requests",
		pauses:  1,
		resumes: 1,
		events: map[time.Duration]time.Duration{
			25 * time.Millisecond: 100 * time.Millisecond,
			75 * time.Millisecond: 200 * time.Millisecond,
		},
	}, {
		name:    "subsumbed request",
		pauses:  1,
		resumes: 1,
		events: map[time.Duration]time.Duration{
			25 * time.Millisecond: 300 * time.Millisecond,
			75 * time.Millisecond: 200 * time.Millisecond,
		},
	}, {
		name:    "start stop start",
		pauses:  2,
		resumes: 2,
		events: map[time.Duration]time.Duration{
			25 * time.Millisecond:  300 * time.Millisecond,
			75 * time.Millisecond:  200 * time.Millisecond,
			850 * time.Millisecond: 300 * time.Millisecond,
			900 * time.Millisecond: 400 * time.Millisecond,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paused := atomic.NewInt64(0)
			pause := func() {
				paused.Inc()
			}

			resumed := atomic.NewInt64(0)
			resume := func() {
				resumed.Inc()
			}

			delegated := atomic.NewInt64(0)
			delegate := func(w http.ResponseWriter, r *http.Request) {
				wait, err := strconv.Atoi(r.Header.Get("wait"))
				if err != nil {
					panic(err)
				}

				time.Sleep(time.Duration(wait))
				delegated.Inc()
			}

			h := gate.New(http.HandlerFunc(delegate), pause, resume)

			var wg sync.WaitGroup
			wg.Add(len(tt.events))
			for delay, length := range tt.events {
				time.AfterFunc(delay, func() {
					w := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://target", nil)
					r.Header.Set("wait", strconv.FormatInt(int64(length), 10))
					h.ServeHTTP(w, r)
					wg.Done()
				})
			}

			wg.Wait()

			if got, want := paused.Load(), tt.pauses; got != want {
				t.Errorf("expected to be paused %d times, but was paused %d times", want, got)
			}

			if got, want := delegated.Load(), int64(len(tt.events)); got != want {
				t.Errorf("expected to be delegated %d times, but delegated %d times", want, got)
			}

			if got, want := resumed.Load(), tt.resumes; got != want {
				t.Errorf("expected to be resumed %d times, but was resumed %d times", want, got)
			}
		})
	}
}
