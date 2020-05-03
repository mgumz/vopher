package vopher

import "time"

type Runtime struct {
	start time.Time
	end   time.Time
}

func (rt Runtime) Duration() time.Duration {
	return rt.end.Sub(rt.start)
}

func (rt Runtime) Start() { rt.start = time.Now() }
func (rt Runtime) Stop()  { rt.end = time.Now() }
