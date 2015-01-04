package main

import (
	"sync"
	"time"
)

type UiOneLine struct {
	sync.WaitGroup
	*ProgressTicker
	prefix   string
	duration time.Duration
}

func (ui *UiOneLine) Start()               { ui.ProgressTicker.Start(ui.prefix, ui.duration) }
func (ui *UiOneLine) Stop()                { ui.ProgressTicker.Stop() }
func (ui *UiOneLine) AddJob(string)        { ui.WaitGroup.Add(1); ui.ProgressTicker.WriteMeter.Max++ }
func (ui *UiOneLine) JobDone(string)       { ui.ProgressTicker.WriteCounter += 1; ui.WaitGroup.Done() }
func (ui *UiOneLine) Wait()                { ui.WaitGroup.Wait(); ui.ProgressTicker.MaxOut(); ui.Refresh() }
func (ui *UiOneLine) Refresh()             { ui.ProgressTicker.Print(ui.prefix) }
func (ui *UiOneLine) Print(string, string) {}
