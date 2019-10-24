package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	cache "github.com/patrickmn/go-cache"
)

func harborHandler(c echo.Context) (err error) {
	r := c.Request()
	body, err := readbody(r)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		E(http.StatusBadRequest, err.Error(), "failed")
		return
	}

	i, err := getHarborEventInfo(body)
	if err != nil {
		log.Printf("body: %v", body)
		err = fmt.Errorf("get event info err: %v", err)
		E(http.StatusBadRequest, err.Error(), "failed")
		return
	}
	out, err := prettyjson.Marshal(i)
	if err != nil {
		err = fmt.Errorf("marshal e err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	if !i.e.IsPush() {
		c.JSONPretty(http.StatusOK, E(0, "ignore non push event", "ok"), " ")
		return
	}
	log.Printf("got push: %s\n\n\n", out)

	err = HarborToDeploy(i)
	if err != nil {
		err = fmt.Errorf("harbor deploy err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, E(0, "push event handle ok", "ok"), " ")
}

// do build if buildmode is manual
// anything comes to harbor is actually test env
func HarborToDeploy(e *HarborEventInfo) (err error) {
	log.Printf("got push from: %v, project: %v:%v\n", e.Name, e.Project, e.Tag)
	project, tag := e.Project, e.Tag

	d, found := getCache(e)
	if found {
		log.Printf("ignore cached event for project: %v:%v, expire in: %v\n", project, tag, d.Format(TimeLayout))
		return
	}
	setCache(e)
	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(tag), projectpkg.SetConfigMustExist(true))
	if err != nil {
		err = fmt.Errorf("project: %v:%v, new err: %v", project, tag, err)
		return
	}
	// if auto build, will not hit this
	// if manual mode, it need to hit this
	if !p.IsEnabled() || !p.IsManual() {
		log.Printf("ignore build for project: %v:%v, project not enabled or buildmode is not manual \n", project, tag)
		return
	}
	log.Printf("start deploy from harbor for project: %v:%v\n", project, tag)

	// can we set commit status? where to get the commitid
	// only if people use correct tag, if use time as tag, there's no correct time

	// at least update the time to togger the change
	yamlbody, out, err := e.applyReleaseFromEvent()
	if err != nil {
		err = fmt.Errorf("create k8s project resource for project: %v, branch: %v, err: %v", project, tag, err)
		log.Println(err)
		log.Printf("yamlbody: %v\n", yamlbody)
		return
	}

	log.Printf("created release ok, out: %v\n", out)

	return
}

func (e *HarborEventInfo) applyReleaseFromEvent() (yamlbody, out string, err error) {
	pp.Printf("try apply %v\n", e)
	yamlbody, err = e.ToProjectYaml()
	if err != nil {
		err = fmt.Errorf("convert event to yaml err: %v", err)
		return
	}
	out, err = applyRelease(yamlbody)
	return
}

func applyRelease(yamlbody string) (out string, err error) {
	return projectpkg.ApplyByKubectlWithString(yamlbody)
}

func deleteRelease(yamlbody string) (out string, err error) {
	return projectpkg.DeleteByKubectlWithString(yamlbody)
}

// name need to be different for different env
// we don't know the email from harbor
var projectYamlTmpl = `
apiVersion: project.haodai.com/v1alpha1
kind: Project
metadata:
  name: %v-%v
  namespace: %v
spec:
  version: "%v"
  userName: "%v"
  userEmail: "%v"
  releaseMessage: "%v"
  releaseAt: "%v"
`

type projectYaml struct {
	name    string
	env     string
	ns      string
	version string
	user    string
	mail    string
	msg     string
	time    string
}

func (p *projectYaml) validate() (err error) {
	if p.name == "" {
		return fmt.Errorf("project is empty")
	}
	if p.ns == "" {
		return fmt.Errorf("namespace is empty")
	}
	if p.version == "" {
		return fmt.Errorf("branch or commitid is empty")
	}
	if p.env == "" {
		return fmt.Errorf("env is empty")
	}
	if p.user == "" {
		return fmt.Errorf("user or commitid is empty")
	}
	// ignored, msg, time, and email
	return
}

func (p *projectYaml) ToProjectYamlSkipValidate() (body string) {
	return p.toProjectYaml()
}

func (p *projectYaml) ToProjectYaml() (body string, err error) {
	if err = p.validate(); err != nil {
		return
	}
	body = p.toProjectYaml()
	return
}

func (p *projectYaml) toProjectYaml() (body string) {
	time := time.Now().Format(TimeLayout)
	log.Printf("construct yaml: project: %v, env: %v, version: %v\n", p.name, p.env, p.version)
	body = fmt.Sprintf(projectYamlTmpl, p.name, p.env, p.ns, p.version,
		p.user, p.mail, p.msg, time)
	return
}

func (e *HarborEventInfo) ToProjectYaml() (body string, err error) {
	ns, name, err := projectpkg.GetProjectName(e.Project)
	if err != nil {
		err = fmt.Errorf("parse project name for %q, err: %v", e.Project, err)
		return
	}

	// all event from harbor is test env
	env := projectpkg.TEST
	log.Printf("got env: %v for %v:%v\n", env, e.Project, e.Tag)

	version := e.Tag

	msg := fmt.Sprintf("[from harbor] %v", name)

	log.Printf("construct yaml: project: %v, env: %v, version: %v\n", e.Project, env, version)

	// convert info to version?
	p := projectYaml{
		name:    name,
		env:     env,
		ns:      ns,
		version: version,
		user:    e.User,
		msg:     msg,
	}
	return p.ToProjectYaml()
}

var C = cache.New(1*time.Minute, 1*time.Minute)

func setCache(i *HarborEventInfo) {
	key := fmt.Sprintf("%v:%v-%v", i.Project, i.Tag, i.Name)
	d, _ := time.ParseDuration("1m")
	C.Set(key, i, d)
}

func getCache(i *HarborEventInfo) (d time.Time, found bool) {
	key := fmt.Sprintf("%v:%v-%v", i.Project, i.Tag, i.Name)
	_, d, found = C.GetWithExpiration(key)
	return
}

func readbody(r *http.Request) (body string, err error) {
	if r.Body != nil {
		var buf bytes.Buffer
		_, err = buf.ReadFrom(r.Body)
		body = buf.String()
	}
	return
}

func unmarshalHarborEvent(body string) (e *HarborEvent, err error) {
	e = &HarborEvent{}
	err = json.Unmarshal([]byte(body), e)
	if err != nil {
		return
	}
	if len(e.Events) == 0 {
		err = fmt.Errorf("no event found")
		return
	}
	return
}

func getHarborEventInfo(body string) (i *HarborEventInfo, err error) {
	e, err := unmarshalHarborEvent(body)
	if err != nil {
		return
	}
	i = e.HarborEventInfo()
	return
}

func (e HarborEvent) IsPush() bool {
	if e.Events[0].Action != "push" {
		return false
	}
	if !strings.Contains(e.Events[0].Target.MediaType, "manifest") {
		return false
	}
	if e.Events[0].Actor.Name == "devuser" || e.Events[0].Actor.Name == "harbor-ui" {
		return false
	}
	return true
}

type HarborEventInfo struct {
	Name    string
	IP      string
	Project string
	Tag     string
	User    string

	e *HarborEvent
}

func (e *HarborEvent) HarborEventInfo() (i *HarborEventInfo) {
	i = &HarborEventInfo{
		Tag:     e.Events[0].Target.Tag,
		Name:    e.Events[0].Actor.Name,
		Project: e.Events[0].Target.Repository,
		IP:      e.Events[0].Request.Addr,
		User:    e.Events[0].Actor.Name,
		e:       e,
	}
	return
}

type HarborEvent struct {
	Events []struct {
		ID        string    `json:"id"`
		Timestamp time.Time `json:"timestamp"`
		Action    string    `json:"action"`
		Target    struct {
			MediaType  string `json:"mediaType"`
			Size       int    `json:"size"`
			Digest     string `json:"digest"`
			Length     int    `json:"length"`
			Repository string `json:"repository"`
			URL        string `json:"url"`
			Tag        string `json:"tag"`
		} `json:"target"`
		Request struct {
			ID        string `json:"id"`
			Addr      string `json:"addr"`
			Host      string `json:"host"`
			Method    string `json:"method"`
			Useragent string `json:"useragent"`
		} `json:"request"`
		Actor struct {
			Name string `json:"name"`
		} `json:"actor"`
		Source struct {
			Addr       string `json:"addr"`
			InstanceID string `json:"instanceID"`
		} `json:"source"`
	} `json:"events"`
}
