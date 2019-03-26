// Package throttle provides a http.RoundTripper to throttle http requests.
package throttle

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"
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
func (l *MultiLimiter) Wait(ctx context.Context) error {
	wg := errgroup.Group{}
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()
	for _, l := range l.limiters {
		wg.Go(func(ctx context.Context, l RateLimiter) func() error {
			return func() error {
				if err := l.Wait(ctx); err != nil {
					cancelFn()
					return err
				}
				return nil
			}
		}(ctx, l))
	}
	return wg.Wait()
}

// MultiLimiters creates a MultiLimiter from limiters.
func MultiLimiters(limiters ...RateLimiter) *MultiLimiter {
	return &MultiLimiter{limiters: limiters}
}
