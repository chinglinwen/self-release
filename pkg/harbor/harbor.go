package harbor

import (
	"encoding/json"
	"fmt"
	"log"

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

// repo is kind like "flow_center/8-yun"
func ListRepoTags(repo string) (tags []harbor.TagResp, err error) {
	tags, _, e := defaultClient.Repositories.ListRepositoryTags(repo)
	if len(e) != 0 {
		err = fmt.Errorf("ListRepositoryTags err: %v", e)
		return
	}
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

func CheckProject(name string) (ok bool, err error) {
	// list non-exist project will list all projects, so we find out ourself
	projects, err := ListProjects()
	if err != nil {
		err = fmt.Errorf("list project err %v", err)
		return
	}
	for _, v := range projects {
		if v.Name == name {
			ok = true
			return
		}
	}
	return
}

// not working for now
func CreateProject(name string) (err error) {
	log.Printf("creating harbor project %v", name)
	resp, e := defaultClient.Projects.CreateProject(harbor.ProjectRequest{Name: name})
	if e != nil {
		err = fmt.Errorf("create project err %v", e)
		return
	}
	// fmt.Printf("respcode", resp.StatusCode)
	// spew.Dump("resp", **resp)
	return ParseResp(resp)
}

func CreateProjectIfNotExist(name string) (created bool, err error) {
	exist, err := CheckProject(name)
	if err != nil {
		return
	}
	if !exist {
		created = true
		return false, CreateProject(name)
	}
	return
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
		return fmt.Errorf("request failed, empty response")
	}
	r := *resp
	if r.StatusCode > 200 {
		return fmt.Errorf("request failed, status: %v", r.Status)
	}
	return nil
}
