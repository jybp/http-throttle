package throttle

import (
	"context"
	"errors"
	"time"
)

// A Quota controls how much events can happen within a timeframe.
// It is useful to enforce long-term rate limits where failing is more appropriate than blocking.
// The timeframe starts when Wait is called for the first time.
type Quota struct {
	Interval time.Duration
	Limit    int

	c    int
	from time.Time
}

// NewQuota returns a new Quota that allows n events to occur wihin duration d.
func NewQuota(d time.Duration, n int) *Quota {
	return &Quota{
		Interval: d,
		Limit:    n,
	}
}

var ErrQuotaExceeded = errors.New("quota exceeded")

// Wait does not block but returns ErrQuotaExceeded when the Limit is reached.
func (q *Quota) Wait(ctx context.Context) error {
	now := time.Now().UTC()
	if q.from == (time.Time{}) {
		q.from = now
	}
	if now.After(q.from.Add(q.Interval)) {
		q.from = now
		q.c = 0
	}
	q.c++
	if q.c > q.Limit {
		return ErrQuotaExceeded
	}
	return nil
}
