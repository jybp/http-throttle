// Package throttle provides a http.RoundTripper to throttle http requests.
package throttle

import (
	"context"
	"net/http"
	"sync"
)

// RateLimiter interface compatible with golang.org/x/time/rate.
type RateLimiter interface {
	Wait(context.Context) error
}

// Transport implements http.RoundTripper.
type Transport struct {
	Transport http.RoundTripper // Used to make actual requests.
	Limiter   RateLimiter
}

// Default returns a RoundTripper capable of rate limiting http requests.
func Default(r ...RateLimiter) *Transport {
	return Custom(http.DefaultTransport, r...)
}

// Custom uses t to make actual requests.
func Custom(t http.RoundTripper, r ...RateLimiter) *Transport {
	return &Transport{Transport: t, Limiter: MultiLimiters(r...)}
}

// RoundTrip ensures requests are performed within the rate limiting constraints.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := t.Limiter.Wait(r.Context()); err != nil {
		return nil, err
	}
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}
	return t.Transport.RoundTrip(r)
}

// MultiLimiter allows to enforce multiple RateLimiter.
type MultiLimiter struct {
	limiters []RateLimiter
}

// Wait invoke the Wait method of all Limiters concurrently.
func (l *MultiLimiter) Wait(ctx context.Context) (err error) {
	wg := sync.WaitGroup{}
	wg.Add(len(l.limiters))
	for _, l := range l.limiters {
		go func(l RateLimiter) {
			if wErr := l.Wait(ctx); wErr != nil {
				err = wErr
			}
			wg.Done()
		}(l)
	}
	wg.Wait()
	return
}

// MultiLimiters creates a MultiLimiter from limiters.
func MultiLimiters(limiters ...RateLimiter) *MultiLimiter {
	return &MultiLimiter{limiters: limiters}
}
