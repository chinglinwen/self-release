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

const InitBranch = "develop"

// type action struct {
// 	name string
// 	fn   func(string) (string, error)
// }

type action struct {
	name string
	help string
	eg   string
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
		{name: "deploy", fn: deploy, help: "deploy project.", eg: "/deploy group/project [branch][nobuild|buildimage]"},
		{name: "deldeploy", fn: deldeploy, help: "delete deploy project.", eg: "/deldeploy group/project [branch]"},
		// {name: "rollback", fn: rollback, help: "rollback project.", eg: "/rollback group/project [branch]"},
		{name: "retry", fn: retry, help: "retry last time deployed project.", eg: "/retry [nobuild|buildimage]"},
		// {name: "reapply", fn: reapply, help: "reapply last time deployed project without build image.", eg: "/reapply [group/project] [branch]"},
		{name: "set", fn: setting, help: "setting project config.", eg: "/set [group/project] [buildmode=auto|disabled|on|manual][devbranch=develop|test]" +
			"[configver=php.v1][selfrelease=enabled|disabled][viewsetting]"},
		// {name: "gen", fn: gen, help: "generate files(yaml) only last time deployed project.", eg: "/gen [group/project] [branch]"},
		{name: "myproject", fn: myproject, help: "show last time project."},
		{name: "helpdocker", fn: helpdocker, help: "help to generate docker files(in branch develop).", eg: "/helpdocker group/project [force]"},
		{name: "init", fn: projectinit, help: "enable project by init config-repo.", eg: "/init group/project [force]"},
	}
)

func doAction(dev, cmd string) (out string, err error) {
	cmd = strings.TrimPrefix(cmd, "/")
	c := strings.Fields(cmd)[0]
	args := strings.TrimPrefix(cmd, c)

	// tolerate other args too
	project, branch, _ := parseProject(args)
	if project != "" {
		err = sse.Lock(project, branch)
		if err != nil {
			return
		}
		defer sse.UnLock(project, branch)
	}

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
	out = "list of actions:\n\n"
	for _, v := range funcs {
		help := v.help
		if v.eg != "" {
			help += "\n          eg: " + v.eg
		}
		out = fmt.Sprintf("%v/%v  -> %v\n\n", out, v.name, help)
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

func helpdocker(dev, args string) (out string, err error) {
	args += " dockeronly"
	return projectinit(dev, args)
}

// how to support setting?
func projectinit(dev, args string) (out string, err error) {
	s := parseSetting(args)
	err = validateSetting(s)
	if err != nil {
		log.Println("validate setting err: ", err)
		return
	}

	project, branch, err := parseProject(args)
	if err != nil {
		return
	}
	f := parseFlag(args)

	if branch == "" {
		branch = InitBranch // set default
	}

	// c := projectpkg.ProjectConfig{
	// 	BuildMode: s.buildmode,
	// 	DevBranch: s.devbranch,
	// 	ConfigVer: s.configver,
	// }

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch), projectpkg.SetNoEnableCheck(true))
	if err != nil {
		err = fmt.Errorf("new project: %v, err: %v", project, err)
		return
	}
	if f.force {
		err = p.Init(projectpkg.SetInitDockerOnly(f.dockeronly), projectpkg.SetInitForce())
	} else {
		err = p.Init(projectpkg.SetInitDockerOnly(f.dockeronly))
	}
	if err == nil {
		out = "init ok"
		log.Printf("init from %v ok\n", dev)
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
	f := parseFlag(args)
	bo := &buildOption{
		// gen:        true,
		nobuild:    f.nobuild,
		buildimage: f.buildimage,
		deploy:     true,
		nonotify:   true,
	}
	e := &sse.EventInfo{
		Project: project,
		Branch:  branch,
		// Env:       env, // default derive from branch
		UserName: dev,
		// UserEmail: useremail,
		Message: fmt.Sprintf("from wechat %v, args: %v ", dev, args),
	}

	b := NewBuilder(project, branch)
	b.log("starting logs")

	err = b.startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("startdeploy for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		return
	}
	if err == nil {
		out = "deployed ok"
		log.Printf("deploy from %v ok\n", dev)
	}
	return
}

// deploy
func gen(dev, args string) (out string, err error) {
	log.Printf("got gen from: %v, args: %v\n", dev, args)
	project, branch, err := parseProject(args)
	if err != nil {
		return
	}
	bo := &buildOption{
		gen:      true,
		nobuild:  true,
		deploy:   false,
		nonotify: true,
	}
	e := &sse.EventInfo{
		Project: project,
		Branch:  branch,
		// Env:       env, // default derive from branch
		UserName: dev,
		// UserEmail: useremail,
		Message: fmt.Sprintf("from %v, args: %v ", dev, args),
	}

	b := NewBuilder(project, branch)
	b.log("starting logs")

	err = b.startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("startgen for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		return
	}
	if err == nil {
		out = "gen ok"
		log.Printf("gen from %v ok\n", dev)
	}
	return
}

// // automatic support specify env(no need specific tag) as second args
// func deldeploy(dev, args string) (out string, err error) {
// 	log.Printf("got deldeploy from: %v, args: %v\n", dev, args)
// 	project, branch, err := parseProject(args)
// 	if err != nil {
// 		return
// 	}
// 	out, err = projectpkg.DeleteByKubectl(project, branch, "")
// 	if err != nil {
// 		err = fmt.Errorf("deldeploy for project: %v, branch: %v, err: %v", project, branch, err)
// 		log.Println(err)
// 		return
// 	}
// 	if err == nil {
// 		out = "deldeploy ok"
// 		log.Printf("deldeploy from %v ok\n", dev)
// 	}
// 	return
// }

// automatic support specify env(no need specific tag) as second args
func deldeploy(dev, args string) (out string, err error) {
	log.Printf("got deldeploy from: %v, args: %v\n", dev, args)
	project, branch, err := parseProject(args)
	if err != nil {
		return
	}
	out, err = deleteReleaseFromCommand(project, branch)
	if err != nil {
		err = fmt.Errorf("deldeploy for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		return
	}
	if err == nil {
		out = "deldeploy ok"
		log.Printf("deldeploy from %v ok\n", dev)
	}
	return
}

func argstoevent(e *sse.EventInfo, args string) {
	project, branch, err := parseProject(args)
	if err != nil {
		return
	}
	if project != "" {
		e.Project = project
	}
	if branch != "" {
		e.Branch = branch
		// env will auto derive later if empty
	}
	log.Printf("convert args info to event for: %v, branch: %v\n", project, branch)
}

// retry
func retry(dev, args string) (out string, err error) {
	log.Println("got retry from ", dev)
	f := parseFlag(args)
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
	bo := &buildOption{
		gen: true,
		// build:    true, // default to configs
		nobuild:    f.nobuild,
		buildimage: f.buildimage, // should no build again?
		deploy:     true,
		nonotify:   true,
	}
	argstoevent(b.Event, args)

	log.Println("start retry build for ", dev)
	err = b.startBuild(b.Event, bo)
	if err == nil {
		out = "retried ok"
		log.Printf("retry from %v ok\n", dev)
	}
	return
}

// retry
func setting(dev, args string) (out string, err error) {
	log.Println("got setting from ", dev)
	s := parseSetting(args)
	err = validateSetting(s)
	if err != nil {
		log.Println("validate setting err: ", err)
		return
	}
	f := parseFlag(args)

	project, _, err := parseProject(args)
	if err != nil {
		brocker, e := sse.GetBrokerFromPerson(dev)
		if e != nil {
			err = fmt.Errorf("no prorject provided and cant find previous project, err: %v", e)
			log.Println(err)
			return
		}
		project = brocker.Project
	}

	log.Println("start project setting for ", dev)
	p, err := projectpkg.NewProject(project, projectpkg.SetNoEnableCheck(true))
	if err != nil {
		err = fmt.Errorf("new project: %v", err)
		return
	}
	if f.viewsetting {
		out = fmt.Sprint(p.Config)
		log.Printf("setting from %v ok\n", dev)
		return
	}

	var enabled bool
	if s.selfrelease == "enabled" {
		enabled = true
	}

	c := projectpkg.ProjectConfig{
		S: projectpkg.SelfRelease{
			BuildMode: s.buildmode,
			DevBranch: s.devbranch,
			ConfigVer: s.configver,
			Enable:    enabled,
			Version:   s.version,
		},
	}
	out, err = p.Setting(c)
	if err != nil {
		return
	}

	out = fmt.Sprintf("setting ok for %v, %v", project, out)
	log.Printf("setting from %v ok\n", dev)
	return
}

// retry
func reapply(dev, args string) (out string, err error) {
	log.Println("got reapply from ", dev)

	brocker, err := sse.GetBrokerFromPerson(dev)
	if err != nil {
		fmt.Println("cant find previous released project")
		return
	}

	b := &builder{
		Broker: sse.NewExist(brocker),
	}
	// spew.Dump(b)
	if b.PWriter == nil {
		err = fmt.Errorf("pwriter nil, can't write msg")
		return
	}

	bo := &buildOption{
		gen:      true,
		nobuild:  true, // no build image again?
		deploy:   true,
		nonotify: true,
	}
	argstoevent(b.Event, args)

	log.Println("start reapply build for ", dev)
	err = b.startBuild(b.Event, bo)
	if err == nil {
		out = "reapply ok"
		log.Printf("reapply from %v ok\n", dev)
	}
	return
}

// rollbacks, need to get last tag? support online only?
// rollback to just pre-version or to any specific version
// kind of redeploy? but auto get the last tag?
// func rollback(dev, args string) (out string, err error) {
// 	log.Printf("got rollback from: %v, args: %v\n", dev, args)
// 	project, branch, _ := parseProject(args) // ignore no project provide error

// 	if project == "" {
// 		brocker, e := sse.GetBrokerFromPerson(dev)
// 		if e != nil {
// 			err = fmt.Errorf("cant find previous released project name to rollback, " +
// 				"try provide project name for rollback")
// 			return
// 		}
// 		// spew.Dump("retry brocker:", brocker)

// 		b := &builder{
// 			Broker: sse.NewExist(brocker),
// 		}
// 		project = b.Event.Project
// 		log.Println("will try rollback project: ", project)
// 	}

// 	var p *projectpkg.Project
// 	if branch == "" {
// 		p, err = getproject(project, branch, true, false)
// 		// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
// 		if err != nil {
// 			err = fmt.Errorf("project: %v, new err: %v", project, err)
// 			return
// 		}
// 		branch = p.Branch
// 	}

// 	// booptions := []string{"gen", "build", "deploy"}
// 	bo := &buildOption{
// 		// gen:      contains(booptions, "gen"),
// 		// build:    contains(booptions, "build"),
// 		// deploy:   contains(booptions, "deploy"),
// 		rollback: true,
// 		nonotify: true,
// 		p:        p,
// 	}
// 	e := &sse.EventInfo{
// 		Project: project,
// 		Branch:  branch, // get branch automatic here if not specified
// 		// Env:       env, // default derive from branch
// 		UserName: dev,
// 		// UserEmail: useremail,
// 		Message: fmt.Sprintf("from %v, args: ", args),
// 	}

// 	b := NewBuilder(project, branch)
// 	b.log("starting logs")

// 	err = b.startBuild(e, bo)
// 	if err != nil {
// 		err = fmt.Errorf("rollback for project: %v, branch: %v, err: %v", project, branch, err)
// 		log.Println(err)
// 		return
// 	}
// 	if err == nil {
// 		out = "rollback ok"
// 		log.Printf("rollback from %v ok\n", dev)
// 	}
// 	return
// }
