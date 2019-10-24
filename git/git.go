// read and write and commit git
package git

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/chinglinwen/log"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

var (
	defaultGitlabURL  string
	defaultUser       string
	defaultPass       string
	gitlabAccessToken string
	defaultRepoDir    string
)

// Init init package setting
func Init(gitlabURL, user, pass, accessToken, repoDir string) {
	log.Println("inited user setting", user)
	defaultUser = user
	defaultPass = pass

	defaultGitlabURL = gitlabURL
	gitlabAccessToken = accessToken
	defaultRepoDir = repoDir

	if accessToken != "" {
		// cache admin projects at start
		go func() {
			ps, err := GetProjects(gitlabAccessToken, "yes")
			if err != nil {
				log.Fatal("fetch admin projects err", err)
			}
			log.Printf("cached admin projects at start, got %v projects\n", len(ps))
		}()
	}

}

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
	if !re.Match([]byte(branch)) {
		return false
	}
	// we may include commit id, so to distinguish it from tag
	return strings.Contains(branch, ".")
}

// pre is a tag, and not online is pre
func BranchIsPre(branch string) bool {
	if BranchIsTag(branch) {
		return !BranchIsOnline(branch)
	}
	return false
}

// v1.0.0 is online, only includes number and dot, prefix by v, suffix by number
func BranchIsOnline(branch string) bool {
	// re := regexp.MustCompile(`[^v][[:alpha:]]+`)  // not branch is a tag
	// re := regexp.MustCompile(`^v[0-9|.]+[0-9]$`) // prefix with v is a tag
	// re := regexp.MustCompile(`^v(\d+\.)(\d+\.)(\d)$`) // prefix with v is a tag, no one dot

	// prefix with v is a tag, v1,v1.0,v1.0.0 are onlines
	re := regexp.MustCompile(`^v(\d+)?(\.)?(\d+\.)?(\d)?$`)
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
	log.Debug.Println("do newrepo: ", project)
	if defaultUser == "" {
		return nil, fmt.Errorf("user empty")
	}
	if defaultPass == "" {
		return nil, fmt.Errorf("pass empty")
	}
	repo := &Repo{
		Project: project,
		Local:   filepath.Join(defaultRepoDir, project),
		URL:     fmt.Sprintf("%v/%v", defaultGitlabURL, project),
		user:    defaultUser,
		pass:    defaultPass,
	}
	for _, op := range options {
		op(repo)
	}
	log.Debug.Println("do newrepo options ok: ", project)

	if repo.Branch == "" && repo.Tag == "" {
		SetBranch("master")(repo)
	}
	log.Debug.Println("do newrepo ok: ", project)
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

	log.Debug.Println("try clone: ", project)
	err = repo.CLone()
	if err != nil {
		err = fmt.Errorf("clone err: %v", err)
		return
	}
	wrk, err := repo.R.Worktree()
	if err != nil {
		err = fmt.Errorf("get worktree error: %v, for repo: %q, branch: %q\n", err, repo.Project, repo.Branch)
		log.Println(err)
		return nil, err
	}
	repo.wrk = wrk

	log.Printf("new repo and get worktree ok, for repo: %q, branch: %q, tag: %q\n", repo.Project, repo.Branch, repo.Tag)

	// check status, if not clean just ignore checkout local, may cause later commit fail?
	status, err := wrk.Status()
	if err != nil {
		err = fmt.Errorf("check status error: %v, for repo: %q, branch: %q\n", err, repo.Project, repo.Branch)
		log.Println(err)
		return nil, err
	}
	// when there's change, we assume it's on the correct branch?
	// add this condition for commit changes by third party ( push by code, not by git command )
	if !status.IsClean() {
		log.Printf("status: %v", status)

		log.Printf("worktree is not clean, skip checkout local, for repo: %q, branch: %q\n", repo.Project, repo.Branch)
		return
	}

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
	log.Debug.Println("try newrepo: ", project)
	repo, err = New(project, options...)
	if err != nil {
		return nil, err
	}
	log.Debug.Println("try pull: ", project)
	err = repo.Pull()
	return
}

// pull will checkout local first, local change(and staged change) will be discard
func (repo *Repo) Pull() (err error) {

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
	log.Debug.Printf("do clone: url: %v\n", repo.URL)

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

	return repo.Fetch()
}

func (r *Repo) GitProjectName() string {
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
		Depth: 1, // let's do shallow fetch
	})
	if err == git.NoErrAlreadyUpToDate {
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("fetch err: %v", err)
	}
	return
}

func (repo *Repo) Tags() (tags []string, err error) {
	iter, err := repo.R.Tags()
	if err != nil {
		err = fmt.Errorf("get tags err: %v", err)
		return
	}
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, strings.TrimPrefix(string(ref.Name()), "refs/tags/"))
		return nil
	})
	return
}

func (repo *Repo) GetPreviousTag() (tag string, err error) {
	tags, err := repo.Tags()
	if err != nil {
		return
	}
	return GetPreviousTag(tags)
}

func GetPreviousTag(tags []string) (tag string, err error) {
	if tags == nil {
		err = fmt.Errorf("empty tags")
		return
	}
	sort.SliceStable(tags, func(i, j int) bool { return tags[i] > tags[j] })
	if len(tags) >= 2 {
		tag = tags[1] // (len(tags) - 2)
	} else {
		tag = tags[0]
	}
	return
}
