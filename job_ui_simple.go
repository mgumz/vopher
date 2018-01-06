package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UISimple is an UI which prints the list of fetched plugins onto the
// terminal
type UISimple struct {
	jobs map[string]Runtime
	Runtime
	sync.WaitGroup
	sync.Mutex
}

func (ui *UISimple) Start() {
	fmt.Println("started")
	ui.Runtime.start = time.Now()
}

func (ui *UISimple) Stop() {
	ui.Runtime.end = time.Now()
	var d time.Duration
	for _, rt := range ui.jobs {
		d += rt.duration()
	}

	dt := ui.duration()

	ui.Lock()
	fmt.Println("done", strconv.FormatFloat(dt.Seconds(), 'f', 2, 64)+"s",
		"(", len(ui.jobs), "jobs, cumulated runtime ",
		strconv.FormatFloat(d.Seconds(), 'f', 2, 64),
		")")
	ui.Unlock()
}

func (ui *UISimple) AddJob(id string) {
	ui.jobs[id] = Runtime{start: time.Now()}
	ui.WaitGroup.Add(1)
}

func (ui *UISimple) JobDone(id string) {
	ui.WaitGroup.Done()
	rt := ui.jobs[id]
	rt.end = time.Now()
	ui.jobs[id] = rt
	ui.Lock()
	fmt.Printf(" job %.2fs %s\n", ui.jobs[id].duration().Seconds(), id)
	ui.Unlock()
}

func (ui *UISimple) Print(id, msg string) {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	scanner.Split(bufio.ScanLines)
	ui.Lock()
	for scanner.Scan() {
		fmt.Println(id, scanner.Text())
	}
	ui.Unlock()
}

func (ui *UISimple) Wait()    { ui.WaitGroup.Wait() }
func (ui *UISimple) Refresh() {}
