package main

//
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

// the ui-interface
type JobUi interface {
	Start()
	Stop()

	Refresh()

	AddJob(job_id string)
	JobDone(job_id string)
	Wait() // wait for all jobs to be .Done()
}
