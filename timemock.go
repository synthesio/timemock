package timemock

import (
	"sync"
	"sync/atomic"
	"time"
)

type timemockClock struct {
	rw         *sync.RWMutex
	frozen     atomic.Bool
	traveled   atomic.Bool
	freezeTime time.Time
	travelTime time.Time
	scale      float64
}

func (c *timemockClock) Scale(scale float64) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.scale = scale
	if !c.traveled.Load() {
		now := time.Now()
		c.freezeTime = now
		c.travelTime = now
		c.traveled.Store(true)
	}
}

func (c *timemockClock) Now() time.Time {
	// fast path
	if !c.frozen.Load() && !c.traveled.Load() {
		return time.Now()
	}

	c.rw.RLock()
	defer c.rw.RUnlock()

	if c.frozen.Load() {
		return c.freezeTime
	}

	if c.traveled.Load() {
		return c.freezeTime.Add(time.Duration(float64(time.Since(c.travelTime)) * c.scale))
	}

	return time.Now()
}

func (c *timemockClock) Freeze(t time.Time) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.freezeTime = t
	c.frozen.Store(true)
}

func (c *timemockClock) Travel(t time.Time) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.freezeTime = t
	c.travelTime = time.Now()
	c.traveled.Store(true)
}

func (c *timemockClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}

func (c *timemockClock) Until(t time.Time) time.Duration {
	return t.Sub(c.Now())
}

func (c *timemockClock) Return() {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.frozen.Store(false)
	c.traveled.Store(false)
	c.scale = 1
}
