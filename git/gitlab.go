package git

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"net/http"

	gitlab "github.com/xanzy/go-gitlab"
)

const (
	EnvOnline    = "online"
	EnvPreOnline = "pre"
	EnvTest      = "test"
)

// var (
// 	// GitlabEndpoint    = flag.String("gitlaburl", "http://g.haodai.net", "gitlab base url")
// 	GitlabAccessToken = flag.String("gitlabtoken", "", "gitlab access token")
// )

func adminclient() *gitlab.Client {
	client := gitlab.NewClient(http.DefaultClient, *gitlabAccessToken)
	client.SetBaseURL(*defaultGitlabURL)
	return client
}

func GetUser(name string) (user *gitlab.User, err error) {
	c := adminclient()
	us, _, err := c.Users.ListUsers(&gitlab.ListUsersOptions{
		Search: &name,
	})
	if err != nil {
		err = fmt.Errorf("getuser: %v err: %v", name, err)
		return
	}
	if len(us) == 0 {
		err = fmt.Errorf("getuser: %v not found", name)
		return
	}
	if len(us) > 1 {
		var many string
		for _, v := range us {
			many = fmt.Sprintf("%v %v", many, v.Name)
		}
		err = fmt.Errorf("getuser: %v, got many (%v)", name, many)
		return
	}
	user = us[0]
	return
}
func getGroupAndName(pathname string) (group, name string) {
	s := strings.Split(pathname, "/")
	if len(s) >= 1 {
		group = s[0]
	}
	if len(s) >= 2 {
		name = s[1]
	}
	return
}

func GetGroup(pathname string) (g *gitlab.Group, err error) {
	group, _ := getGroupAndName(pathname)
	c := adminclient()
	ps, _, err := c.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: &group,
	})
	if err != nil {
		err = fmt.Errorf("%v", strings.Split(err.Error(), "\n")[0])
		return
	}
	if len(ps) == 0 {
		err = fmt.Errorf("group: %v not found", group)
		return
	}
	g = ps[0]
	return
}

func GetProject(pathname string) (p *gitlab.Project, err error) {
	_, name := getGroupAndName(pathname)
	ok := true
	ps, _, err := adminclient().Projects.ListProjects(&gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10000,
		},
		Search: &name,
		Simple: &ok,
	})
	if err != nil {
		err = fmt.Errorf("getproject: %v err: %v", pathname, err)
		return
	}
	if len(ps) == 0 {
		err = fmt.Errorf("getproject: %v not found", pathname)
		return
	}
	var many string
	for _, v := range ps {
		// log.Printf("p: %v", v.PathWithNamespace)
		many = fmt.Sprintf("%v %v", many, v.PathWithNamespace)
		if v.PathWithNamespace == pathname {
			p = v
			return
		}
	}
	// if len(us) > 1 {
	// 	var many string
	// 	for _, v := range us {
	// 		many = fmt.Sprintf("%v %v", many, v.Name)
	// 	}
	// 	err = fmt.Errorf("getproject: %v, got many (%v)", name, many)
	// 	return
	// }
	err = fmt.Errorf("getproject not matched, similar projects: %v", many)
	return
}

// can't use this for gitlab v10
func GetProjectMembers(pathname string) (err error) {
	p, err := GetProject(pathname)
	if err != nil {
		return
	}
	// This is only available since 11.2. You would need to upgrade your instance for this endpoint to be available:
	// ps, _, err := c.ProjectMembers.ListAllProjectMembers(p.ID, &gitlab.ListProjectMembersOptions{})
	ps, _, err := adminclient().ProjectMembers.ListProjectMembers(p.ID, &gitlab.ListProjectMembersOptions{})
	if err != nil {
		err = fmt.Errorf("listmember err: %v", err)
		return
	}
	for _, v := range ps {
		// log.Printf("u: %v\n", v.Name)
		spew.Dump("u", v)
	}
	return
}

func CheckPerm(projectPath, user, env string) (allow bool, err error) {
	g, err := GetGroup(projectPath)
	if err != nil {
		err = fmt.Errorf("get group err: %v", err)
		return
	}

	p, err := GetProject(projectPath)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		return
	}
	// spew.Dump("p:", p)
	u, err := GetUser(user)
	if err != nil {
		err = fmt.Errorf("get user err: %v", err)
		return
	}
	// spew.Dump("u:", u)
	al, err := getAccessLevel(p.ID, g.ID, u.ID)
	if err != nil {
		err = fmt.Errorf("get access level err: %v", err)
		return
	}
	envs := getAllowedEnv(al)
	allow = isEnvOk(env, envs)
	if !allow {
		err = fmt.Errorf("permission denied, allowed envs: %v", envs)
		return
	}
	return
}

func getAccessLevel(pid, gid, userid int) (accessLevel gitlab.AccessLevelValue, err error) {
	var groupAccessLevel, projectAccessLevel gitlab.AccessLevelValue
	c := adminclient()
	// spew.Dump("p:", p)
	// spew.Dump("pns", p.Namespace)
	groupMember, _, e := c.GroupMembers.GetGroupMember(gid, userid)
	// fmt.Println("groupmembers err", err)
	if e == nil {
		groupAccessLevel = groupMember.AccessLevel
	}
	projectMember, _, e := c.ProjectMembers.GetProjectMember(pid, userid)
	//fmt.Println("projectmember err", err, project, userid)
	if e == nil {
		projectAccessLevel = projectMember.AccessLevel
	}
	if groupAccessLevel > projectAccessLevel {
		accessLevel = groupAccessLevel
	} else {
		accessLevel = projectAccessLevel
	}
	return
}

func getAllowedEnv(accessLevel gitlab.AccessLevelValue) (envs []string) {
	if accessLevel >= gitlab.DeveloperPermissions {
		envs = append(envs, EnvPreOnline, EnvTest)
	}
	if accessLevel >= gitlab.MasterPermissions {
		envs = append(envs, EnvOnline)
	}
	return
}

func isEnvOk(env string, envs []string) bool {
	for _, v := range envs {
		if env == v {
			return true
		}
	}
	return false
}

// func listProjects(c *gitlab.Client, g *gitlab.Group, queue chan []*gitlab.Project) {
// 	a := gitlab.PrivateVisibility
// 	access := gitlab.DeveloperPermissions
// 	list := gitlab.ListOptions{Page: 1, PerPage: 1000} //perpage doesn't work

// 	// ps, _, e := client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 	// 	Visibility: &a,
// 	// })
// 	// if e != nil {
// 	// 	return
// 	// }
// 	// println("len for ", len(ps), g.Path)
// 	// printproject(ps, "ham")
// 	// queue <- ps
// 	// ps, _, e = client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 	// 	MinAccessLevel: &access,
// 	// })
// 	// if e != nil {
// 	// 	return
// 	// }
// 	// println("len for ", len(ps), g.Path)
// 	// printproject(ps, "ham")
// 	// queue <- ps
// 	for i := 1; ; i++ {
// 		ps, resp, e := c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 			ListOptions:    list,
// 			Visibility:     &a, //this cause need second list
// 			MinAccessLevel: &access,
// 		})
// 		if e != nil {
// 			return
// 		}

// 		// spew.Dump("resp", resp)
// 		// println("len for ", len(ps), g.Path)
// 		// printproject(ps, "agent")
// 		queue <- ps

// 		ps, _, e = c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 			ListOptions:    list,
// 			MinAccessLevel: &access,
// 		})
// 		if e != nil {
// 			return
// 		}
// 		// spew.Dump("resp", resp)
// 		// println("len for ", len(ps), g.Path)
// 		// printproject(ps, "agent")
// 		queue <- ps

// 		if resp.TotalPages < i {
// 			break
// 		}
// 		list.Page = i
// 	}
// }

// func GetProjects() (c *gitlab.Client, pss []*gitlab.Project, err error) {
// 	// for all group projects
// 	c, gs, err := GetGroups(user, pass)
// 	if err != nil {
// 		log.Println("getgroups err", err)
// 		return
// 	}
// 	// yes := true
// 	a := gitlab.PrivateVisibility
// 	for _, g := range gs {
// 		ps, _, e := c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 			// Membership: &yes,
// 			Visibility: &a,
// 		})
// 		if e != nil {
// 			continue
// 		}
// 		pss = append(pss, ps[:]...)
// 	}

// 	// for all personal projects inclusion
// 	ps, _, err := c.Projects.ListProjects(&gitlab.ListProjectsOptions{})
// 	if err != nil {
// 		log.Println("listprojects err", err)
// 		return
// 	}
// 	pss = append(pss, ps[:]...)

// 	if len(pss) == 0 {
// 		err = fmt.Errorf("there's no any projects")
// 		log.Println(err)
// 		return
// 	}
// 	// for _, v := range pss {
// 	// 	if strings.Contains(v.WebURL, "yunwei/worktile") || strings.Contains(v.WebURL, "yunwei/trx") {
// 	// 		spew.Dump("v:", v)
// 	// 		// fmt.Println(v.WebURL, v.RequestAccessEnabled)
// 	// 	}
// 	// 	fmt.Println(v.WebURL, v.RequestAccessEnabled)
// 	// }
// 	return
// }
