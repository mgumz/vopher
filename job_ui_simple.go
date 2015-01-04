package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
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
	sync.Mutex
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

	dt := ui.runtime.end.Sub(ui.runtime.start)

	ui.Lock()
	fmt.Println("done", strconv.FormatFloat(dt.Seconds(), 'f', 2, 64)+"s",
		"(", len(ui.jobs), "jobs, cumulated runtime ",
		strconv.FormatFloat(d.Seconds(), 'f', 2, 64),
		")")
	ui.Unlock()
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
	d := ui.jobs[id].end.Sub(ui.jobs[id].start)
	ui.Lock()
	fmt.Println(" job", strconv.FormatFloat(d.Seconds(), 'f', 2, 64)+"s", id)
	ui.Unlock()
}

func (ui *UiSimple) Print(id, msg string) {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	scanner.Split(bufio.ScanLines)
	ui.Lock()
	for scanner.Scan() {
		fmt.Println(" job", id, scanner.Text())
	}
	ui.Unlock()
}

func (ui *UiSimple) Wait()    { ui.WaitGroup.Wait() }
func (ui *UiSimple) Refresh() {}
