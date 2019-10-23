package harbor

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	harbor "github.com/TimeBye/go-harbor"
	"github.com/parnurzeal/gorequest"
)

var defaultClient *harbor.Client

func Setting(url, user, pass string) {
	defaultClient = harbor.NewClient(nil, url, user, pass)
}

func ListProjects() (ps []harbor.Project, err error) {
	// opt := &harbor.ListProjectsOptions{Name: "wenzhenglin"}
	opt := &harbor.ListProjectsOptions{}
	projects, _, e := defaultClient.Projects.ListProject(opt)
	if e != nil {
		err = fmt.Errorf("list project err %v", e)
		return nil, err
	}
	return projects, nil
}

const TimeLayout = "2006-1-2_15:04:05"

func timeToLocal(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

func parseTimeLocal(ts string) (t time.Time, err error) {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		err = fmt.Errorf("load localtimezone err: %v", err)
		return
	}
	return time.ParseInLocation(TimeLayout, ts, loc)
}

func ListRepoTagLatestName(repo string, ts string) (name string, err error) {
	tag, err := ListRepoTagLatest(repo, ts)
	if err != nil {
		return
	}
	name = tag.Name
	return
}

// get the latest tag before a release
// if ts is empty string, just return latest tag
func ListRepoTagLatest(repo string, ts string) (tag harbor.TagResp, err error) {
	tags, err := ListRepoTags(repo)
	if err != nil {
		return
	}
	if len(tags) == 0 {
		err = fmt.Errorf("no tags found")
		return
	}
	if ts == "" {
		tag = tags[0]
		return
	}
	t, err := parseTimeLocal(ts)
	if err != nil {
		err = fmt.Errorf("parse time err: %v", err)
		return
	}
	log.Printf("parsed ts: %v, to t: %v\n", ts, t)
	for _, v := range tags {
		// fmt.Printf("tag time: %v, t: %v\n", v.Created, t)
		t2, e := timeToLocal(v.Created, "Local")
		if e != nil {
			err = e
			return
		}
		fmt.Printf("tag time: %v, ts: %v\n", t2, t)
		if t2.Before(t) {
			fmt.Printf("got tag time: %v, ts: %v\n", t2, t)
			tag = v
			return
		}
	}
	err = fmt.Errorf("no tags before the time: %v", ts)
	return
}

// func ListRepoThreeTags(repo string) (tags []harbor.TagResp, err error) {
// 	tags, err = ListRepoTags(repo)
// 	if err != nil {
// 		err = fmt.Errorf("ListRepoTags err: %v", err)
// 		return
// 	}
// 	if len(tags) < 3 {
// 		err = fmt.Errorf("not enough images, got %v, expect 3", len(tags))
// 		return
// 	}
// 	tags = tags[:3]
// 	return
// }

// repo is kind like "flow_center/8-yun"
func ListRepoTags(repo string) (tags []harbor.TagResp, err error) {
	log.Printf("list tags for [%s]\n", repo)
	tags, _, e := defaultClient.Repositories.ListRepositoryTags(repo)
	if len(e) != 0 {
		err = fmt.Errorf("ListRepositoryTags got %v err: %v", len(e), e)
		return
	}
	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].Created.After(tags[j].Created)
	})
	return
}

func RepoTagIsExist(repo, tag string) (exist bool, err error) {
	tags, err := ListRepoTags(repo)
	if err != nil {
		return
	}
	for _, v := range tags {
		if v.Name == tag {
			exist = true
			return
		}
	}
	return
}

func CheckProject(name string) (pid int64, err error) {
	// list non-exist project will list all projects, so we find out ourself
	projects, err := ListProjects()
	if err != nil {
		err = fmt.Errorf("list project err %v", err)
		return
	}
	// fmt.Printf("got %v project\n", len(projects))
	for _, v := range projects {
		// fmt.Printf("project id: %v, name: %v\n", v.ProjectID, v.Name)
		if v.Name == name {
			pid = v.ProjectID
			// fmt.Printf("got project %#v\n", v)
			return
		}
	}
	if pid == 0 {
		err = fmt.Errorf("project not found")
	}
	return
}

func CreateProject(name string) (err error) {
	log.Printf("creating harbor project %v", name)
	resp, e := defaultClient.Projects.CreateProject(harbor.ProjectRequest{Name: name})
	if e != nil {
		err = fmt.Errorf("create project err %v", e)
		return
	}
	// fmt.Printf("respcode", resp.StatusCode)
	// spew.Dump("resp", **resp)
	err = ParseResp(resp)
	if err != nil {
		err = fmt.Errorf("create project: %v, failed %v", name, err)
		return
	}
	return
}

func DeleteProject(name string) (err error) {
	log.Printf("deleting harbor project %v", name)

	pid, err := CheckProject(name)
	if err != nil {
		err = fmt.Errorf("check if project exist err: %v", err)
		return
	}

	resp, e := defaultClient.Projects.DeleteProject(pid)
	if e != nil {
		err = fmt.Errorf("create project err %v", e)
		return
	}
	// fmt.Printf("respcode", resp.StatusCode)
	// spew.Dump("resp", **resp)
	err = ParseResp(resp)
	if err != nil {
		err = fmt.Errorf("delete project: %v, failed %v", name, err)
		return
	}
	return
}

func CreateProjectIfNotExist(name string) (err error) {
	_, err = CheckProject(name)
	if err == nil {
		log.Println("project exist")
		return
	}
	return CreateProject(name)
}

// func CheckProject(name string) bool {
// 	resp, err := defaultClient.Projects.CheckProject(name)
// 	pretty("resp", resp)
// 	pretty("err", err)
// 	if err != nil {
// 		return false
// 	}
// 	return true
// }

func pretty(t string, a interface{}) {
	b, _ := json.MarshalIndent(a, "", "  ")
	fmt.Println("pretty", t, string(b))
}

// *gorequest.Response *http.Response
func ParseResp(resp *gorequest.Response) (err error) {
	if resp == nil {
		return fmt.Errorf("empty response")
	}
	r := *resp
	c := r.StatusCode
	if c == 200 || c == 201 {
		return
	}
	return fmt.Errorf("status code: %v", c)
}
