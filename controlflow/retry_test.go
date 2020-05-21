package controlflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	policies := map[string]BackoffPolicy{
		"exponential": ExponentialBackoff(time.Millisecond, 20*time.Millisecond),
		"random":      RandomBackoff(10*time.Millisecond, 20*time.Millisecond),
		"static":      StaticBackoff(time.Millisecond),
		"no wait":     NoWait(),
	}

	attempts := 5

	for name, policy := range policies {
		t.Run(name, func(t *testing.T) {
			run := 0
			lastRun := time.Now()

			e := Retry(context.Background(), attempts, policy, func() error {
				run++
				now := time.Now()
				wait := now.Sub(lastRun).Milliseconds()
				lastRun = now
				e := fmt.Errorf("[%s][%d/%d]: run after: %d", name, run, attempts, wait)
				fmt.Println(e.Error())
				if run >= attempts {
					return nil
				}
				return e
			})

			assert.NoError(t, e)
		})
	}
}
