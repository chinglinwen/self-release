package git

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	EnvOnline    = "online"
	EnvPreOnline = "pre"
	EnvTest      = "test"
)

// var client *gitlab.Client

// func client() *gitlab.Client {
// 	client := gitlab.NewClient(http.DefaultClient, gitlabAccessToken)
// 	client.SetBaseURL(*defaultGitlabURL)
// 	return client
// }

func adminclient() *gitlab.Client {
	client := gitlab.NewClient(http.DefaultClient, gitlabAccessToken)
	client.SetBaseURL(defaultGitlabURL)
	return client
}

func userclient(token string) *gitlab.Client {
	client := gitlab.NewClient(http.DefaultClient, token)
	client.SetBaseURL(defaultGitlabURL)
	return client
}

func GetUserByToken(token string) (user *gitlab.User, err error) {
	c := userclient(token)
	u, _, err := c.Users.CurrentUser()
	if err != nil {
		log.Println("getuser err", err)
		return
	}
	return u, nil
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

func GetGroups(token string) (c *gitlab.Client, gs []*gitlab.Group, err error) {
	c = userclient(token)
	ps, _, e := c.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		err = fmt.Errorf("%v", strings.Split(e.Error(), "\n")[0])
		return
	}
	if len(ps) == 0 {
		err = fmt.Errorf("group: there's no any git group")
		return
	}
	gs = ps
	return
}

func GetGroupLists(token string) (gs []string, err error) {
	// for all group projects
	_, gss, err := GetGroups(token)
	if err != nil {
		log.Println("getgroups err", err)
		return
	}
	for _, g := range gss {
		// log.Println("g", g.Path)
		gs = append(gs, g.Path)
	}
	return
}

// shomehow miss some group, don't use it
// func localexist(list []string, name string) bool {
// 	k8sname := strings.Replace(name, "_", "-", -1)
// 	for _, v := range list {
// 		p := strings.Split(v, "/")[0]
// 		if k8sname == p {
// 			return true
// 		}
// 	}
// 	return false
// }

// // https://gitlab.com/gitlab-org/gitlab-ce/issues/51508, version 11.2 only
// func userIsInGroup(g *gitlab.Group, user string) bool {
// 	_, _, err := client().Groups.ListAllGroupMembers(g.ID, &gitlab.ListGroupMembersOptions{
// 		Query: &user,
// 	})
// 	fmt.Println(err)
// 	return err == nil
// }

// fetch admin differently, because fetch by ownership doesn't work for admin
func GetProjectsAdminByGroup(token string) (pss []*gitlab.Project, err error) {
	// for all group projects
	c, gs, err := GetGroups(token)
	if err != nil {
		log.Println("getgroups err", err)
		return
	}

	var wg sync.WaitGroup

	queue := make(chan []*gitlab.Project, 100) // estimate value
	wg.Add(len(gs))

	// a := gitlab.PrivateVisibility
	// access := gitlab.NoPermissions
	// list := gitlab.ListOptions{PerPage: -1}
	for _, g := range gs {
		go func(g *gitlab.Group) {
			// fmt.Println("got group", g.Path)
			defer wg.Done()

			// ps, _, e := client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			// 	Visibility: &a,
			// })
			// if e != nil {
			// 	return
			// }
			// println("len for ", len(ps), g.Path)
			// printproject(ps, "ham")
			// queue <- ps
			// ps, _, e = client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			// 	MinAccessLevel: &access,
			// })
			// if e != nil {
			// 	return
			// }
			// println("len for ", len(ps), g.Path)
			// printproject(ps, "ham")
			// queue <- ps

			// ps, _, e := c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			// 	ListOptions: list,
			// 	Visibility:  &a,
			// })
			// if e != nil {
			// 	return
			// }
			// // spew.Dump("resp", resp)
			// println("len for ", len(ps), g.Path)
			// printproject(ps, "agent")
			// queue <- ps
			// ps, _, e = c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			// 	ListOptions:    list,
			// 	MinAccessLevel: &access,
			// })
			// if e != nil {
			// 	return
			// }
			// // spew.Dump("resp", resp)
			// println("len for ", len(ps), g.Path)
			// printproject(ps, "agent")
			// queue <- ps

			listProjects(c, g, queue)
		}(g)
	}

	go func() {
		defer wg.Done()
		for ps := range queue {
			pss = append(pss, ps[:]...)
		}
	}()

	wg.Add(1)
	go func() {
		// for all personal projects inclusion
		// ps, _, err := c.Projects.ListProjects(&gitlab.ListProjectsOptions{
		// 	Visibility: &a,
		// })
		// if err != nil {
		// 	log.Println("listprojects err", err)
		// 	return
		// }
		// // println("len for personal", len(ps))
		// // printproject(ps, "agent")
		// queue <- ps
		listPersonalProjects(c, queue)

		wg.Done()
	}()
	wg.Wait()
	fmt.Println("got ", len(pss))

	pss = uniqproject(pss)
	fmt.Println("after unique ", len(pss))

	if len(pss) == 0 {
		err = fmt.Errorf("there's no any projects")
		log.Println(err)
		return
	}

	sort.Slice(pss, func(i, j int) bool {
		return pss[i].WebURL < pss[j].WebURL
	})

	// fmt.Println("len", len(pss))
	return
}

func uniqproject(pss []*gitlab.Project) (ps []*gitlab.Project) {
	keys := make(map[int]bool)
	// list := []string{}
	for _, v := range pss {
		if _, value := keys[v.ID]; !value {
			keys[v.ID] = true
			ps = append(ps, v)
		}
	}
	return
}

func listPersonalProjects(c *gitlab.Client, queue chan []*gitlab.Project) {
	// a := gitlab.PrivateVisibility
	// access := gitlab.NoPermissions
	list := gitlab.ListOptions{Page: 1, PerPage: 10000} //perpage doesn't work

	for i := 1; ; i++ {
		ps, resp, err := c.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: list,
			// Visibility:  &a,
		})
		if err != nil {
			log.Println("listprojects err", err)
			return
		}

		// println("len for personal", len(ps))
		// printproject(ps, "agent")
		queue <- ps

		if resp.TotalPages < i {
			break
		}
		list.Page = i
	}
}

func listProjects(c *gitlab.Client, g *gitlab.Group, queue chan []*gitlab.Project) {
	// a := gitlab.PrivateVisibility
	// access := gitlab.DeveloperPermissions
	list := gitlab.ListOptions{Page: 1, PerPage: 10000} //perpage doesn't work

	// ps, _, e := client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
	// 	Visibility: &a,
	// })
	// if e != nil {
	// 	return
	// }
	// println("len for ", len(ps), g.Path)
	// printproject(ps, "ham")
	// queue <- ps
	// ps, _, e = client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
	// 	MinAccessLevel: &access,
	// })
	// if e != nil {
	// 	return
	// }
	// println("len for ", len(ps), g.Path)
	// printproject(ps, "ham")
	// queue <- ps
	for i := 1; ; i++ {
		ps, resp, e := c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			ListOptions: list,
			// Visibility:  &a, //this cause need second list
			// MinAccessLevel: &access,
		})
		if e != nil {
			return
		}

		// spew.Dump("resp", resp)
		// println("len for ", len(ps), g.Path)
		// printproject(ps, "agent")
		queue <- ps

		ps, _, e = c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			ListOptions: list,
			// MinAccessLevel: &access,
		})
		if e != nil {
			return
		}
		// spew.Dump("resp", resp)
		// println("len for ", len(ps), g.Path)
		// printproject(ps, "agent")
		queue <- ps

		if resp.TotalPages < i {
			break
		}
		list.Page = i
	}
}

func printproject(ps []*gitlab.Project, name string) {
	for _, v := range ps {
		if strings.Contains(v.WebURL, name) {
			fmt.Println("ha", v.WebURL)
		}
	}
}

func GetProjects(token, refresh string) (ps []*gitlab.Project, err error) {
	u, err := GetUserByToken(token)
	if err != nil {
		return
	}
	if u.IsAdmin {
		log.Println("getting projects for admin user", u.Name)
		if ps, ok := projectsCache["admin"]; ok && refresh != "yes" {
			log.Println("get projects from cache for admin user", u.Name)
			return ps, nil
		}
		ps, err = GetProjectsAdmin()
		if err != nil {
			return
		}
		// cache it
		log.Println("cached projects for admin user", u.Name)
		projectsCache["admin"] = ps
		return
	}
	log.Println("getting projects for user", u.Name)
	if ps, ok := projectsCache[token]; ok && refresh != "yes" {
		log.Println("get projects from cache for user", u.Name)
		return ps, nil
	}
	ps, err = GetProjectsUser(token)
	if err != nil {
		return
	}
	log.Println("cached projects for user", u.Name)
	projectsCache[token] = ps
	return
}

func GetProjectsAdmin() (ps []*gitlab.Project, err error) {
	c := adminclient()

	// a := gitlab.PrivateVisibility
	// access := gitlab.NoPermissions
	list := gitlab.ListOptions{Page: 1, PerPage: 10000} //perpage doesn't work

	// ps, _, err = c.Projects.ListProjects(&gitlab.ListProjectsOptions{
	// 	ListOptions: list,
	// 	// Visibility:  &a,
	// })
	// if err != nil {
	// 	log.Println("listprojects2 err", err)
	// 	return
	// }
	// return

	for i := 1; ; i++ {
		p, resp, e := c.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: list,
			// Visibility:  &a,
		})
		if e != nil {
			log.Println("listprojects err", e)
			err = e
			return
		}

		// println("len for personal", len(ps))
		// printproject(ps, "agent")
		// queue <- ps
		if resp.TotalPages < i {
			break
		}
		ps = append(ps, p...)
		list.Page = i
	}

	// fmt.Println("got ", len(ps))

	ps = uniqproject(ps)
	// fmt.Println("after unique ", len(ps))
	return
}

// only got one project
// func GetProjectsByUser(token string) (ps []*gitlab.Project, err error) {
// 	u, err := GetUserByToken(token)
// 	if err != nil {
// 		return
// 	}
// 	a := gitlab.PrivateVisibility
// 	list := gitlab.ListOptions{Page: 1, PerPage: 10000}
// 	c := userclient(token)
// 	ps, _, err = c.Projects.ListUserProjects(u.ID, &gitlab.ListProjectsOptions{
// 		ListOptions: list,
// 		Visibility:  &a,
// 	})
// 	return
// }

// https://docs.gitlab.com/ce/api/projects.html#list-projects
func GetProjectsUser(token string) (ps []*gitlab.Project, err error) {
	c := userclient(token)
	list := gitlab.ListOptions{Page: 1, PerPage: 10000}

	yes := true
	// by := "id"
	ps, _, err = c.Projects.ListProjects(&gitlab.ListProjectsOptions{
		ListOptions: list,
		Membership:  &yes,
		// OrderBy:     &by,
	})
	if err != nil {
		log.Println("listprojects err", err)
		return
	}

	if len(ps) == 0 {
		err = fmt.Errorf("there's no any projects")
		log.Println(err)
		return
	}

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].WebURL < ps[j].WebURL
	})
	return
}

// func GetProjectLists(token string) (admin bool, projects []string, err error) {
// 	isadmin, e := IsAdmin(token)
// 	if err != nil {
// 		err = fmt.Errorf("check admin error %v", e)
// 		return
// 	}
// 	if isadmin {
// 		// // filter list to reduce project searching time
// 		// list, e := listpods()
// 		// if e != nil {
// 		// 	err = fmt.Errorf("walk error %v", e)
// 		// 	return
// 		// }
// 		admin = true
// 		return
// 	}

// 	pss, err := GetProjects(token)
// 	if err != nil {
// 		log.Println("getprojects err", err)
// 		return
// 	}
// 	for _, p := range pss {
// 		// fmt.Println("--", p.PathWithNamespace)
// 		// spew.Dump("p", p)
// 		// url := strings.Split(p.WebURL, "/")
// 		// if len(url) != 5 {
// 		// 	log.Println("get project list warn: bad format %v", p.WebURL)
// 		// 	continue
// 		// }
// 		// git := fmt.Sprintf("%v/%v", url[3], url[4])
// 		git := strings.Replace(p.PathWithNamespace, " ", "", -1) //remove empty space
// 		projects = append(projects, git)
// 	}
// 	return false, unique(projects), nil
// }

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// projectPath is org/repo
func GetGitProject(projectPath string) (project *gitlab.Project, err error) {
	project, _, err = adminclient().Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	return
}

func IsAdmin(token string) (isadmin bool, err error) {
	u, err := GetUser(token)
	if err != nil {
		err = fmt.Errorf("get user err: %v", err)
		return
	}
	if u.IsAdmin {
		// envs = append(envs, EnvPreOnline, EnvTest)
		isadmin = true
		return
	}
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

// get gitlab tags
func CheckTagExist(projectPath, tag string) (t *gitlab.Tag, err error) {
	p, err := GetProject(projectPath)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		return
	}
	t, _, err = adminclient().Tags.GetTag(p.ID, tag)
	if err != nil {
		if strings.Contains(err.Error(), "404 Tag Not Found") {
			err = fmt.Errorf("tag not found")
		}
		err = fmt.Errorf("get tag %v err: %v", tag, err)
		return
	}
	return
}
func SetCommitStatusSuccess(projectPath, tag, logurl string) (t *gitlab.CommitStatus, err error) {
	return SetCommitStatus(projectPath, tag, logurl, string(gitlab.Success))
}
func SetCommitStatusPending(projectPath, tag, logurl string) (t *gitlab.CommitStatus, err error) {
	return SetCommitStatus(projectPath, tag, logurl, string(gitlab.Pending))
}
func SetCommitStatusRunning(projectPath, tag, logurl string) (t *gitlab.CommitStatus, err error) {
	return SetCommitStatus(projectPath, tag, logurl, string(gitlab.Running))
}
func SetCommitStatusFailed(projectPath, tag, logurl string) (t *gitlab.CommitStatus, err error) {
	return SetCommitStatus(projectPath, tag, logurl, string(gitlab.Failed))
}

// it can be tag
func SetCommitStatus(projectPath, tag, logurl, state string) (t *gitlab.CommitStatus, err error) {
	p, err := GetProject(projectPath)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		return
	}
	name := "self-release"
	t, _, err = adminclient().Commits.SetCommitStatus(p.ID, tag, &gitlab.SetCommitStatusOptions{
		State:     gitlab.BuildStateValue(state),
		Name:      &name,
		TargetURL: &logurl,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			err = fmt.Errorf("commit not found")
		}
		err = fmt.Errorf("get commit %v err: %v", tag, err)
		return
	}
	if t.Status != state {
		err = fmt.Errorf("set commit status to %v failed, got %v", state, t.Status)
		return
	}
	return
}

func GetCommitIDFromTag(projectPath, tag string) (id string, err error) {
	t, err := GetCommitFromTag(projectPath, tag)
	if err != nil {
		return
	}
	// we only use first 8 bytes
	id = t.ID[:8]
	return
}

func GetCommitFromTag(projectPath, tag string) (t *gitlab.Commit, err error) {
	p, err := GetProject(projectPath)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		return
	}
	t, _, err = adminclient().Commits.GetCommit(p.ID, tag)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			err = fmt.Errorf("commit not found")
		}
		err = fmt.Errorf("get commit %v err: %v", tag, err)
		return
	}
	return
}

func listAllTags(projectPath string) (ts []*gitlab.Tag, err error) {
	p, err := GetProject(projectPath)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		return
	}
	list := gitlab.ListOptions{Page: 1, PerPage: 10000}
	ts, _, err = adminclient().Tags.ListTags(p.ID, &gitlab.ListTagsOptions{
		ListOptions: list,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			err = fmt.Errorf("commit not found")
		}
		err = fmt.Errorf("list tags for %v err: %v", projectPath, err)
		return
	}
	if len(ts) == 0 {
		err = fmt.Errorf("empty tags for %v", projectPath)
		return
	}
	return
}

func GetLastTagCommitID(projectPath string) (onlineid, preid string, err error) {
	o, p, err := GetLastTag(projectPath)
	if err != nil {
		return
	}
	if o != nil {
		onlineid = o.Commit.ShortID
	}
	if p != nil {
		preid = p.Commit.ShortID
	}
	return
}
func GetLastTag(projectPath string) (online, pre *gitlab.Tag, err error) {
	ts, err := listAllTags(projectPath)
	if err != nil {
		err = fmt.Errorf("get last tag err: %v", err)
		return
	}
	preTags := []*gitlab.Tag{}
	onlineTags := []*gitlab.Tag{}
	for _, v := range ts {
		if BranchIsOnline(v.Name) {
			onlineTags = append(onlineTags, v)
		} else {
			preTags = append(preTags, v)
		}
	}
	if len(onlineTags) != 0 {
		online = onlineTags[0]
	}
	if len(preTags) != 0 {
		pre = preTags[0]
	}
	return
}

func CheckPerm(projectPath, user, env string) (err error) {
	u, err := GetUser(user)
	if err != nil {
		err = fmt.Errorf("get user err: %v", err)
		return
	}
	if u.IsAdmin {
		log.Printf("%v is admin, allowed\n", u.Name)
		return nil
	}
	group, _ := getGroupAndName(projectPath)
	if group == user {
		return nil
	}
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
	// spew.Dump("u:", u)
	al, err := getAccessLevel(p.ID, g.ID, u.ID)
	if err != nil {
		err = fmt.Errorf("get access level err: %v", err)
		return
	}
	envs := getAllowedEnv(al)
	if !isEnvOk(env, envs) {
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

// func filterEnvs(envs []string, path string) (err error) {
// 	dirs, err := getDirs(path)
// 	if err != nil {
// 		return
// 	}
// 	for _,v:=range envs{
// 		for _,v:=range dirs{

// 		}
// 	}
// }

// func getDirs(path string) (dirs []string, err error) {
// 	files, err := ioutil.ReadDir(path)
// 	if err != nil {
// 		return
// 	}
// 	for _, file := range files {
// 		if !file.IsDir() {
// 			continue
// 		}
// 		dirs = append(dirs, file.Name())
// 	}
// 	if len(dirs) == 0 {
// 		return nil, fmt.Errorf("no log dirs found")
// 	}
// 	return
// }

// func GetGitGroup(group string) (g *gitlab.Group, err error) {
// 	ps, _, err := client().Groups.ListGroups(&gitlab.ListGroupsOptions{
// 		// Search: &group,
// 	})
// 	for _, v := range ps {
// 		fmt.Println(v.WebURL)
// 	}
// 	if err != nil {
// 		err = fmt.Errorf("%v", strings.Split(err.Error(), "\n")[0])
// 		return
// 	}
// 	if len(ps) == 0 {
// 		err = fmt.Errorf("group: %v not found", group)
// 		return
// 	}
// 	g = ps[0]
// 	return
// }

// func GetGitProject(git string) (p *gitlab.Project, err error) {
// 	gr := strings.Split(git, "/")
// 	if len(gr) != 2 {
// 		err = fmt.Errorf("git: %v invalid format, eg: group/project ", git)
// 		return
// 	}
// 	group, repo := gr[0], gr[1]

// 	var g *gitlab.Group
// 	g, err = GetGitGroup(group)
// 	if err != nil {
// 		return
// 	}
// 	ps, _, e := client().Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
// 		// Membership: &a,
// 		Search: &repo,
// 	})
// 	if e != nil {
// 		err = e
// 		return
// 	}
// 	if len(ps) == 0 {
// 		err = fmt.Errorf("repo: %v not found", repo)
// 		return
// 	}
// 	p = ps[0]
// 	return
// }

// func ListUsers() ([]*gitlab.User, error) {
// 	u, _, err := client().Users.ListUsers(&gitlab.ListUsersOptions{})
// 	return u, err
// }
