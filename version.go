package main

import "fmt"

var (
	Version   = "0.6.0"
	GitHash   = ""
	BuildDate = ""
)

func printVersion() {

	fmt.Println("vopher:\t" + Version)
	if GitHash != "" {
		fmt.Println("git:\t" + GitHash)
	}
	if BuildDate != "" {
		fmt.Println("build:\t" + BuildDate)
	}
}
