package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

type ProgressTicker struct {
	WriteMeter
	ticker *time.Ticker
	stopCh chan bool
}

func newProgressTicker(max int64) *ProgressTicker {
	return &ProgressTicker{WriteMeter: WriteMeter{max: max}}
}

func (pt *ProgressTicker) start(prefix string, dur time.Duration) {

	pt.stopCh = make(chan bool)
	pt.ticker = time.NewTicker(dur)

	go func() {
		fmt.Println()
		for {
			select {
			case <-pt.ticker.C:
				pt.print(prefix)
			case <-pt.stopCh:
				pt.ticker.Stop()
				pt.print(prefix)
				fmt.Println()
				return
			}
		}
	}()
}

func (pt *ProgressTicker) stop() {
	pt.stopCh <- true
}

func (pt *ProgressTicker) print(prefix string) {

	if pt.max == 0 {
		return
	}

	cols, _, _ := terminalSize(os.Stdout)

	if cols <= 0 {
		log.Fatal("can't get TerminalSize(), use other -ui type")
		return
	}

	info := fmt.Sprintf("%s: (%d/%d)", prefix, pt.counter, pt.max)
	full := bytes.Repeat([]byte("."), cols-len(info)-2)
	ticks := int(math.Max(1.0, math.Floor(float64(len(full))*pt.progress())))
	i := 0
	for ; i < ticks; i++ {
		full[i] = '='
	}
	full[0] = '['
	if i < len(full)-1 {
		full[i] = '>'
	}
	full[len(full)-1] = ']'

	if len(full) > 10 {
		progress := fmt.Sprintf(" %d%% ", int(100.0*pt.progress()))
		copy(full[(len(full)/2)-(len(progress)/2):], progress)
	}

	// using cursor-up+progress+newline works more stable than to \r
	// the cursor.
	cursorNUp(os.Stdout, 1)
	fmt.Println(info, string(full))
}

func (pt *ProgressTicker) maxOut() {
	pt.counter = WriteCounter(pt.max)
}

// ========================================================================

type WriteMeter struct {
	counter WriteCounter
	max     int64
}

func (meter *WriteMeter) progress() float64 {
	return float64(meter.counter) / float64(meter.max)
}

func (meter *WriteMeter) String() string {
	return strconv.FormatFloat(meter.progress(), 'f', 2, 64)
}

// ========================================================================

type WriteCounter int64

func (counter *WriteCounter) Write(data []byte) (n int, _ error) {
	n = len(data)
	*counter += WriteCounter(n)
	return
}
