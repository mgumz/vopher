package main

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

// JobUI defines the interface for all vopher-UIs
type JobUI interface {
	Start()
	Stop()

	Refresh()

	AddJob(jobID string)
	Print(jobID, msg string)
	JobDone(jobID string)
	Wait() // wait for all jobs to be .Done()
}
