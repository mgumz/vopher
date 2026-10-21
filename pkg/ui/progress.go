package ui

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type ProgressTicker struct {
	WriteMeter
	note   func() string // optional, shown after the counter
	ticker *time.Ticker
	stopCh chan bool
}

func newProgressTicker(max int64) *ProgressTicker {
	pt := &ProgressTicker{}
	pt.max.Store(max)
	return pt
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

	// one snapshot, so counter, max and bar agree with each other
	counter, max := pt.counter.Load(), pt.max.Load()
	if max == 0 {
		return
	}
	progress := float64(counter) / float64(max)

	var cols int

	if colStr, exists := os.LookupEnv("COLUMNS"); exists {
		val, err := strconv.ParseInt(colStr, 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("COLUMNS is not a number: %s: %s", colStr, err)
			log.Fatal(errStr)
			return
		}
		cols = int(val)
	} else {
		cols, _, _ = terminalSize(os.Stdout)
		if cols <= 0 {
			log.Fatal("can't get TerminalSize(), use other -ui type")
			return
		}
	}

	//
	// |$prefix: (16/32) [=====>.. 50% .......] |
	// ^ ^^^^^^^^^^^^^^^                        ^
	// | info part       ^^^^^^^^^^^^^^^^^^^^^^ |
	// |                  progress bar          |
	// \ left screen               right screen /
	const (
		borderWidth            = 2  // 1 left/right of progress bar
		progressThresholdWidth = 10 // "[= 100% =]" <- 10 chars
		infoFmt                = "%s: (%d/%d)"
		infoNoteFmt            = "%s: (%d/%d %s)"
		progressNumberFmt      = " %d%% "
	)

	// the info field / component
	info := fmt.Sprintf(infoFmt, prefix, counter, max)
	if pt.note != nil {
		if note := pt.note(); note != "" {
			info = fmt.Sprintf(infoNoteFmt, prefix, counter, max, note)
		}
	}

	// the progress bar
	barWidth := cols - utf8.RuneCountInString(info) - borderWidth
	bar := bytes.Repeat([]byte("."), barWidth)

	nticks := int(math.Max(1.0, math.Floor(float64(barWidth)*progress)))

	for i := range nticks {
		bar[i] = '='
	}
	if nticks < barWidth {
		bar[nticks-1] = '>'
	}
	bar[0], bar[barWidth-1] = '[', ']'

	showProgressNumber := (barWidth > progressThresholdWidth)
	if showProgressNumber {
		p := int(100.0 * progress)
		progress := fmt.Sprintf(progressNumberFmt, p)
		halfBar := barWidth / 2
		halfProgress := len(progress) / 2
		posProgress := (halfBar - halfProgress)
		copy(bar[posProgress:], progress)
	}

	// using cursor-up+progress+newline works more stable than to \r
	// the cursor.
	_ = cursorNUp(os.Stdout, 1)

	// finally: display the progress bar
	fmt.Println(info, string(bar))
}

func (pt *ProgressTicker) maxOut() {
	pt.counter.Store(pt.max.Load())
}

// ========================================================================

// WriteMeter is safe for concurrent use
type WriteMeter struct {
	counter WriteCounter
	max     atomic.Int64
}

func (meter *WriteMeter) progress() float64 {
	return float64(meter.counter.Load()) / float64(meter.max.Load())
}

func (meter *WriteMeter) String() string {
	return strconv.FormatFloat(meter.progress(), 'f', 2, 64)
}

// ========================================================================

type WriteCounter struct{ atomic.Int64 }

func (counter *WriteCounter) Write(data []byte) (n int, _ error) {
	n = len(data)
	counter.Add(int64(n))
	return
}
