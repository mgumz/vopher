package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

const CURSOR_UP = "\x1b[1A"

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

	fmt.Println()
	go func() {
		for {
			select {
			case <-pt.ticker.C:
				pt.Print(prefix)
			case <-pt.stop:
				pt.ticker.Stop()
				pt.Print(prefix)
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

	cols, _, _ := TerminalSize(os.Stdout)
	info := fmt.Sprintf("%s: (%d/%d)", prefix, pt.WriteCounter, pt.Max)
	full := bytes.Repeat([]byte("."), cols-len(info)-2)
	n_ticks := int(math.Max(1.0, math.Floor(float64(len(full))*pt.Progress())))
	i := 0
	for ; i < n_ticks; i++ {
		full[i] = '='
	}
	full[0] = '['
	if i < len(full)-1 {
		full[i] = '>'
	}
	full[len(full)-1] = ']'

	if len(full) > 10 {
		progress := fmt.Sprintf(" %d%% ", int(100.0*pt.Progress()))
		copy(full[(len(full)/2)-(len(progress)/2):], progress)
	}

	// using cursor-up+progress+newline works more stable than to \r
	// the cursor.
	fmt.Println(CURSOR_UP, info, string(full))
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
