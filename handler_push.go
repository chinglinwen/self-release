package main

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"wen/self-release/git"
	"wen/self-release/pkg/notify"
	"wen/self-release/pkg/sse"

	"github.com/chinglinwen/log"
	"github.com/k0kubun/pp"

	projectpkg "wen/self-release/project"
)

const ingressSuffix = "newops.haodai.net"

// get autoenv from events

// const DevBranch = "develop"

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

	log.Debug.Println("try lock for project", project)
	err = sse.Lock(project, branch)
	if err != nil {
		return
	}
	defer sse.UnLock(project, branch)

	log.Debug.Println("start new builder for project", project)
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

	err = sse.Lock(project, branch)
	if err != nil {
		return
	}
	defer sse.UnLock(project, branch)

	log.Debug.Println("start new builder for project", project)

	b := NewBuilder(project, branch)
	b.log("starting logs")
	return b.startBuild(event, nil)
}

type buildOption struct {
	gen        bool
	nobuild    bool
	force      bool
	buildimage bool
	deploy     bool
	rollback   bool
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
	log.Debug.Println("creating builder for project", project)
	b = &builder{
		Broker: sse.New(project, branch),
	}
	b.logf("<h1>created log for project: %v</h1>", project)
	return
}

func (b *builder) logf(s string, msgs ...interface{}) {
	msg := fmt.Sprintf(s, msgs...)
	// log.Println(msg)
	b.write(msg)
	// b.Messages <- msg
}

func (b *builder) log(msgs ...interface{}) {
	msg := fmt.Sprint(msgs...)
	// log.Println(msg)
	b.write(msg)
	// b.Messages <- msg
}

func (b *builder) logerr(msgs ...interface{}) {
	msg := fmt.Sprint(msgs...)
	log.Println(msg)
	b.write(msg)
	// b.Messages <- msg
}

func (b *builder) write(msg string) {
	if !checkIsHeader(msg) {
		msg += "\n"
	}
	fmt.Fprint(b.PWriter, msg)
}
func checkIsHeader(text string) bool {
	return regexp.MustCompile(`<h.+</h`).MatchString(text)
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
	e, err := event.GetInfo()
	if err != nil {
		err = fmt.Errorf("GetInfo for %v, err: %v", e.Project, err)
		return
	}
	b.Event = e

	// // check if from harbor
	// fromHarbor := strings.Contains(e.Message, "harbor")

	// if e.CommitID == "" && !fromHarbor {
	// 	err = fmt.Errorf("commit id is empty for %v", e.Project)
	// 	return
	// }
	pp.Printf("commitid: %v\n", e.CommitID)
	// spew.Dump("build event", e)

	project := e.Project
	branch := e.Branch
	// from gitlab true
	env := projectpkg.GetEnvFromBranchOrCommitID(e.Project, e.Branch, true)
	commitid := e.CommitID

	// bname := strings.Replace(fmt.Sprintf("%v-%v", project, branch), "/", "-", -1)
	// b := NewBuilder(bname)
	// defer b.Close()

	tip := fmt.Sprintf("start build for project %v, branch: %v, env: %v\n", project, branch, env)
	b.logf(tip)

	log.Debug.Printf(tip)

	notifytext := fmt.Sprintf("%vlog url: %v/logs?key=%v", tip, *selfURL, b.Key)
	b.notify(notifytext, e.UserName)

	defer func() {
		log.Debug.Printf("try close broker now\n")
		b.Close()
		log.Debug.Printf("try close broker ok\n")
		if bo != nil && bo.nonotify {
			return
		}
		if err != nil {
			b.notify("build err:\n"+err.Error(), b.Event.UserName)
		} else {
			url := getProjectURL(project, env)
			text := fmt.Sprintf("release for project: %v, branch: %v, env: %v ok\n项目访问地址: %v", b.Project, b.Branch, env, url)
			b.notify(text, b.Event.UserName)

		}
		log.Debug.Printf("exit startBuild now\n")
	}()

	// check permission
	err = git.CheckPerm(project, e.UserName, env)
	if err != nil {
		err = fmt.Errorf("check permission for %q, user: %v, err: %v", project, e.UserName, err)
		return
	}
	log.Debug.Printf("check permission for %q, user: %v ok\n", project, e.UserName)

	if bo == nil {
		bo = &buildOption{
			// gen: true,
			// build:  true,
			deploy: true,
		}
	} else {
		if bo.deploy == false {
			err = fmt.Errorf("nothing to do, gen,build,deploy and rollback are false for %q, err: %v", e.Project, err)
			b.logerr(err)
			return
		}
	}

	// only build for develop branch, need confirm?
	// we shoult not limit the branch, let them easy to change? change in config.yaml, based on tag?
	//so release need to build too? or just add addition condition to build for other branch?

	// what to do with master branch as dev?  init by commit text?

	var p *projectpkg.Project

	// since we have ui, let's ignore here

	// // create project need to distinguish if it's a init
	// var init, forceinit bool
	// if !projectpkg.BranchIsTag(branch) {
	// 	if strings.Contains(e.Message, "/helpdocker") {
	// 		b.log("will do init")
	// 		init = true
	// 	}
	// 	if strings.Contains(e.Message, "/forcehelpdocker") {
	// 		b.log("will do forceinit")
	// 		forceinit = true
	// 	}
	// }

	if bo.p == nil {
		// if not inited, just using default setting?
		p, err = getproject(project, branch)
		// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
		if err != nil {
			err = fmt.Errorf("get project: %v, err: %v", project, err)
			b.logerr(err)
			return
		}
		// b.log("clone or open project ok")
	} else {
		p = bo.p
	}

	// branch is not tag
	// TODO(wen): build image only for test?
	if env == projectpkg.TEST {
		if branch != p.Config.S.DevBranch { // tag should be release, not build?
			err = fmt.Errorf("ignore build of branch: %v (devBranch=%q) from project: %v", branch, p.Config.S.DevBranch, project)
			// log.Println(a)
			b.log(err)
			return
		}
		// check if force is enabled
		// check if inited, do init by manual trigger?
		// if not inited before, or force is specified, do init now?
		// this will trigger auto build for everyproject? just don't do it?

		// how people trigger init at first place? release text or commit text?

		// if !p.Inited() && init {
		// 	b.log("<h2>Init project</h2>")
		// 	err = p.Init()
		// 	if err != nil {
		// 		err = fmt.Errorf("project: %v, init err: %v", project, err)
		// 		b.logerr(err)
		// 		return
		// 	}
		// 	// log.Printf("inited for project: %v", project)
		// 	b.logf("inited for project: %v", project)
		// 	// return // return for init operation?
		// }
		// if forceinit {
		// 	b.log("<h2>Force Init project</h2>")
		// 	err = p.Init(projectpkg.SetInitForce())
		// 	if err != nil {
		// 		err = fmt.Errorf("project: %v, forceinit err: %v", project, err)
		// 		b.logerr(err)
		// 		return
		// 	}
		// 	b.logf("forceinit for project: %v ok", project)
		// 	// return // return for init operation?
		// }
	}

	// // TODO: not support yet, if rollback is set, get previous tag as branch
	// if bo.rollback {
	// 	// e.Branch = b.p.Branch
	// 	// build already, no need to build again?
	// 	// TODO: what if no build before? let's just build it?
	// 	// detect k8s-online.yaml see if exist and what's the tag?
	// 	bo.gen = true
	// 	bo.deploy = true
	// 	b.log("this is a rollback operation")
	// }

	// envmap, err := EventInfoToMap(e)
	// if err != nil {
	// 	err = fmt.Errorf("EventInfoToMap for %q, err: %v", project, err)
	// 	b.logerr(err)
	// 	return
	// }

	// mergenote, envMap, err := p.ReadEnvs(autoenv)
	// if err != nil {
	// 	// err = fmt.Errorf("readenvs err: %v", err)
	// 	log.Printf("readenvs err: %v, will ignore\n", err)
	// 	// envMap = make(map[string]string)
	// }
	// else {
	// 	log.Printf("merged envs from config.env to autoenv: \n%v", mergenote)
	// }

	b.log("<h2>Info</h2>")

	ebytes, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		err = fmt.Errorf("marshal event to json for %v err: %v", project, err)
		b.logerr(err)
		return
	}

	eventstr := strings.ReplaceAll(html.EscapeString(string(ebytes)), "\n", "<br>")
	b.logf("<pre>%v</pre>", eventstr)

	// for k, v := range envmap {
	// 	log.Printf("env: %v = %q\n", k, v)
	// 	b.logf("%v = %q\n", k, v)
	// }

	// for _, v := range mergenote {
	// 	log.Print(v)
	// 	b.log(v)
	// }

	// it should be config from repo or template now
	// if p.DevBranch == "" {
	// 	p.DevBranch = "develop"
	// }

	// check this only for init?
	//
	// this should be check later, by see config first?
	// if I were them, I just do release, let the system figure out when to init?
	// release to test? it's better to init by tag msg?

	// // skip init push event
	// if strings.Contains(e.Message, "init config.yaml") {
	// 	a := fmt.Sprintf("ignore build for project: %v, branch: %v, it's a init project config event", project, branch)
	// 	// log.Println(a)
	// 	b.log(a)
	// 	return
	// }

	// do gen if deploy is needed
	// if bo.deploy {
	// 	bo.gen = true
	// }

	// var finalyaml string
	// if bo.gen {
	// 	b.log("<h2>Generate k8s yaml</h2>")

	// 	// almost generate everytime, except config
	// 	finalyaml, err = p.Generate(projectpkg.SetGenAutoEnv(envMap), projectpkg.SetGenEnv(env))
	// 	if err != nil {
	// 		err = fmt.Errorf("project: %v, generate before build err: %v", project, err)
	// 		b.logerr(err)
	// 		return
	// 	}
	// 	b.logf("done generate for project: %v", project)
	// }

	// everytime build, need generate first

	// write to a auto.env? or
	//envsubst.Eval()

	b.log("<h2>Docker build</h2>")

	// is devbranch, or tag not exist yet

	imageexist, needbuild := p.NeedBuild(commitid)

	if ((!bo.nobuild) && needbuild) || bo.buildimage {
		// out := make(chan string, 10)

		b.logf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)
		out, e := p.Build(project, branch, env, commitid)
		// e := p.Build(project, branch, env, out)
		if e != nil {
			err = fmt.Errorf("build err: %v", e)
			b.logerr(err)
			return
		}
		b.log("docker build outputs:<br>")

		// some error not retuned, so let's detect it
		detector := "digest: sha256"
		var buildSuccess bool
		for text := range out {
			if strings.Contains(text, detector) {
				buildSuccess = true
			}
			b.log(text)
		}
		// build need to check image to see if it success, or parse log?

		log.Println("done of receiving build outputs")

		if buildSuccess {
			b.log("build is ok.")
		} else {
			err = fmt.Errorf("build is failed, maybe internal error, or build-script error.")
			b.logerr(err)
			return
		}
	} else {
		b.logf("will not build, for flags:")
		b.logf("runtime options: nobuild: %v", bo.nobuild)
		b.logf("runtime options: buildimage: %v", bo.buildimage)

		b.logf("config buildmode: %v", p.Config.S.BuildMode)
		b.logf("needbuild detect result: %v", needbuild)
		b.logf("imageexist check result: %v", imageexist)
	}
	b.log("<h2>K8s project</h2>")

	if bo.deploy {
		var yamlbody, out string
		yamlbody, out, err = applyReleaseFromEvent(e)
		if err != nil {
			err = fmt.Errorf("create k8s release for project: %v, branch: %v, err: %v", project, branch, err)
			b.logerr(err)
			return
		}
		log.Printf("create release ok, out: %v", out)
		outyaml := strings.ReplaceAll(html.EscapeString(yamlbody), "\n", "<br>")
		b.logf("created project yaml: <pre>%v</pre>", outyaml)
		b.logf("apply output:")
		b.logf("%v", out)
		b.log("<br>")
	} else {
		err = fmt.Errorf("deploy flag not set, skip.")
		b.logerr(err)
	}

	b.logf("<hr>end at %v .", time.Now().Format(TimeLayout))
	return
}

func getProjectURL(project, env string) string {
	project = strings.Replace(project, "/", "-", -1)
	return fmt.Sprintf("https://%v-%v.%v", project, env, ingressSuffix)
}

func getproject(project, branch string) (p *projectpkg.Project, err error) {
	return projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	// p, err = projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	// if err != nil {
	// 	return
	// }
	// if rollback {
	// 	argBranch := branch
	// 	branch, err = p.GetPreviousTag() // rollback before specific tag, just redeploy then?
	// 	if err != nil {
	// 		err = fmt.Errorf("get previous tag err: %v", err)
	// 		return
	// 	}

	// 	if argBranch != branch {
	// 		// re-open the branch
	// 		p, err = projectpkg.NewProject(project, projectpkg.SetBranch(branch), projectpkg.SetNoEnableCheck(tr))
	// 		if err != nil {
	// 			err = fmt.Errorf("project: %v, branch: %v, new(rollback ) err: %v", project, branch, err)
	// 			return
	// 		}
	// 	}
	// 	p.Branch = branch
	// }
	// return
}

// func apply(ns, target string) (out string, err error) {
// 	// check ns or create ns first?
// 	if ns != "" {
// 		_, err = projectpkg.CheckOrCreateNamespace(ns)
// 		if err != nil {
// 			log.Printf("create namespace %v err: %v\n", ns, err)
// 		}
// 		log.Printf("check or create namespace ok\n")
// 	} else {
// 		log.Printf("got empty namespace, will not check or create ns before apply\n")
// 	}

// 	// auto apply by default?
// 	return projectpkg.ApplyByKubectl(target)
// }

const errParseRefs = "parseRefsError"

func parseBranch(refs string) string {
	refss := strings.SplitAfter(refs, "/")
	if len(refss) == 3 {
		return refss[2]
	}
	return errParseRefs
}
