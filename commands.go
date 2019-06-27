package main

import (
	"fmt"
	"wen/self-release/pkg/sse"

	"github.com/chinglinwen/log"
)

func demo(dev string) (out string, err error) {
	out = "hello there"
	return
}

// retry
func retry(dev string) (out string, err error) {
	log.Println("got retry from ", dev)

	brocker, err := sse.GetBrokerFromPerson(dev)
	if err != nil {
		fmt.Println("cant find previous released project")
		return
	}
	b := &builder{
		Broker: brocker,
	}
	booptions := []string{"gen", "build", "deploy"}
	bo := &buildOption{
		gen:    contains(booptions, "gen"),
		build:  contains(booptions, "build"),
		deploy: contains(booptions, "deploy"),
	}
	out = "retried"
	err = b.startBuild(b.Event, bo)
	return
}

// rollbacks
