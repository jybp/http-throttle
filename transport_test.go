package throttle_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	throttle "github.com/jybp/http-throttle"
	"golang.org/x/time/rate"
)

func TestTransport(t *testing.T) {
	l := rate.NewLimiter(3, 3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	client := &http.Client{Transport: &throttle.Transport{Limiter: rate.NewLimiter(2, 2)}}
	assertStatusFn := func(expected int) {
		resp, err := client.Get(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != expected {
			t.Fatalf("expected:%d\tactual:%d\n", expected, resp.StatusCode)
		}
	}
	assertStatusFn(http.StatusOK)
	assertStatusFn(http.StatusOK)
	assertStatusFn(http.StatusOK)
	assertStatusFn(http.StatusOK)
}

func TestMultiLimiter(t *testing.T) {
	l := throttle.MultiLimiters(
		throttle.NewQuota(time.Second*2, 101),
		rate.NewLimiter(rate.Every(time.Second/100), 1),
	)
	start := time.Now()
	for i := 0; i < 101; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Fatalf("100 wait took %v", elapsed)
	}
	if err := l.Wait(context.Background()); err != throttle.ErrQuotaExceeded {
		t.Fatal(err)
	}
}

func Example() {
	client := &http.Client{
		Transport: throttle.Default(
			// Returns ErrQuotaExceeded if more than 36000 requests occured within an hour.
			throttle.NewQuota(time.Hour, 36000),
			// Blocks to never exceed 99 requests per second.
			rate.NewLimiter(99, 1),
		),
	}
	resp, err := client.Get("https://golang.org/")
	if err == throttle.ErrQuotaExceeded {
		// Handle err.
	}
	if err != nil {
		// Handle err.
	}
	_ = resp // Do something with resp.
}
