package pressure

import (
	"fmt"
	"math"
	"sync"
)

type Controller struct {
	mu      sync.Mutex
	current float64
	enabled bool
	stopped bool
	history []float64
}

func NewController() *Controller { return &Controller{history: make([]float64, 0)} }
func (c *Controller) Authorize() { c.mu.Lock(); c.enabled = true; c.stopped = false; c.mu.Unlock() }
func (c *Controller) Apply(runID string, value float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled || c.stopped {
		return fmt.Errorf("pressure not authorized")
	}
	if runID == "" {
		return fmt.Errorf("missing run")
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("pressure value must be finite")
	}
	c.current = value
	c.history = append(c.history, value)
	return nil
}
func (c *Controller) Stop() { c.mu.Lock(); c.stopped = true; c.enabled = false; c.mu.Unlock() }
