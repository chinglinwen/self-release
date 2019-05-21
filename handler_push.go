package main

import (
	"fmt"
	"log"
	"strings"

	"wen/self-release/template"
)

// get autoenv from events

const DevBranch = "develop"

/*
existing env

[wen@234 k8snew ~]$ awk '{ print $2 }' FS='{{'  a| tr -d '}' | grep -v -e '^$' | sort -n | uniq
 $CI_ENV
 $CI_IMAGE
 $CI_NAMESPACE
 $CI_NAMESPACE ,project=
 $CI_PROJECT_NAME
 $CI_PROJECT_NAME_WITH_ENV
 $CI_REPLICAS
 $CI_TIME
 $CI_USER_NAME
 $NODE_PORT  # ????
[wen@234 k8snew ~]$
*/

// receive push, do the build for test,  or filter out based on commit text? Force keyword?
func handlePush(event *PushEvent) (err error) {
	log.Printf("got project %v to build for test env\n", event.Project.PathWithNamespace)

	project := event.Project.PathWithNamespace
	branch := parseBranch(event.Ref)
	if branch == errParseRefs {
		err = fmt.Errorf("project: %v, parse branch err for refs: %v", project, event.Ref)
		return
	}
	autoenv := make(map[string]string)
	autoenv["PROJECTPATH"] = project
	autoenv["BRANCH"] = branch
	autoenv["USERNAME"] = event.UserName
	autoenv["USEREMAIL"] = event.UserEmail
	autoenv["MSG"] = event.Commits[0].Message
	log.Println("autoenv:", autoenv)

	// only build for develop branch, need confirm?
	// we shoult not limit the branch, let them easy to change? change in config.yaml, based on tag?
	//so release need to build too? or just add addition condition to build for other branch?

	// what to do with master branch as dev?  init by commit text?

	// if not inited, just using default setting?
	p, err := template.NewProject(project, template.SetBranch(branch), template.SetAutoEnv(autoenv))
	if err != nil {
		err = fmt.Errorf("project: %v, new err: %v", project, err)
		return
	}

	// it should be config from repo or template now
	// if p.DevBranch == "" {
	// 	p.DevBranch = "develop"
	// }

	// this should be check later, by see config first
	// if I were them, I just do release, let the system figure out when to init?
	// release to test? it's better to init by tag msg?
	if branch != p.DevBranch { // tag should be release, not build?
		log.Printf("ignore build of branch: %v (devBranch=%q) from project: %v", branch, p.DevBranch, project)
		return
	}

	var force bool
	if strings.Contains(event.Commits[0].Message, "/force") {
		force = true
	}
	// check if force is enabled

	// check if inited

	if !p.Inited() || force {
		err = p.Init(template.SetInitForce())
		if err != nil {
			err = fmt.Errorf("project: %v, init err: %v", project, err)
			return
		}
	} else {
		log.Printf("will not init for project: %v", project)
	}

	// almost generate everytime, except config
	err = p.Generate()
	if err != nil {
		err = fmt.Errorf("project: %v, generate before build err: %v", project, err)
		return
	}
	log.Printf("done generate for project: %v", project)

	// everytime build, need generate first

	// write to a auto.env? or
	//envsubst.Eval()

	log.Printf("start building for project: %v, branch: %v\n", project, branch)
	out, err := p.Build(project, branch)
	if err != nil {
		err = fmt.Errorf("build err: %v", err)
		return
	}
	fmt.Println("output:", out)
	// check if inited or force provide, if not, init first

	// builded, how to relate the image?

	// have a apply script to do that? passing same tag to it, for the image part?
	// is it need re-generate? provided env is change everytime though

	return
}

const errParseRefs = "parseRefsError"

func parseBranch(refs string) string {
	refss := strings.SplitAfter(refs, "/")
	if len(refss) == 3 {
		return refss[2]
	}
	return errParseRefs
}

// receive tag release, do the build for pre,  or filter based on commit text?
// it should be the same image as test, so no need to build image again? image name is been fixed by build
//
// if project set to auto, we auto tag for master? or just directly
func handleRelease(event *TagPushEvent) (err error) {
	log.Printf("got project %v to build for pre or online env\n", event.Project.PathWithNamespace)

	autoenv := make(map[string]string)
	autoenv["PROJECTPATH"] = event.Project.PathWithNamespace
	autoenv["BRANCH"] = event.Ref
	autoenv["USERNAME"] = event.UserName
	autoenv["USEREMAIL"] = event.UserEmail
	autoenv["MSG"] = event.Message

	fmt.Println("autoenv:", autoenv)

	return
}
