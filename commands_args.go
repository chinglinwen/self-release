package main

import (
	"fmt"
	"strings"
)

type flagOption struct {
	force   bool
	nobuild bool
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
	}
	return
}

func parseProject(args string) (project, branch string, err error) {
	s := strings.Fields(args)
	a := []string{}
	for _, v := range s {
		if strings.Contains(v, "=") {
			continue
		}
		a = append(a, v)
	}

	if len(a) < 1 {
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
		if branch == "force" || branch == "nobuild" {
			branch = "develop"
		}
	}
	return
}

type setOption struct {
	buildmode string
	configver string
	devbranch string
}

func validateSetting(s setOption) error {
	if s.buildmode != "" {
		if s.buildmode != "auto" && s.buildmode != "disabled" && s.buildmode != "on" {
			return fmt.Errorf("expect auto,disabled,on for imagebuild, but got: %v", s.buildmode)
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
	return
}
