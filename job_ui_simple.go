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
	runtime map[string]_ri
}

func (ui *UiSimple) Start() {
	fmt.Println("started")
}

func (ui *UiSimple) Stop() {
	var d time.Duration
	for id := range ui.runtime {
		rt := ui.runtime[id]
		d += rt.end.Sub(rt.start)
	}
	fmt.Println("finish (", len(ui.runtime), "jobs, runtime ", d, ")")
}

func (ui *UiSimple) AddJob(id string) {
	ui.runtime[id] = _ri{start: time.Now()}
	ui.WaitGroup.Add(1)
}

func (ui *UiSimple) JobDone(id string) {
	ui.WaitGroup.Done()
	rt := ui.runtime[id]
	rt.end = time.Now()
	ui.runtime[id] = rt
	fmt.Println("done", id, ui.runtime[id].end.Sub(ui.runtime[id].start))
}

func (ui *UiSimple) Wait()    { ui.WaitGroup.Wait() }
func (ui *UiSimple) Refresh() {}
