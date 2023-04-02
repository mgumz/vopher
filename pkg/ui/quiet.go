package ui

import (
	"sync"

	"github.com/mgumz/vopher/pkg/vopher"
)

// Quiet is kind of a "no UI" coz it gives no feedback at all
type Quiet struct {
	sync.WaitGroup
	vopher.Runtime
}

func (quiet *Quiet) Start()                    { quiet.Runtime.Start() }
func (quiet *Quiet) Stop()                     { quiet.Runtime.Stop() }
func (quiet *Quiet) AddJob(id string)          { quiet.WaitGroup.Add(1) }
func (quiet *Quiet) JobDone(id string)         { quiet.WaitGroup.Done() }
func (quiet *Quiet) Print(id, msg string)      {}
func (quiet *Quiet) PrintShort(id, msg string) {}
func (quiet *Quiet) Wait()                     { quiet.WaitGroup.Wait() }
func (quiet *Quiet) Refresh()                  {}
