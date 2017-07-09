package main

import (
	"sync"
	"time"
)

type UiQuiet struct {
	sync.WaitGroup
	runtime _ri
}

func (ui *UiQuiet) Start()               { ui.runtime.start = time.Now() }
func (ui *UiQuiet) Stop()                { ui.runtime.end = time.Now() }
func (ui *UiQuiet) AddJob(id string)     { ui.WaitGroup.Add(1) }
func (ui *UiQuiet) JobDone(id string)    { ui.WaitGroup.Done() }
func (ui *UiQuiet) Print(id, msg string) {}
func (ui *UiQuiet) Wait()                { ui.WaitGroup.Wait() }
func (ui *UiQuiet) Refresh()             {}
