package ui

import (
	"sync"
	"time"
)

// UIOneLine implements a UI which draws the progress in "one line"
// onto the terminal.
type OneLine struct {
	sync.WaitGroup
	pt       *ProgressTicker
	prefix   string
	duration time.Duration
}

func (ol *OneLine) Start()               { ol.pt.start(ol.prefix, ol.duration) }
func (ol *OneLine) Stop()                { ol.pt.stop() }
func (ol *OneLine) AddJob(string)        { ol.WaitGroup.Add(1); ol.pt.max++ }
func (ol *OneLine) JobDone(string)       { ol.pt.counter += 1; ol.WaitGroup.Done() }
func (ol *OneLine) Wait()                { ol.WaitGroup.Wait(); ol.pt.maxOut(); ol.Refresh() }
func (ol *OneLine) Refresh()             { ol.pt.print(ol.prefix) }
func (ol *OneLine) Print(string, string) {}
