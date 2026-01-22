package railzwayclient

import (
	"context"
	"time"
)

type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
}

func (r RetryPolicy) Do(ctx context.Context, safe bool, fn func() error) error {
	var err error
	for i := 0; i <= r.MaxRetries; i++ {
		err = fn()
		if err == nil || !safe {
			return err
		}
		time.Sleep(r.BaseDelay * time.Duration(i+1))
	}
	return err
}
