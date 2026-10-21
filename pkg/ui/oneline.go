package ui

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mgumz/vopher/pkg/vopher"
)

// UIOneLine implements a UI which draws the progress in "one line"
// onto the terminal.
type OneLine struct {
	sync.WaitGroup
	pt       *ProgressTicker
	prefix   string
	duration time.Duration
	origins  [vopher.OriginKept + 1]atomic.Int32
}

func (ol *OneLine) Start() {
	ol.pt.note = ol.originNote
	ol.pt.start(ol.prefix, ol.duration)
}
func (ol *OneLine) Stop()                     { ol.pt.stop() }
func (ol *OneLine) AddJob(string)             { ol.Add(1); ol.pt.max.Add(1) }
func (ol *OneLine) JobDone(string)            { ol.pt.counter.Add(1); ol.Done() }
func (ol *OneLine) Wait()                     { ol.WaitGroup.Wait(); ol.pt.maxOut(); ol.Refresh() }
func (ol *OneLine) Refresh()                  { ol.pt.print(ol.prefix) }
func (ol *OneLine) Print(string, string)      {}
func (ol *OneLine) PrintShort(string, string) {}

func (ol *OneLine) JobOrigin(_ string, origin vopher.Origin) {
	if int(origin) < len(ol.origins) {
		ol.origins[origin].Add(1)
	}
}

// originNote renders "↓downloaded =unchanged #kept", or "" as long as no
// job reported its origin
func (ol *OneLine) originNote() string {
	down := ol.origins[vopher.OriginDownloaded].Load()
	same := ol.origins[vopher.OriginNotModified].Load()
	kept := ol.origins[vopher.OriginKept].Load()
	if down+same+kept == 0 {
		return ""
	}
	return fmt.Sprintf("↓%d =%d #%d", down, same, kept)
}
