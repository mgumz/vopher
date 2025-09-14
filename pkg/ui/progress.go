package ui

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

	cols := 0

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
		progressNumberFmt      = " %d%% "
	)

	// the info field / component
	info := fmt.Sprintf(infoFmt, prefix, pt.counter, pt.max)

	// the progress bar
	barWidth := cols - len(info) - borderWidth
	bar := bytes.Repeat([]byte("."), barWidth)

	nticks := int(math.Max(1.0, math.Floor(float64(barWidth)*pt.progress())))

	for i := range nticks {
		bar[i] = '='
	}
	if nticks < barWidth {
		bar[nticks-1] = '>'
	}
	bar[0], bar[barWidth-1] = '[', ']'

	showProgressNumber := (barWidth > progressThresholdWidth)
	if showProgressNumber {
		p := int(100.0 * pt.progress())
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
