package main

import (
	"log"

	"github.com/google/gops/agent"
)

func gopsStart() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true, // automatically closes on os.Interrupt
	}); err != nil {
		log.Fatal(err)
	}
}

func gopsStop() {
	agent.Close()
}
