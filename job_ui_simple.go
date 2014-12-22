package main

import (
	"fmt"
	"sync"
	"time"
)

type _ri struct {
	start time.Time
	end   time.Time
}

type UiSimple struct {
	sync.WaitGroup
	runtime _ri
	jobs    map[string]_ri
}

func (ui *UiSimple) Start() {
	fmt.Println("started")
	ui.runtime.start = time.Now()
}

func (ui *UiSimple) Stop() {
	ui.runtime.end = time.Now()
	var d time.Duration
	for id := range ui.jobs {
		rt := ui.jobs[id]
		d += rt.end.Sub(rt.start)
	}
	fmt.Println("finish (", ui.runtime.end.Sub(ui.runtime.start), "for", len(ui.jobs), "jobs, cumulated runtime ", d, ")")
}

func (ui *UiSimple) AddJob(id string) {
	ui.jobs[id] = _ri{start: time.Now()}
	ui.WaitGroup.Add(1)
}

func (ui *UiSimple) JobDone(id string) {
	ui.WaitGroup.Done()
	rt := ui.jobs[id]
	rt.end = time.Now()
	ui.jobs[id] = rt
	fmt.Println("done", id, ui.jobs[id].end.Sub(ui.jobs[id].start))
}

func (ui *UiSimple) Wait()    { ui.WaitGroup.Wait() }
func (ui *UiSimple) Refresh() {}
