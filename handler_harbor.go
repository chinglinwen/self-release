package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wen/self-release/pkg/sse"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	cache "github.com/patrickmn/go-cache"
)

// do build if buildmode is manual
func HarborToDeploy(i *HarborEventInfo) (err error) {
	log.Printf("got push from: %v, project: %v:%v\n", i.Name, i.Project, i.Tag)
	project, tag, name := i.Project, i.Tag, i.Name

	d, found := getCache(i)
	if found {
		log.Printf("ignore cached event for project: %v:%v, expire in: %v\n", project, tag, d.Format(TimeLayout))
		return
	}
	setCache(i)
	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(tag), projectpkg.SetConfigMustExist(true))
	if err != nil {
		err = fmt.Errorf("project: %v:%v, new err: %v", project, tag, err)
		return
	}
	if !p.IsEnabled() {
		log.Printf("ignore build for project: %v:%v, project not enabled\n", project, tag)
		return
	}
	log.Printf("start deploy from harbor for project: %v:%v\n", project, tag)

	// bo := &buildOption{
	// 	gen: true,
	// 	// nobuild:    f.nobuild,
	// 	buildimage: false,
	// 	deploy:     true,
	// 	// nonotify:   true,
	// 	p: p,
	// }
	e := &sse.EventInfo{
		Project: project,
		Branch:  tag,
		// Env:       env, // default derive from branch
		UserName: name,
		// UserEmail: useremail,
		Message: fmt.Sprintf("[from harbor] %v", name),
	}
	log.Printf("ignore harbor event for now, it has duplicate evenets\n")
	return

	// at least update the time to togger the change
	yamlbody, out, err := applyReleaseFromEvent(e)
	if err != nil {
		err = fmt.Errorf("create k8s project resource for project: %v, branch: %v, err: %v", project, tag, err)
		log.Println(err)
		log.Printf("yamlbody: %v\n", yamlbody)
		return
	}

	log.Printf("created release ok, out: %v\n", out)

	// b := NewBuilder(project, branch)
	// b.log("starting logs trigger by harbor image push, buildmode: manual")

	// err = b.startBuild(e, bo)
	// if err != nil {
	// 	err = fmt.Errorf("startdeploy for project: %v, branch: %v, err: %v", project, branch, err)
	// 	log.Println(err)
	// 	return
	// }
	return
}

func applyReleaseFromEvent(e *sse.EventInfo) (yamlbody, out string, err error) {
	pp.Printf("try apply %v\n", e)
	yamlbody, err = EventInfoToProjectYaml(e)
	if err != nil {
		err = fmt.Errorf("convert event to yaml err: %v", err)
		return
	}
	// pp.Printf("yaml %v\n", r)
	out, err = projectpkg.ApplyByKubectlWithString(yamlbody)
	return
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

func harborHandler(c echo.Context) (err error) {
	//may do redirect later?
	r := c.Request()
	body, err := readbody(r)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		E(http.StatusBadRequest, err.Error(), "failed")
		return
	}
	// fmt.Printf("r: %#v\n", r)
	// log.Printf("body: %v", body)

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

	e *HarborEvent
}

func (e *HarborEvent) HarborEventInfo() (i *HarborEventInfo) {
	i = &HarborEventInfo{
		Tag:     e.Events[0].Target.Tag,
		Name:    e.Events[0].Actor.Name,
		Project: e.Events[0].Target.Repository,
		IP:      e.Events[0].Request.Addr,
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
