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
	"google.golang.org/grpc/status"

	projectpkg "wen/self-release/project"
)

const ingressSuffix = "newops.haodai.net"

// convert event to build for project
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

	nonotify bool
	p        *projectpkg.Project // to avoid re-open or git pull
}

type builder struct {
	*sse.Broker
	// p *projectpkg.Project
	// Event EventInfo // for later modified to restart event
}

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
	b.write(msg)
}

func (b *builder) log(msgs ...interface{}) {
	msg := fmt.Sprint(msgs...)
	b.write(msg)
}

func (b *builder) logerr(msgs ...interface{}) {
	msg := fmt.Sprint(msgs...)
	log.Println(msg)
	b.write("\n" + msg)
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

func validateRequest(project, branch, env, commitid string) (err error) {
	if project == "" {
		return fmt.Errorf("project is empty")
	}
	if branch == "" {
		return fmt.Errorf("branch is empty")
	}
	if env == "" {
		return fmt.Errorf("env is empty")
	}
	if commitid == "" {
		return fmt.Errorf("commitid is empty")
	}
	return
}

func (b *builder) startBuild(event Eventer, bo *buildOption) (err error) {
	e, err := event.GetInfo()
	if err != nil {
		err = fmt.Errorf("GetInfo for %v, err: %v", e.Project, err)
		return
	}
	b.Event = e

	project := e.Project
	branch := e.Branch

	env := projectpkg.TEST
	if e.EventType == string(tagEventType) {
		env = projectpkg.GetEnvFromBranch(e.Project, e.Branch)
		log.Printf("got env from branch: %v\n", env)
	}
	commitid := e.CommitID
	if commitid == "" {
		var e error
		// build from wechat need to fetch commitid
		commitid, e = git.GetCommitIDFromTag(project, branch)
		if e != nil {
			log.Printf("get commitid from tag for %v err: %v\n", project, branch)
		}
	}

	if err = validateRequest(project, branch, env, commitid); err != nil {
		return
	}
	logurl := fmt.Sprintf("%v/logs?key=%v", selfURL, b.Key)

	tip := fmt.Sprintf("start build for project %v, branch: %v, env: %v, commitid: %v\n",
		project, branch, env, commitid)
	b.logf(tip)

	log.Debug.Printf(tip)

	p, err := getproject(project, branch)
	if err != nil {
		err = fmt.Errorf("get project: %v, err: %v", project, err)
		b.logerr(err)
		return
	}

	// notify only if project enabled
	notifytext := fmt.Sprintf("%vlog url: %v", tip, logurl)
	b.notify(notifytext, e.UserName)

	defer func() {
		log.Debug.Printf("try close broker now\n")
		b.Close()
		log.Debug.Printf("try close broker ok\n")
		if err != nil {
			if bo != nil && bo.nonotify {
				return
			}
			b.notify("build err:\n"+err.Error(), b.Event.UserName)

		} else {
			url := getProjectURL(project, env)
			text := fmt.Sprintf("release for project: %v, branch: %v, env: %v ok\n项目访问地址: %v", b.Project, b.Branch, env, url)
			b.notify(text, b.Event.UserName)
		}
		log.Debug.Printf("exit startBuild now\n")
	}()

	go func() {
		_, err := git.SetCommitStatusRunning(project, commitid, logurl)
		if err != nil {
			log.Println("SetCommitStatusRunning err: ", err)
		}
	}()

	defer func() {
		if err != nil {
			_, err := git.SetCommitStatusFailed(project, commitid, logurl)
			if err != nil {
				log.Println("SetCommitStatusFailed err: ", err)
			}
		} else {
			_, err := git.SetCommitStatusSuccess(project, commitid, logurl)
			if err != nil {
				log.Println("SetCommitStatusSuccess err: ", err)
			}
		}
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

	// // TODO: not support yet, if rollback is set, get previous tag as branch

	b.log("<h2>Info</h2>")

	ebytes, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		err = fmt.Errorf("marshal event to json for %v err: %v", project, err)
		b.logerr(err)
		return
	}

	eventstr := strings.ReplaceAll(html.EscapeString(string(ebytes)), "\n", "<br>")
	b.logf("<pre>%v</pre>", eventstr)

	b.log("<h2>Docker build</h2>")

	var buildSuccess bool

	// build only for test env
	if env == projectpkg.TEST {
		if branch != p.Config.S.DevBranch { // tag should be release, not build?
			err = fmt.Errorf("ignore build of branch: %v (devBranch=%q) from project: %v",
				branch, p.Config.S.DevBranch, project)
			b.log(err)
			return
		}

		// is devbranch, or tag not exist yet
		_, needbuild := p.NeedBuild(commitid)

		if ((!bo.nobuild) && needbuild) || bo.buildimage {
			// out := make(chan string, 10)

			b.logf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)
			err = p.Build(env, commitid)
			// e := p.Build(project, branch, env, out)
			if err != nil {
				err = fmt.Errorf("build err: %v", e)
				b.logerr(err)
				return
			}
			b.log("docker build outputs:<br>")

			// some error not retuned, so let's detect it
			detector := "digest: sha256"

			out, e := p.GetBuildOutput()
			if e != nil {
				err = fmt.Errorf("build getoutput err: %v", e)
				b.logerr(err)
				return
			}

			for {
				text, ok := <-out
				if !ok {
					break
				}
				if strings.Contains(text, detector) {
					buildSuccess = true
				}
				b.log(text)
			}

			// to know if err happen
			if e := p.GetBuildError(); e != nil {
				if st, ok := status.FromError(e); ok {
					err = fmt.Errorf("build got err: %v", st.Message())
				} else {
					err = fmt.Errorf("build got err: %v", e)
				}
				b.logerr(err)
				return
			}

			// build need to check image to see if it success, or parse log?

			log.Println("done of receiving build outputs")

			if buildSuccess {
				b.log("build is ok.")
			} else {
				err = fmt.Errorf("no keyword digest in build logs, so build is failed")
				b.logerr(err)
				return
			}
		} else {
			b.logf("will not build, for flags:")
			b.logf("runtime options: nobuild: %v", bo.nobuild)
			b.logf("runtime options: buildimage: %v", bo.buildimage)

			b.logf("config buildmode: %v", p.Config.S.BuildMode)
			b.logf("needbuild detect result: %v", needbuild)
			// b.logf("imageexist check result: %v", imageexist)
		}
	}

	if env == projectpkg.TEST && !buildSuccess {
		err = fmt.Errorf("build not success, skip deploy.")
		b.logerr(err)
		return
	} else {
		b.log("for time reducing and for consistency(single image for all env)")
		b.log("re-using test build image, so will not build image for pre or online env")
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

func getProjectURL(project, env string) (url string) {
	project = strings.Replace(project, "/", "-", -1)
	if env == projectpkg.ONLINE {
		return fmt.Sprintf("https://%v.%v", project, ingressSuffix)
	} else {
		return fmt.Sprintf("https://%v-%v.%v", project, env, ingressSuffix)
	}
}

func getproject(project, branch string) (p *projectpkg.Project, err error) {
	return projectpkg.NewProject(project, projectpkg.SetBranch(branch))
}

const errParseRefs = "parseRefsError"

func parseBranch(refs string) string {
	refss := strings.SplitAfter(refs, "/")
	if len(refss) == 3 {
		return refss[2]
	}
	return errParseRefs
}
