package throttle_test

import (
	"context"
	"testing"
	"time"

	throttle "github.com/jybp/http-throttle"
)

func TestQuota(t *testing.T) {
	ctx := context.Background()
	q := throttle.NewQuota(time.Millisecond*10, 2)
	assertFn := func() {
		if err := q.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if err := q.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if err := q.Wait(ctx); err != throttle.ErrQuotaExceeded {
			t.Fatal("ErrQuotaExceeded expected", err)
		}
		if err := q.Wait(ctx); err != throttle.ErrQuotaExceeded {
			t.Fatal("ErrQuotaExceeded expected", err)
		}
	}
	assertFn()
	time.Sleep(time.Millisecond * 10)
	assertFn()
	time.Sleep(time.Millisecond * 20)
	assertFn()
}
