package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type ProgressTicker struct {
	WriteMeter
	ticker *time.Ticker
	stop   chan bool
}

func NewProgressTicker(max int64) *ProgressTicker {
	return &ProgressTicker{WriteMeter: WriteMeter{Max: max}}
}

func (pt *ProgressTicker) Start(prefix string, dur time.Duration) {
	pt.stop = make(chan bool)
	pt.ticker = time.NewTicker(dur)

	go func() {
		for {
			select {
			case <-pt.ticker.C:
				pt.Print(prefix)
			case <-pt.stop:
				pt.ticker.Stop()
				pt.Print(prefix)
				fmt.Println()
				fmt.Println()
				return
			}
		}
	}()
}

func (pt *ProgressTicker) Stop() {
	pt.stop <- true
}

func (pt *ProgressTicker) Print(prefix string) {
	if pt.Max == 0 {
		return
	}

	ticks := strings.Repeat("=====", 10)
	n := math.Max(1.0, math.Floor(float64(len(ticks))*pt.Progress()))
	fmt.Printf("\r%s: (%d/%d) %s|",
		prefix, pt.WriteCounter, pt.Max, ticks[:int(n)-1])
}

func (pt *ProgressTicker) MaxOut() {
	pt.WriteCounter = WriteCounter(pt.Max)
}

// ========================================================================

type WriteMeter struct {
	WriteCounter
	Max int64
}

func (meter *WriteMeter) Progress() float64 {
	return float64(meter.WriteCounter) / float64(meter.Max)
}

func (meter *WriteMeter) String() string {
	return strconv.FormatFloat(meter.Progress(), 'f', 2, 64)
}

// ========================================================================

type WriteCounter int64

func (counter *WriteCounter) Write(data []byte) (n int, _ error) {
	n = len(data)
	*counter += WriteCounter(n)
	return
}
