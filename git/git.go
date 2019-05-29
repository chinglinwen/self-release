// read and write and commit git
package git

import (

	// "gopkg.in/src-d/go-billy.v4"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

var (
	// https://github.com/src-d/go-git/issues/999 token is deprecated
	// defaultToken = flag.String("gitlab-token", "", "default gitlab token")

	defaultGitlabURL = flag.String("gitlab-url", "http://g.haodai.net", "default gitlab url")
	defaultUser      = flag.String("gitlab-user", "", "default gitlab user")
	defaultPass      = flag.String("gitlab-pass", "", "default gitlab pass(personal token is ok")

	defaultRepoDir = flag.String("repoDir", "/home/wen/t/repos", "default path to store cloned projects")
)

func Init(user, pass string) {
	log.Println("inited user setting", user)
	*defaultUser = user
	*defaultPass = pass
}

// var (
// 	fs = osfs.New("gitdir")
// )

// func init(){
// fs=osfs.New("gitdir")
// }

type Repo struct {
	Project string // org/repo
	Local   string
	URL     string
	Tag     string // what version, what branch?
	Branch  string

	R          *git.Repository
	wrk        *git.Worktree
	nocheckout bool
	nopull     bool
	force      bool // default no force

	refs      string // with  remote refs
	localrefs string // with local refs

	// token     string //gitlab token
	user string
	pass string
}

/*
$ git show-ref
e48ff73f447c62fc1fc704ab598aa02ce6ac71ae refs/heads/feature1
15c00cb6b4606d3799768a8d0a1a9d51a182c1dc refs/heads/master
e48ff73f447c62fc1fc704ab598aa02ce6ac71ae refs/heads/v1.0.0
e48ff73f447c62fc1fc704ab598aa02ce6ac71ae refs/remotes/origin/feature1
1ae1df47e41b4074521c32ed1ad042c5d5fbca72 refs/remotes/origin/feature2
15c00cb6b4606d3799768a8d0a1a9d51a182c1dc refs/remotes/origin/master
774d8b527de79c337f4f282f7531db1a418ca960 refs/tags/v1.0.0
*/

// using branch
// if it empty, later it will set to master
// prefix with v will set as tag
func SetBranch(branch string) func(*Repo) {
	if BranchIsTag(branch) {
		return SetTag(branch)
	}
	return func(r *Repo) {
		r.Branch = branch
		r.refs = fmt.Sprintf("refs/remotes/origin/%v", branch)
		r.localrefs = fmt.Sprintf("refs/heads/%v", branch)
	}
}

// prefix with v is a tag
func BranchIsTag(branch string) bool {
	// re := regexp.MustCompile(`[^v][[:alpha:]]+`)  // not branch is a tag
	re := regexp.MustCompile(`^v.+`) // prefix with v is a tag
	return re.Match([]byte(branch))
}

// not online is pre
func BranchIsPre(branch string) bool {
	return !BranchIsOnline(branch)
}

// v1.0.0 is online, only includes number and dot, prefix by v, suffix by number
func BranchIsOnline(branch string) bool {
	// re := regexp.MustCompile(`[^v][[:alpha:]]+`)  // not branch is a tag
	re := regexp.MustCompile(`^v[0-9|.]+[0-9]$`) // prefix with v is a tag
	return re.Match([]byte(branch))
}

func SetTag(tag string) func(*Repo) {
	return func(r *Repo) {
		r.Tag = tag
		r.refs = fmt.Sprintf("refs/tags/%v", tag)
		r.localrefs = fmt.Sprintf("refs/heads/%v", tag)
	}
}

func SetLocalPath(localpath string) func(*Repo) {
	return func(r *Repo) {
		r.Local = localpath
	}
}

func SetNoPull() func(*Repo) {
	return func(r *Repo) {
		r.nopull = true
	}
}
func SetNoCheckout() func(*Repo) {
	return func(r *Repo) {
		r.nocheckout = true
	}
}

// setforce does not fix non-fast-forward update(which often cause by human edit?)
func SetForce() func(*Repo) {
	return func(r *Repo) {
		r.force = true
	}
}

func newrepo(project string, options ...func(*Repo)) (*Repo, error) {
	if *defaultUser == "" {
		return nil, fmt.Errorf("user empty")
	}
	if *defaultPass == "" {
		return nil, fmt.Errorf("pass empty")
	}

	// t := strings.Split(strings.TrimSuffix(localpath, ".git"), "/")
	// name := t[len(t)-1]
	repo := &Repo{
		Project: project,
		Local:   filepath.Join(*defaultRepoDir, project),
		URL:     fmt.Sprintf("%v/%v", *defaultGitlabURL, project),
		user:    *defaultUser,
		pass:    *defaultPass,
	}
	for _, op := range options {
		op(repo)
	}

	if repo.Branch == "" && repo.Tag == "" {
		SetBranch("master")(repo)
		// repo.Branch = "master"
		// repo.refs = fmt.Sprintf("refs/remotes/origin/%v", repo.Branch)
		// repo.localrefs = fmt.Sprintf("refs/heads/%v", repo.Branch)
	}
	return repo, nil
}

func New(project string, options ...func(*Repo)) (repo *Repo, err error) {
	if project == "" {
		err = fmt.Errorf("project name is empty")
		return
	}
	repo, err = newrepo(project, options...)
	if err != nil {
		return nil, err
	}

	err = repo.CLone()
	if err != nil {

	}
	wrk, err := repo.R.Worktree()
	if err != nil {
		err = fmt.Errorf("get worktree error: %v, for repo: %q, branch: %q\n", err, repo.Project, repo.Branch)
		log.Println(err)
		return nil, err
	}
	repo.wrk = wrk

	log.Printf("new repo and get worktree ok, for repo: %q, branch: %q, tag: %q\n", repo.Project, repo.Branch, repo.Tag)

	// this will make local changes lost
	// checkout is needed after new, so we can work on correct branch
	err = repo.CheckoutLocal()
	if err != nil {
		err = fmt.Errorf("checkout to local error: %v, for repo: %q, branch: %q\n", err, repo.Project, repo.Branch)
		log.Println(err)
		return nil, err
	}
	return
}

func NewWithPull(project string, options ...func(*Repo)) (repo *Repo, err error) {
	repo, err = New(project, options...)
	if err != nil {
		return nil, err
	}
	err = repo.Pull()
	return
}

// pull will checkout local first, local change(and staged change) will be discard
func (repo *Repo) Pull() (err error) {
	// if !repo.nocheckout {
	// 	// 	// err = repo.CheckoutLocal()  // this checkout two times
	// 	if repo.Branch != "master" {
	// 		err = repo.wrk.Checkout(&git.CheckoutOptions{
	// 			Branch: plumbing.ReferenceName("refs/heads/master"),
	// 			Force:  repo.force,
	// 		})
	// 	}
	// 	if err != nil {
	// 		err = fmt.Errorf("checkoutlocal to master before pull error: %v, for repo: %v\n", err, repo.Project)
	// 		log.Println(err)
	// 		return
	// 	}
	// 	log.Println("checkoutlocal to master before pull ok, for repo:", repo.Project)
	// 	// }
	// } else {
	// 	log.Printf("will not do checkout local for: %v, branch: %v\n", repo.Project, repo.Branch)
	// }

	// skip pull for tag
	if repo.nopull || repo.Tag != "" {
		log.Println("will not do pull for", repo.Project)
		return
	}
	// pull can be done if all commit been pushed ( otherwise result non-fast-forward error )
	err = repo.wrk.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(repo.localrefs), // refs/heads/v1.0.1
		// ReferenceName: plumbing.ReferenceName(repo.refs), // refs/tags/v1.0.1 //reference not found, object not found for tag
		// SingleBranch:  true,
		Auth: &http.BasicAuth{
			Username: repo.user,
			Password: repo.pass,
		},
		// Depth: 1,
		Force: repo.force, // TODO: default no force?
	})
	// spew.Dump("pull err", err, err == git.NoErrAlreadyUpToDate)
	if err == git.NoErrAlreadyUpToDate {
		err = nil
	}
	if err != nil && err != git.NoErrAlreadyUpToDate {
		err = fmt.Errorf("pull error: %v, for repo: %v", err, repo.Project)
		return
	}
	log.Printf("pull ok, for repo: %q\n", repo.Project)

	return repo.CheckoutLocal()
}

func (repo *Repo) GetWorkDir() string {
	return repo.Local
}

func (repo *Repo) CLone() (err error) {
	var r *git.Repository
	r, err = git.PlainOpen(repo.Local)
	if err != nil {
		// Clones the repository into the given dir, just as a normal git clone does
		r, err = git.PlainClone(repo.Local, false, &git.CloneOptions{
			URL: repo.URL,
			Auth: &http.BasicAuth{
				Username: repo.user,
				Password: repo.pass,
			},
			// NoCheckout: repo.force,
			// Depth:         1,  // depth 1 will cause object not found
			// enable ReferenceName will cause non-fast-forward update error
			// ReferenceName: plumbing.ReferenceName(repo.refs), // default all branches
			Tags: git.AllTags,
		})
		log.Println("cloned new repo :", repo.Project)
	} else {
		log.Printf("got existing repo ok, for repo: %q\n", repo.Project)
	}
	repo.R = r

	// if repo.Tag != "" {
	return repo.Fetch()
	// }
	// return
}

func (r *Repo) GitProjectName() string {
	if r.Project == "" {
		// get name from url?
		//r.Name=
	}
	return r.Project
}

func (repo *Repo) Fetch() (err error) {
	err = repo.R.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			// config.RefSpec(repo.refs),
			// config.RefSpec("+" + repo.refs + ":" + repo.refs),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
		Auth: &http.BasicAuth{
			Username: repo.user,
			Password: repo.pass,
		},
		Tags:  git.AllTags,
		Force: true,
	})
	if err != nil {
		err = fmt.Errorf("fetch err: %v", err)
	}
	return
}
