package git

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/chinglinwen/log"

	"github.com/davecgh/go-spew/spew"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	EnvOnline    = "online"
	EnvPreOnline = "pre"
	EnvTest      = "test"
)

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
		many = fmt.Sprintf("%v %v", many, v.PathWithNamespace)
		if v.PathWithNamespace == pathname {
			p = v
			return
		}
	}
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
			defer wg.Done()

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
	return
}

func uniqproject(pss []*gitlab.Project) (ps []*gitlab.Project) {
	keys := make(map[int]bool)
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
	for i := 1; ; i++ {
		ps, resp, e := c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			ListOptions: list,
			// Visibility:  &a, //this cause need second list
			// MinAccessLevel: &access,
		})
		if e != nil {
			return
		}
		queue <- ps

		ps, _, e = c.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
			ListOptions: list,
			// MinAccessLevel: &access,
		})
		if e != nil {
			return
		}
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
		if resp.TotalPages < i {
			break
		}
		ps = append(ps, p...)
		list.Page = i
	}
	ps = uniqproject(ps)
	return
}

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
func SetCommitStatusSkipped(projectPath, tag, logurl string) (t *gitlab.CommitStatus, err error) {
	return SetCommitStatus(projectPath, tag, logurl, string(gitlab.Skipped))
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

// func listLastTwoCommits(projectPath string) (ts []*gitlab.Commit, err error) {
// 	p, err := GetProject(projectPath)
// 	if err != nil {
// 		err = fmt.Errorf("get project err: %v", err)
// 		return
// 	}
// 	list := gitlab.ListOptions{Page: 1, PerPage: 2}
// 	ts, _, err = adminclient().Commits.ListCommits(p.ID, &gitlab.ListCommitsOptions{
// 		ListOptions: list,
// 	})
// 	if err != nil {
// 		if strings.Contains(err.Error(), "Not Found") {
// 			err = fmt.Errorf("commit not found")
// 		}
// 		err = fmt.Errorf("list commit for %v err: %v", projectPath, err)
// 		return
// 	}
// 	if len(ts) == 0 {
// 		err = fmt.Errorf("empty commit for %v", projectPath)
// 		return
// 	}
// 	if len(ts) < 2 {
// 		err = fmt.Errorf("no enough previous commit, got %v, expect 2", len(ts))
// 		return
// 	}
// 	return
// }

// func GetLastTagCommitID(projectPath string) (onlineid, preid string, err error) {
// 	o, p, err := GetLastTag(projectPath)
// 	if err != nil {
// 		return
// 	}
// 	if o != nil {
// 		onlineid = o.Commit.ShortID
// 	}
// 	if p != nil {
// 		preid = p.Commit.ShortID
// 	}
// 	return
// }

// func GetLastTag(projectPath string) (online, pre *gitlab.Tag, err error) {
// 	ts, err := listAllTags(projectPath)
// 	if err != nil {
// 		err = fmt.Errorf("get last tag err: %v", err)
// 		return
// 	}
// 	preTags := []*gitlab.Tag{}
// 	onlineTags := []*gitlab.Tag{}
// 	for _, v := range ts {
// 		if BranchIsOnline(v.Name) {
// 			onlineTags = append(onlineTags, v)
// 		} else {
// 			preTags = append(preTags, v)
// 		}
// 	}
// 	if len(onlineTags) != 0 {
// 		online = onlineTags[0]
// 	}
// 	if len(preTags) != 0 {
// 		pre = preTags[0]
// 	}
// 	return
// }

// // get previous commit before commitid
// func GetPreviousCommit(projectPath, commitid string) (previous *gitlab.Commit, err error) {
// 	ts, err := listLastTwoCommits(projectPath)
// 	if err != nil {
// 		err = fmt.Errorf("listLastTwoCommits err: %v", err)
// 		return
// 	}
// 	if ts[0].ShortID != commitid {
// 		err = fmt.Errorf("no previous tags found")
// 		return
// 	}
// 	previous = ts[1]
// 	return
// }

// const TimeLayout = "2006-1-2_15:04:05"

// error-prone, forget it
// if no previous commit, any imagetag is exist, return false
// if have previous commit, there's imagetag after previous commit time, return true

// calc how many images between commit
// if previous image before previous commit, compare to 2
// if previous image after previous commit, compare to 3
//    a  b
//   i  j  k
// func ManualPushedImage(projectPath, commitid string) (imagetag string, err error) {
// 	ts, err := listLastTwoCommits(projectPath)
// 	if err != nil {
// 		err = fmt.Errorf("listLastTwoCommits err: %v", err)
// 		return
// 	}

// 	// ts := oldc.CreatedAt.Local().Format(TimeLayout)
// 	ts := oldc.CreatedAt.Format(TimeLayout)
// 	log.Printf("check if image exist before ts: %v for %v, commitid: %v\n", ts, projectPath, commitid)
// 	imagetag, err = harbor.ListRepoTagLatestName(projectPath, ts)
// 	if err != nil {
// 		return
// 	}
// 	return
// }
// func ManualPushedImage(projectPath, commitid string) (imagetag string, err error) {
// 	oldc, err := GetPreviousCommit(projectPath, commitid)
// 	if err != nil {
// 		return
// 	}
// 	// ts := oldc.CreatedAt.Local().Format(TimeLayout)
// 	ts := oldc.CreatedAt.Format(TimeLayout)
// 	log.Printf("check if image exist before ts: %v for %v, commitid: %v\n", ts, projectPath, commitid)
// 	imagetag, err = harbor.ListRepoTagLatestName(projectPath, ts)
// 	if err != nil {
// 		return
// 	}
// 	return
// }

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
