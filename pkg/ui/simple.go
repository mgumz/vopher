package ui

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mgumz/vopher/pkg/vopher"
)

// UISimple is an UI which prints the list of fetched plugins onto the
// terminal
type Simple struct {
	jobs    map[string]*vopher.Runtime
	origins map[string]vopher.Origin
	vopher.Runtime
	sync.WaitGroup
	sync.Mutex
	hasStarted bool
}

func (simple *Simple) Start() {
	if simple.hasStarted {
		return
	}
	fmt.Println("started")
	simple.Runtime.Start()
	simple.hasStarted = true
}

func (simple *Simple) Stop() {
	simple.Runtime.Stop()
	var d time.Duration
	for _, rt := range simple.jobs {
		d += rt.Duration()
	}

	dt := simple.Duration()

	simple.Lock()
	fmt.Println("done", strconv.FormatFloat(dt.Seconds(), 'f', 2, 64)+"s",
		"(", len(simple.jobs), "jobs, cumulated runtime ",
		strconv.FormatFloat(d.Seconds(), 'f', 2, 64),
		")")
	simple.Unlock()
}

func (simple *Simple) AddJob(id string) {
	simple.jobs[id] = &vopher.Runtime{}
	simple.jobs[id].Start()
	simple.Add(1)
}

func (simple *Simple) JobDone(id string) {
	simple.Done()
	rt := simple.jobs[id]
	rt.Stop()
	simple.jobs[id] = rt
	simple.Lock()
	origin := ""
	if o, ok := simple.origins[id]; ok {
		origin = "  " + o.String()
	}
	fmt.Printf(" job %.2fs %s%s\n", simple.jobs[id].Duration().Seconds(), id, origin)
	simple.Unlock()
}

func (simple *Simple) Print(id, msg string) {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	scanner.Split(bufio.ScanLines)
	simple.Lock()
	for scanner.Scan() {
		fmt.Println(id, scanner.Text())
	}
	simple.Unlock()
}

func (simple *Simple) PrintShort(id, msg string) {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	scanner.Split(bufio.ScanLines)
	simple.Lock()
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println(id, text)
		if strings.HasPrefix(strings.TrimSpace(text), "*") {
			break
		}
	}
	simple.Unlock()
}

func (simple *Simple) Wait()    { simple.WaitGroup.Wait() }
func (simple *Simple) Refresh() {}

func (simple *Simple) JobOrigin(id string, origin vopher.Origin) {
	simple.Lock()
	if simple.origins == nil {
		simple.origins = make(map[string]vopher.Origin)
	}
	simple.origins[id] = origin
	simple.Unlock()
}
