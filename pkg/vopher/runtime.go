package vopher

import "time"

// Runtime is a small time helper to collect runtime stats
type Runtime struct {
	start time.Time
	end   time.Time
}

// Duration returns the duration of the run
func (rt Runtime) Duration() time.Duration {
	return rt.end.Sub(rt.start)
}

// Start starts collecting runtime stats
func (rt *Runtime) Start() { rt.start = time.Now() }

// Stop stops collecting runtime stats
func (rt *Runtime) Stop() { rt.end = time.Now() }
