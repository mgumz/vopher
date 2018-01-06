package main

import "time"

type Runtime struct {
	start time.Time
	end   time.Time
}

func (rt Runtime) duration() time.Duration {
	return rt.end.Sub(rt.start)
}
