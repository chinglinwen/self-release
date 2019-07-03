package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"wen/self-release/pkg/notify"
	"wen/self-release/pkg/sse"

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
	project := event.Project.PathWithNamespace
	branch := parseBranch(event.Ref)

	log.Printf("got push for project %v to build for test env\n", project)

	b := NewBuilder(project, branch)
	b.log("starting logs")

	return b.startBuild(event, nil)
}

// do we really need to handle this, since we will merge to such branch
// say pre branch, then it will trigger build? ( only is only for dev now )
//
// receive tag release, do the build for pre,  or filter based on commit text?
// it should be the same image as test, so no need to build image again? image name is been fixed by build
//
// if project set to auto, we auto tag for master? or just directly
func handleRelease(event *TagPushEvent) (err error) {
	project := event.Project.PathWithNamespace
	branch := parseBranch(event.Ref)
	log.Printf("got release project %v to build for pre or online env\n", project)

	b := NewBuilder(project, branch)
	b.log("starting logs")
	return b.startBuild(event, nil)
}

type buildOption struct {
	gen      bool
	build    bool
	deploy   bool
	rollback bool
	// no easy way to delete? why need delete?
	nonotify bool
	p        *projectpkg.Project // to avoid re-open or git pull
}

type builder struct {
	*sse.Broker
	// p *projectpkg.Project
	// Event EventInfo // for later modified to restart event
}

// how user send the commands without release: commit? ci trigger, wechat msg?

// try grab the event too, so it can trigger again, or even changed event
func NewBuilder(project, branch string) (b *builder) {
	b = &builder{
		Broker: sse.New(project, branch),
	}
	b.logf("<h1>created log for project: %v</h1>", project)
	return
}

func (b *builder) logf(s string, msgs ...interface{}) {
	msg := fmt.Sprintf(s, msgs...)
	// log.Println(msg)
	fmt.Fprint(b.PWriter, msg)
	// b.Messages <- msg
}

func (b *builder) log(msgs ...interface{}) {
	msg := fmt.Sprint(msgs...)
	// log.Println(msg)
	fmt.Fprint(b.PWriter, msg)
	// b.Messages <- msg
}

func (b *builder) notify(msg, username string) {
	if username == "" {
		log.Printf("username is empty for %v, ignore notify msg: %v\n", b.Project, msg)
		return
	}
	reply, err := notify.Send(username, msg)
	if err != nil {
		log.Printf("send err: %v\nout: %v\n", err, reply)
	}
	log.Println("sended notify to ", username)
	return
}

func (b *builder) startBuild(event Eventer, bo *buildOption) (err error) {
	defer func() {
		if bo.nonotify {
			return
		}
		if err != nil {
			b.notify("build err:\n"+err.Error(), b.Event.UserName)
		} else {
			text := fmt.Sprintf("release for project: %v, branch: %v, env: %v ok", b.Project, b.Branch, b.Event.Env)
			b.notify(text, b.Event.UserName)
		}
	}()
	e, err := event.GetInfo()
	if err != nil {
		err = fmt.Errorf("GetInfo for %q, err: %v", e.Project, err)
		return
	}
	b.Event = e

	// spew.Dump("build event", e)

	project := e.Project
	branch := e.Branch
	env := projectpkg.GetEnvFromBranch(e.Branch)

	// bname := strings.Replace(fmt.Sprintf("%v-%v", project, branch), "/", "-", -1)
	// b := NewBuilder(bname)
	defer b.Close()

	tip := fmt.Sprintf("start build for project %v, branch: %v, env: %v\n", project, branch, env)
	b.logf(tip)

	notifytext := fmt.Sprintf("%vlog url: http://build.newops.haodai.net/logs?key=%v", tip, b.Key)
	b.notify(notifytext, e.UserName)

	if bo == nil {
		bo = &buildOption{
			gen:    true,
			build:  true,
			deploy: true,
		}
	} else {
		if bo.gen == false && bo.build == false && bo.deploy == false && bo.rollback == false {
			err = fmt.Errorf("nothing to do, gen,build,deploy and rollback are false for %q, err: %v", e.Project, err)
			b.log(err)
			return
		}
	}

	// only build for develop branch, need confirm?
	// we shoult not limit the branch, let them easy to change? change in config.yaml, based on tag?
	//so release need to build too? or just add addition condition to build for other branch?

	// what to do with master branch as dev?  init by commit text?

	var p *projectpkg.Project

	if bo.p == nil {
		// if not inited, just using default setting?
		p, err = getproject(project, branch, bo.rollback)
		// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
		if err != nil {
			err = fmt.Errorf("project: %v, new err: %v", project, err)
			b.log(err)
			return
		}
		b.log("clone or open project ok")

	} else {
		p = bo.p
	}

	// if rollback is set, get previous tag as branch
	if bo.rollback {
		// e.Branch = b.p.Branch
		// build already, no need to build again?
		// TODO: what if no build before? let's just build it?
		// detect k8s-online.yaml see if exist and what's the tag?
		bo.gen = true
		bo.deploy = true
		b.log("this is a rollback operation")
	}

	autoenv, err := EventInfoToMap(e)
	if err != nil {
		err = fmt.Errorf("EventInfoToMap for %q, err: %v", project, err)
		b.log(err)
		return
	}

	for k, v := range autoenv {
		log.Printf("autoenv: %v=%v", k, v)
		b.logf("autoenv: %v=%v", k, v)
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

	// skip init push event
	if strings.Contains(e.Message, "init config.yaml") {
		a := fmt.Sprintf("ignore build for project: %v, branch: %v, it's a init project config event", project, branch)
		// log.Println(a)
		b.log(a)
		return
	}

	if !projectpkg.BranchIsTag(branch) {
		if branch != p.DevBranch { // tag should be release, not build?
			a := fmt.Sprintf("ignore build of branch: %v (devBranch=%q) from project: %v", branch, p.DevBranch, project)
			// log.Println(a)
			b.log(a)
			return
		}

		var init, reinit bool
		if strings.Contains(e.Message, "/init") {
			b.log("will do init")
			init = true
		}
		if strings.Contains(e.Message, "/reinit") {
			b.log("will do reinit")
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
				b.log(err)
				return
			}
			// log.Printf("inited for project: %v", project)
			b.logf("inited for project: %v", project)
		}
		if reinit {
			err = p.Init(projectpkg.SetInitForce())
			if err != nil {
				err = fmt.Errorf("project: %v, reinit err: %v", project, err)
				b.log(err)
				return
			}
			b.logf("reinited for project: %v", project)
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
			b.log(err)
			return
		}
		b.logf("done generate for project: %v", project)
	}

	// everytime build, need generate first

	// write to a auto.env? or
	//envsubst.Eval()

	if bo.build {
		// out := make(chan string, 10)

		b.logf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)
		out, e := p.Build(project, branch, env)
		// e := p.Build(project, branch, env, out)
		if e != nil {
			err = fmt.Errorf("build err: %v", e)
			b.log(err)
			return
		}
		b.log("<h2>docker build outputs:</h2>")
		for text := range out {
			b.log(text)
		}
		// scanner := bufio.NewScanner(strings.NewReader(out))
		// // scanner.Split(bufio.ScanLines)
		// for scanner.Scan() {
		// 	b.log(scanner.Text())
		// }

		// for v := range out {
		// 	b.log("output:", v)
		// }
		b.log("build is done.")
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
			b.log(err)
			return
		}
		// b.logf("apply for %v ok\n<h2>k8s apply output:</h2>%v\n", project, out)
		b.log("<h2>k8s apply</h2>")
		b.logf("apply for %v ok\n", project)
		b.logf("k8s apply output:%v\n", out)
	}

	b.logf("<hr>end at %v .", time.Now().Format(TimeLayout))
	return
}

func getproject(project, branch string, rollback bool) (p *projectpkg.Project, err error) {
	p, err = projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		return
	}
	if rollback {
		argBranch := branch
		branch, err = p.GetRepo().GetPreviousTag() // rollback before specific tag, just redeploy then?
		if err != nil {
			err = fmt.Errorf("GetPreviousTag err: %v", err)
			return
		}

		if argBranch != branch {
			// re-open the branch
			p, err = projectpkg.NewProject(project, projectpkg.SetBranch(branch))
			if err != nil {
				err = fmt.Errorf("project: %v, branch: %v, new(rollback ) err: %v", project, branch, err)
				return
			}
		}
		p.Branch = branch
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
