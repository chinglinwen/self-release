package main

import (
	"fmt"
	"strings"
)

type flagOption struct {
	force       bool
	buildimage  bool
	nobuild     bool
	dockeronly  bool
	viewsetting bool
}

func parseFlag(args string) (f flagOption) {
	s := strings.Fields(args)
	a := []string{}
	for _, v := range s {
		if strings.Contains(v, "=") {
			continue
		}
		a = append(a, v)
	}
	for _, v := range a {
		if strings.Contains(v, "=") {
			continue
		}
		if strings.Contains(v, "force") {
			f.force = true
		}
		if strings.Contains(v, "nobuild") {
			f.nobuild = true
		}
		if strings.Contains(v, "dockeronly") {
			f.dockeronly = true
		}
		if strings.Contains(v, "viewsetting") {
			f.viewsetting = true
		}
		if strings.Contains(v, "buildimage") {
			f.buildimage = true
		}
	}
	return
}

func parseProject(args string) (project, branch string, err error) {
	args1 := strings.NewReplacer(" force", " ", " nobuild", " ", " dockeronly",
		" ", " viewsetting", " ", " buildimage", " ").Replace(args)
	s := strings.Fields(args1)
	a := []string{}
	for _, v := range s {
		if strings.Contains(v, "=") {
			continue
		}
		a = append(a, v)
	}

	if len(a) == 0 {
		err = fmt.Errorf("no project arg provided")
		return
	}
	if len(a) < 2 {
		project = a[0]
		branch = "develop" // TODO: using config?
	}
	if len(a) >= 2 {
		project = a[0]
		branch = a[1]
		// if branch == "force" || branch == "nobuild" {
		// 	branch = "develop"
		// }
	}
	if strings.TrimSpace(project) == "" {
		err = fmt.Errorf("no project arg provided")
		return
	}
	return
}

type setOption struct {
	buildmode   string
	configver   string
	devbranch   string
	selfrelease string
	version     string
}

func validateSetting(s setOption) error {
	if s.buildmode != "" {
		if s.buildmode != "auto" && s.buildmode != "disabled" && s.buildmode != "on" {
			return fmt.Errorf("expect auto,disabled,on for imagebuild, but got: %v", s.buildmode)
		}
	}
	if s.selfrelease != "" {
		if s.selfrelease != "enabled" && s.selfrelease != "disabled" {
			return fmt.Errorf("expect enabled,disabled for selfrelease, but got: %v", s.selfrelease)
		}
	}
	return nil
}

func parseSetting(args string) (f setOption) {
	// f = setOption{}
	s := strings.Fields(args)
	m := make(map[string]string)
	for _, v := range s {
		if !strings.Contains(v, "=") {
			continue
		}
		ss := strings.Split(v, "=")
		if len(ss) != 2 {
			continue
		}
		m[ss[0]] = ss[1]
	}
	if m["imagebuild"] != "" {
		f.buildmode = m["imagebuild"]
	}
	if m["devbranch"] != "" {
		f.devbranch = m["devbranch"]
	}
	if m["configver"] != "" {
		f.configver = m["configver"]
	}
	if m["selfrelease"] != "" {
		f.selfrelease = m["selfrelease"]
	}
	if m["version"] != "" {
		f.version = m["version"]
	}

	return
}
