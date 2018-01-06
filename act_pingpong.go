package main

import (
	"time"
)

// actPingPong is test-ui
func actPingPong(ui JobUI) {
	ui.Start()

	for i := 0; i < 10; i++ {
		ui.AddJob("ping")
		<-time.After(500 * time.Millisecond)
		ui.JobDone("pong")
	}

	ui.Wait()
	ui.Stop()
}
