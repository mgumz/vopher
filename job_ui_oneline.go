package main

import (
	"sync"
	"time"
)

// UIOneLine implements a UI which draws the progress in "one line"
// onto the terminal.
type UIOneLine struct {
	sync.WaitGroup
	pt       *ProgressTicker
	prefix   string
	duration time.Duration
}

func (ui *UIOneLine) Start()               { ui.pt.start(ui.prefix, ui.duration) }
func (ui *UIOneLine) Stop()                { ui.pt.stop() }
func (ui *UIOneLine) AddJob(string)        { ui.WaitGroup.Add(1); ui.pt.max++ }
func (ui *UIOneLine) JobDone(string)       { ui.pt.counter += 1; ui.WaitGroup.Done() }
func (ui *UIOneLine) Wait()                { ui.WaitGroup.Wait(); ui.pt.maxOut(); ui.Refresh() }
func (ui *UIOneLine) Refresh()             { ui.pt.print(ui.prefix) }
func (ui *UIOneLine) Print(string, string) {}
