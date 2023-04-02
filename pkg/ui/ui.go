package ui

import (
	"time"

	"github.com/mgumz/vopher/pkg/vopher"
)

// ui-ideas
//
// * https://godoc.org/github.com/jroimartin/gocui
//
//  global-progress [..............]
//  plugin1         [....]
//  plugin2         [............]
//  plugin3         [..............]
//
// cons: vertical space
//
// ui-option2:
//   <-> global progress
//  [....|.....|.....|....|....|....]
//   ^
//   | plugin-progress via _-=#░█▓▒░█
//   v
//
// cons: horizontal space
//        plugin-name fehlt

// UI defines the interface for all vopher-UIs
type UI interface {
	Start()
	Stop()

	Refresh()

	AddJob(id string)
	Print(id, msg string)
	PrintShort(id, msg string)
	JobDone(id string)
	Wait() // wait for all jobs to be .Done()
}

func NewUI(ui string) UI {
	switch ui {
	case "oneline":
		return &OneLine{
			pt:       newProgressTicker(0),
			prefix:   "vopher",
			duration: 25 * time.Millisecond,
		}
	case "simple":
		return &Simple{jobs: make(map[string]*vopher.Runtime)}
	case "quiet":
		return &Quiet{}
	}
	return nil
}
