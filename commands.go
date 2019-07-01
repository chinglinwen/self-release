package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"wen/self-release/pkg/sse"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
)

// type action struct {
// 	name string
// 	fn   func(string) (string, error)
// }

type action struct {
	name string
	help string
	fn   func(string, string) (string, error)
}

var (
	// funcs = map[string]action{
	// 	// "help":      help, // can't refer back to help
	// 	"demo":      demo,
	// 	"deploy":    deploy,
	// 	"retry":     retry,
	// 	"myproject": myproject,
	// }

	funcs = []action{
		// {name: "help", fn: help},
		{name: "hi", fn: hi, help: "say hi."},
		{name: "deploy", fn: deploy, help: "deploy project. (eg: /deploy group/project [branch] )"},
		{name: "rollback", fn: rollback, help: "rollback project. (eg: /rollback group/project [branch] )"},
		{name: "retry", fn: retry, help: "retry last time deployed project."},
		{name: "myproject", fn: myproject, help: "get last time project."},
	}
)

func doAction(dev, cmd string) (out string, err error) {
	cmd = strings.TrimPrefix(cmd, "/")
	c := strings.Fields(cmd)[0]
	args := strings.TrimPrefix(cmd, c)

	// fn, ok := funcs[c]
	// if !ok {
	// 	return help(dev, "")
	// }
	// return fn(dev, args)

	var found bool
	for _, v := range funcs {
		if v.name != c {
			continue
		}
		found = true
		return v.fn(dev, args)
	}
	if !found {
		return help(dev, "")
	}
	return
}

func help(dev, args string) (out string, err error) {
	out = "list of actions:\n"
	for _, v := range funcs {
		out = fmt.Sprintf("%v/%v   -> %v\n", out, v.name, v.help)
	}
	// out = fmt.Sprintf("list of actions:\n%v", helplist())
	return
}

func helplist() (out string) {
	w := new(tabwriter.Writer)
	var b bytes.Buffer
	w.Init(&b, 5, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)

	for _, v := range funcs {
		fmt.Fprintf(w, "%v \t %v\n", v.name, v.help)
		// out = fmt.Sprintf("%v\t%v\n  %v", out, v.name, v.help)
	}
	w.Flush()
	out = b.String()
	return
}

func hi(dev, args string) (out string, err error) {
	out = fmt.Sprintf("hello %v, you provided cmd: hi, args: %v", dev, args)
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

// rollbacks, need to get last tag? support online only?
// rollback to just pre-version or to any specific version
// kind of redeploy? but auto get the last tag?
func rollback(dev, args string) (out string, err error) {
	log.Printf("got rollback from: %v, args: %v\n", dev, args)
	project, branch, _ := parseProject(args) // ignore no project provide error

	if project == "" {
		brocker, e := sse.GetBrokerFromPerson(dev)
		if e != nil {
			err = fmt.Errorf("cant find previous released project name to rollback, " +
				"try provide project name for rollback")
			return
		}
		// spew.Dump("retry brocker:", brocker)

		b := &builder{
			Broker: sse.NewExist(brocker),
		}
		project = b.Event.Project
		log.Println("will try rollback project: ", project)
	}

	var p *projectpkg.Project
	if branch == "" {
		p, err = getproject(project, branch, true)
		// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
		if err != nil {
			err = fmt.Errorf("project: %v, new err: %v", project, err)
			return
		}
		branch = p.Branch
	}

	// booptions := []string{"gen", "build", "deploy"}
	bo := &buildOption{
		// gen:      contains(booptions, "gen"),
		// build:    contains(booptions, "build"),
		// deploy:   contains(booptions, "deploy"),
		rollback: true,
		nonotify: true,
		p:        p,
	}
	e := &sse.EventInfo{
		Project: project,
		Branch:  branch, // get branch automatic here if not specified
		// Env:       env, // default derive from branch
		UserName: dev,
		// UserEmail: useremail,
		Message: fmt.Sprintf("from %v, args: ", args),
	}

	b := NewBuilder(project, branch)
	b.log("starting logs")

	err = b.startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("rollback for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		return
	}
	if err == nil {
		out = "rollback ok"
		log.Printf("rollback from %v ok\n", dev)
	}
	return
}
