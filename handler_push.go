package main

import (
	"fmt"
	"log"
	"strings"

	projectpkg "wen/self-release/project"
)

// get autoenv from events

const DevBranch = "develop"

/*
existing env

$ awk '{ print $2 }' FS='{{'  a| tr -d '}' | grep -v -e '^$' | sort -n | uniq
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
$
*/

// should we build for every branch
//
// receive push, do the build for test,  or filter out based on commit text? Force keyword?
func handlePush(event *PushEvent) (err error) {
	log.Printf("got project %v to build for test env\n", event.Project.PathWithNamespace)

	// project := event.Project.PathWithNamespace // we don't use event.Project.Name, since it may be chinese
	// namespace, projectName, err := projectpkg.GetProjectName(project)
	// if err != nil {
	// 	err = fmt.Errorf("parse project name for %q, err: %v", project, err)
	// 	return
	// }

	// branch := parseBranch(event.Ref)
	// if branch == errParseRefs {
	// 	err = fmt.Errorf("project: %v, parse branch err for refs: %v", project, event.Ref)
	// 	return
	// }
	// env := projectpkg.GetEnvFromBranch(branch)

	// autoenv := make(map[string]string)
	// autoenv["CI_PROJECT_PATH"] = project
	// autoenv["CI_BRANCH"] = branch
	// autoenv["CI_ENV"] = env
	// autoenv["CI_NAMESPACE"] = namespace
	// autoenv["CI_PROJECT_NAME"] = projectName
	// autoenv["CI_PROJECT_NAME_WITH_ENV"] = projectName + "-" + env
	// autoenv["CI_REPLICAS"] = "2"

	// autoenv["CI_REGISTRY_IMAGE"] = "image?" // or using project_path

	// autoenv["CI_USER_NAME"] = event.UserName
	// autoenv["CI_USER_EMAIL"] = event.UserEmail
	// autoenv["CI_MSG"] = event.Commits[0].Message
	// autoenv["CI_TIME"] = time.Now().Format("2006-1-2 15:04:05")

	return startBuild(event, nil)
}

type buildOption struct {
	gen    bool
	build  bool
	deploy bool
	// no easy way to delete? why need delete?
}

func startBuild(event Eventer, bo *buildOption) (err error) {
	e, err := event.GetInfo()
	if err != nil {
		err = fmt.Errorf("GetInfo for %q, err: %v", e.Project, err)
		return
	}
	project := e.Project
	branch := e.Branch
	env := e.Branch

	log.Printf("got project %v, branch: %v, env: %v\n", project, branch, env)

	if bo == nil {
		bo = &buildOption{
			gen:    true,
			build:  true,
			deploy: true,
		}
	} else {
		if bo.gen == false && bo.build == false && bo.deploy == false {
			err = fmt.Errorf("nothing to do, gen=build=deploy=false for %q, err: %v", e.Project, err)
			return
		}
	}

	autoenv, err := EventInfoToMap(e)
	if err != nil {
		err = fmt.Errorf("EventInfoToMap for %q, err: %v", project, err)
		return
	}

	for k, v := range autoenv {
		log.Printf("autoenv: %v=%v", k, v)
	}

	// only build for develop branch, need confirm?
	// we shoult not limit the branch, let them easy to change? change in config.yaml, based on tag?
	//so release need to build too? or just add addition condition to build for other branch?

	// what to do with master branch as dev?  init by commit text?

	// if not inited, just using default setting?
	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		err = fmt.Errorf("project: %v, new err: %v", project, err)
		return
	}

	// it should be config from repo or template now
	// if p.DevBranch == "" {
	// 	p.DevBranch = "develop"
	// }

	// check this only for init?
	//
	// this should be check later, by see config first?
	// if I were them, I just do release, let the system figure out when to init?
	// release to test? it's better to init by tag msg?

	if !projectpkg.BranchIsTag(branch) {
		if branch != p.DevBranch { // tag should be release, not build?
			log.Printf("ignore build of branch: %v (devBranch=%q) from project: %v", branch, p.DevBranch, project)
			return
		}

		var init, reinit bool
		if strings.Contains(e.Message, "/init") {
			init = true
		}
		if strings.Contains(e.Message, "/reinit") {
			reinit = true
		}

		// check if force is enabled

		// check if inited, do init by manual trigger?

		// if not inited before, or force is specified, do init now?
		// this will trigger auto build for everyproject? just don't do it?

		// how people trigger init at first place? release text or commit text?

		if !p.Inited() && init {
			err = p.Init()
			if err != nil {
				err = fmt.Errorf("project: %v, init err: %v", project, err)
				return
			}
			log.Printf("inited for project: %v", project)
		}
		if reinit {
			err = p.Init(projectpkg.SetInitForce())
			if err != nil {
				err = fmt.Errorf("project: %v, reinit err: %v", project, err)
				return
			}
			log.Printf("reinited for project: %v", project)
		}
	}
	// do gen if deploy is needed
	if bo.deploy {
		bo.gen = true
	}

	var finalyaml string
	if bo.gen {

		// almost generate everytime, except config
		finalyaml, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv))
		if err != nil {
			err = fmt.Errorf("project: %v, generate before build err: %v", project, err)
			return
		}
		log.Printf("done generate for project: %v", project)
	}

	// everytime build, need generate first

	// write to a auto.env? or
	//envsubst.Eval()

	if bo.build {
		log.Printf("start building for project: %v, branch: %v\n", project, branch)
		out, e := p.Build(project, branch)
		if e != nil {
			err = fmt.Errorf("build err: %v", e)
			return
		}
		fmt.Println("output:", out)
	}
	// check if inited or force provide, if not, init first

	// builded, how to relate the image?

	// have a apply script to do that? passing same tag to it, for the image part?
	// is it need re-generate? provided env is change everytime though

	if bo.deploy {
		ns := autoenv["CI_NAMESPACE"]
		out, e := apply(ns, finalyaml)
		if e != nil {
			err = fmt.Errorf("apply for project: %v, err: %v", project, e)
			return
		}
		log.Printf("apply for %v ok\noutput: %v\n", project, out)
	}
	return
}

func apply(ns, target string) (out string, err error) {
	// check ns or create ns first?
	if ns != "" {
		_, err = projectpkg.CheckOrCreateNamespace(ns)
		if err != nil {
			log.Printf("create namespace %v err: %v\n", ns, err)
		}
		log.Printf("check or create namespace ok\n")
	} else {
		log.Printf("got empty namespace, will not check or create ns before apply\n")
	}

	// auto apply by default?
	return projectpkg.ApplyByKubectl(target, target)
}

const errParseRefs = "parseRefsError"

func parseBranch(refs string) string {
	refss := strings.SplitAfter(refs, "/")
	if len(refss) == 3 {
		return refss[2]
	}
	return errParseRefs
}

// do we really need to handle this, since we will merge to such branch
// say pre branch, then it will trigger build? ( only is only for dev now )
//
// receive tag release, do the build for pre,  or filter based on commit text?
// it should be the same image as test, so no need to build image again? image name is been fixed by build
//
// if project set to auto, we auto tag for master? or just directly
func handleRelease(event *TagPushEvent) (err error) {
	log.Printf("got project %v to build for pre or online env\n", event.Project.PathWithNamespace)

	// autoenv := make(map[string]string)
	// autoenv["PROJECTPATH"] = event.Project.PathWithNamespace
	// autoenv["BRANCH"] = event.Ref
	// autoenv["USERNAME"] = event.UserName
	// autoenv["USEREMAIL"] = event.UserEmail
	// autoenv["MSG"] = event.Message

	// fmt.Println("autoenv:", autoenv)

	// e, err := event.GetInfo()
	// if err != nil {
	// 	err = fmt.Errorf("GetInfo for %q, err: %v", e.Project, err)
	// 	return
	// }

	// project := e.Project
	// branch := e.Branch
	// env := e.Branch

	// if !projectpkg.BranchIsTag(branch) {
	// 	err = fmt.Errorf("project %v release tag format incorrect, should prefix with v, got %v", project, branch)
	// 	return
	// }
	// log.Printf("got project %v, branch: %v, env: %v\n", project, branch, env)

	// autoenv, err := EventInfoToMap(e)
	// if err != nil {
	// 	err = fmt.Errorf("EventInfoToMap for %q, err: %v", project, err)
	// 	return
	// }

	// for k, v := range autoenv {
	// 	log.Printf("autoenv: %v=%v", k, v)
	// }

	// // set branch or set tag?
	// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	// if err != nil {
	// 	err = fmt.Errorf("project: %v, new err: %v", project, err)
	// 	return
	// }

	// // almost generate everytime, except config
	// err = p.Generate(projectpkg.SetGenAutoEnv(autoenv))
	// if err != nil {
	// 	err = fmt.Errorf("project: %v, generate before build err: %v", project, err)
	// 	return
	// }
	// log.Printf("done generate for project: %v", project)

	// // everytime build, need generate first

	// // write to a auto.env? or
	// //envsubst.Eval()

	// log.Printf("start building for project: %v, branch: %v\n", project, branch)
	// out, err := p.Build(project, branch)
	// if err != nil {
	// 	err = fmt.Errorf("build err: %v", err)
	// 	return
	// }
	// fmt.Println("output:", out)

	// ns := autoenv["CI_NAMESPACE"]
	// err = apply(ns, finalyaml)
	// if err != nil {
	// 	err = fmt.Errorf("apply for project: %v, err: %v", project, err)
	// 	return
	// }
	// log.Printf("apply for %v ok\n", project)
	return startBuild(event, nil)
}
