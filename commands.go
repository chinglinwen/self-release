package main

import (
	"fmt"
	"strings"
	"wen/self-release/pkg/sse"

	"github.com/chinglinwen/log"
)

// type action struct {
// 	name string
// 	fn   func(string) (string, error)
// }

type action func(string, string) (string, error)

var (
	funcs = map[string]action{
		// "help":      help, // can't refer back to help
		"demo":      demo,
		"deploy":    deploy,
		"retry":     retry,
		"myproject": myproject,
	}

// funcs = []action{}
// {name: "help", fn: help},
// "help":      help,
// "demo":      demo,
// "retry":     retry,
// "myproject": myproject,
// }
)

func doAction(dev, cmd string) (out string, err error) {
	cmd = strings.TrimPrefix(cmd, "/")
	c := strings.Fields(cmd)[0]
	args := strings.TrimPrefix(cmd, c)

	fn, ok := funcs[c]
	if !ok {
		return help(dev, "")
	}
	return fn(dev, args)
}

func help(dev, args string) (out string, err error) {
	out = "list of actions:"
	for k, _ := range funcs {
		out = fmt.Sprintf("%v\n  %v", out, k)
	}
	return
}

func demo(dev, args string) (out string, err error) {
	out = fmt.Sprintf("hello %v, you provided cmd: demo, args: %v", dev, args)
	return
}

func myproject(dev, args string) (out string, err error) {
	log.Println("got myproject from ", dev)

	brocker, err := sse.GetBrokerFromPerson(dev)
	if err != nil {
		fmt.Println("cant find previous released project")
		return
	}
	b := &builder{
		Broker: brocker,
	}
	out = fmt.Sprintf("project: %v, branch: %v", b.Project, b.Branch)
	return
}

// // make this into project config?
// func convertback(name string) string {
// 	if name == "wen" {
// 		return "wenzhenglin"
// 	}
// 	return name
// }

func parseProject(args string) (project, branch string, err error) {
	s := strings.Fields(args)
	if len(s) < 1 {
		err = fmt.Errorf("no project arg provided")
		return
	}
	if len(s) < 2 {
		project = s[0]
		branch = "develop" // TODO: using config?
	}
	if len(s) >= 2 {
		project = s[0]
		branch = s[1]
	}
	return
}

// deploy
func deploy(dev, args string) (out string, err error) {
	log.Printf("got deploy from: %v, args: %v\n", dev, args)
	project, branch, err := parseProject(args)
	if err != nil {
		return
	}
	booptions := []string{"gen", "build", "deploy"}
	bo := &buildOption{
		gen:      contains(booptions, "gen"),
		build:    contains(booptions, "build"),
		deploy:   contains(booptions, "deploy"),
		nonotify: true,
	}
	e := &sse.EventInfo{
		Project: project,
		Branch:  branch,
		// Env:       env, // default derive from branch
		UserName: dev,
		// UserEmail: useremail,
		Message: fmt.Sprintf("from %v, args: ", args),
	}

	b := NewBuilder(project, branch)
	b.log("starting logs")

	err = b.startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("startBuild for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		return
	}
	if err == nil {
		out = "deployed ok"
		log.Printf("deploy from %v ok\n", dev)
	}
	return
}

// retry
func retry(dev, args string) (out string, err error) {
	log.Println("got retry from ", dev)

	brocker, err := sse.GetBrokerFromPerson(dev)
	if err != nil {
		fmt.Println("cant find previous released project")
		return
	}
	// spew.Dump("retry brocker:", brocker)

	b := &builder{
		Broker: sse.NewExist(brocker),
	}
	// spew.Dump(b)
	if b.PWriter == nil {
		err = fmt.Errorf("pwriter nil, can't write msg")
		return
	}
	booptions := []string{"gen", "build", "deploy"}
	bo := &buildOption{
		gen:      contains(booptions, "gen"),
		build:    contains(booptions, "build"),
		deploy:   contains(booptions, "deploy"),
		nonotify: true,
	}
	log.Println("start retry build for ", dev)
	err = b.startBuild(b.Event, bo)
	if err == nil {
		out = "retried ok"
		log.Printf("retry from %v ok\n", dev)
	}
	return
}

// rollbacks
