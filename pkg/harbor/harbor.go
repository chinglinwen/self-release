package harbor

import (
	"encoding/json"
	"fmt"
	"log"

	harbor "github.com/TimeBye/go-harbor"
	"github.com/parnurzeal/gorequest"
)

var defaultClient *harbor.Client

func init() {
	// defaultClient = New("http://harbor.haodai.net", "wenzhengnlin", "360860aA")
	defaultClient = New("http://harbor.haodai.net", "devuser", "Ln28ohyDn")

}
func New(url, user, pass string) *harbor.Client {
	return harbor.NewClient(nil, "http://harbor.haodai.net", "wenzhengnlin", "360860aA")
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