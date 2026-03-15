package ratelimit

import (
	"context"
	"sync/atomic"
	"time"
)

type TPSController struct {
	tps atomic.Int64
}

func NewTPSController(initial int64) *TPSController {
	c := &TPSController{}
	c.tps.Store(initial)
	return c
}

func (c *TPSController) Set(tps int64) {
	c.tps.Store(tps)
}

func (c *TPSController) Run(ctx context.Context, permits chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		tps := c.tps.Load()
		if tps <= 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		interval := time.Second / time.Duration(tps)
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			select {
			case permits <- struct{}{}:
			default:
			}
		}
	}
}
