package controlflow

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"
)

// Retry calls the function with given backoff.
func Retry(ctx context.Context, attempts int, policy BackoffPolicy, f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v, %s", r, string(debug.Stack()))
		}
	}()

	w := wait(ctx, policy, attempts)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case _, ok := <-w:
			if !ok {
				return err
			}

			err = f()
			if err == nil {
				return nil
			}
		}
	}
}

// BackoffPolicy returns next wait duration
type BackoffPolicy func(last time.Duration) time.Duration

func wait(ctx context.Context, next BackoffPolicy, attempts int) <-chan struct{} {
	goon := make(chan struct{})

	go func() {
		defer close(goon)

		// do not wait before first run
		goon <- struct{}{}

		dur := time.Duration(0)
		for run := 1; attempts <= 0 || run < attempts; run++ {
			dur = next(dur)
			if dur <= 0 {
				goon <- struct{}{}
				continue
			}

			tic := time.NewTicker(dur)
			select {
			case <-ctx.Done():
				tic.Stop()
				return
			case <-tic.C:
				goon <- struct{}{}
			}
		}
	}()

	return goon
}

func ExponentialBackoff(initial, cap time.Duration) BackoffPolicy {
	return func(last time.Duration) time.Duration {
		if last <= 0 {
			return initial
		}

		next := last * 2
		if cap > 0 && next > cap {
			next = cap
		}
		return next
	}
}

func RandomBackoff(min, max time.Duration) BackoffPolicy {
	return func(time.Duration) time.Duration {
		return min + time.Duration(rand.Int63n(int64(max-min)))
	}
}

func StaticBackoff(interval time.Duration) BackoffPolicy {
	return func(time.Duration) time.Duration {
		return interval
	}
}

func NoWait() BackoffPolicy {
	return func(time.Duration) time.Duration {
		return 0
	}
}
