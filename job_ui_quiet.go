package main

import (
	"sync"
	"time"
)

// UIQuiet is kind of a "no UI" coz it gives no feedback at all
type UIQuiet struct {
	sync.WaitGroup
	Runtime
}

func (ui *UIQuiet) Start()               { ui.Runtime.start = time.Now() }
func (ui *UIQuiet) Stop()                { ui.Runtime.end = time.Now() }
func (ui *UIQuiet) AddJob(id string)     { ui.WaitGroup.Add(1) }
func (ui *UIQuiet) JobDone(id string)    { ui.WaitGroup.Done() }
func (ui *UIQuiet) Print(id, msg string) {}
func (ui *UIQuiet) Wait()                { ui.WaitGroup.Wait() }
func (ui *UIQuiet) Refresh()             {}
