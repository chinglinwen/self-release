// template ops relate files
package project

import (
	"fmt"
	"path/filepath"
	"strings"
	"wen/self-release/git"
	"wen/self-release/pkg/harbor"

	"github.com/chinglinwen/log"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

const (
	defaultConfigBase     = "yunwei/config-deploy"
	defaultRepoConfigPath = "_ops" // configpath becomes project path in config-deploy
)

type Project struct {
	Project          string
	Branch           string // build branch
	Config           ProjectConfig
	configrepo       *git.Repo
	repo_            *git.Repo // not for directly use, used for avoid re-clone for same project and branch?
	envMap           map[string]string
	configConfigPath string // configpath in config-deploy

	// buildsvc for image build
	buildsvc *buildsvc
	op       projectOption
}

func (p *Project) GetBuildOutput() (chan string, error) {
	if p.buildsvc == nil {
		err := fmt.Errorf("buildsvc haven't created, should not happen")
		log.Printf("getbuild err: %v\n", err)
		return nil, err
	}
	return p.buildsvc.GetOutput()
}

func (p *Project) GetBuildError() error {
	if p.buildsvc == nil {
		err := fmt.Errorf("buildsvc haven't created, should not happen")
		log.Printf("getbuilderr err: %v\n", err)
		return err
	}
	return p.buildsvc.GetError()
}

type ProjectConfig struct {
	S SelfRelease `yaml:"selfrelease" json:"selfrelease,omitempty"` // enable or not for self-release on this project
}

type SelfRelease struct {
	Enable    bool   `yaml:"enable" json:"enable"`                 // flag to enable
	DevBranch string `yaml:"devbranch" json:"devBranch,omitempty"` // default dev branch name
	BuildMode string `yaml:"buildmode" json:"buildMode,omitempty"` // used to disable auto build [default, auto, disabled]
	ConfigVer string `yaml:"configver" json:"configVer,omitempty"` // specify different version
	Version   string `yaml:"version" json:"version,omitempty"`     // upgrade concern, different logic based on version?
}

func (c ProjectConfig) String() string {
	return fmt.Sprintf("devbranch: %v\nbuildmode: %v\nconfigver: %v\nenable: %v\nversion: %v\n",
		c.S.DevBranch, c.S.BuildMode, c.S.ConfigVer, c.S.Enable, c.S.Version)
}

const (
	// buildmodeOn   = "on"
	buildmodeAuto   = "auto"
	buildmodeManual = "manual" // for manual build
)

// the key is when does imagecheck, before auto, or after auto
// can user build before git commit? they can, but they don't know commitid?
// let's use timesplit too (but may get old image? how to filter out)
//   using previous commit time, to get last imagetag
//   see if there's imagetag after last imagetag
// user need to know image exist or not
func (p *Project) NeedBuild(commitid string) (exist, build bool) {
	switch p.Config.S.BuildMode {
	case buildmodeAuto:
		// if p.Branch == p.Config.S.DevBranch {
		// 	build = true
		// 	return
		// }
		// exist = p.ImageIsExist(commitid)
		// build = !exist
		// return
		build = true
		return
	// case buildmodeDisabled:
	// 	return false
	case buildmodeManual:
		build = false
		return
	default:
		build = true
		return
	}
}
func (p *Project) IsManual() bool {
	return p.Config.S.BuildMode == buildmodeManual
}

func (p *Project) IsEnabled() bool {
	return p.Config.S.Enable
}

// func CheckImageExist() (exist bool, err error) {
// 	tag := p.Branch
// 	if projectpkg.GetEnvFromBranch(p.Branch) == projectpkg.TEST {
// 		tag = p.CommitId
// 	}
// 	return harbor.RepoTagIsExist(p.getprojectpath(), tag)
// }

// so, let's not pass extra commitid, but let branch be commitid
func (p *Project) ImageIsExist(commitid string) bool {
	exist, err := ImageIsExist(p.Project, commitid)
	if err != nil {
		log.Printf("check if image: %v:%v exist err: %v", p.Project, p.Branch, err)
		return false
	}
	return exist
}

func ImageIsExist(project, tag string) (exist bool, err error) {
	log.Debug.Printf("check if image exist for: %v, tag: %v\n", project, tag)
	exist, err = harbor.RepoTagIsExist(project, tag)
	log.Debug.Printf("check if image exist for: %v, tag: %v, exist: %v\n", project, tag, exist)
	return
}

func BranchIsTag(branch string) bool {
	return git.BranchIsTag(branch)
}

func (p *Project) Inited() bool {
	if p != nil {
		config := filepath.Join(p.Project, configYaml)
		return p.configrepo.IsExist(config)
	} // p nil should not happen
	return false
}

// TODO: separate two type of init, let init check if docker is inited?
// gen dockerfile in user's repo?
func (p *Project) DockerInited() bool {
	repo, err := p.GetRepo()
	if err != nil {
		log.Println("getrepo err", err)
		return false
	}
	return repo.IsExist("Dockerfile")
}

// func SetGitForce() func(*Project) {
// 	return func(p *Project) {
// 		p.GitForce = true
// 	}
// }

func SetBranch(branch string) func(*projectOption) {
	return func(p *projectOption) {
		p.branch = branch
	}
}

func SetNoPull() func(*projectOption) {
	return func(p *projectOption) {
		p.nopull = true
	}
}

func SetNoReadConfig() func(*projectOption) {
	return func(p *projectOption) {
		p.noreadconfig = true
	}
}

func SetNoEnableCheck(init bool) func(*projectOption) {
	return func(p *projectOption) {
		p.noenablecheck = init
	}
}
func SetConfigMustExist(exist bool) func(*projectOption) {
	return func(p *projectOption) {
		p.configMustExist = exist
	}
}

type projectOption struct {
	nopull bool
	branch string

	devBranch string
	configVer string
	buildMode string

	noreadconfig  bool
	noenablecheck bool

	configMustExist bool
}

// a pure concept, only build need to fetch project repo
func NewProject(project string, options ...func(*projectOption)) (p *Project, err error) {
	project = strings.TrimSpace(project)
	if project == "" {
		err = fmt.Errorf("empty project name")
		return
	}
	if !strings.Contains(project, "/") {
		err = fmt.Errorf("invalid format for project, should be \"group-name/repo-name\"")
		return
	}
	log.Debug.Printf("try create project %v\n", project)

	// not inited repo, just return
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	c := &projectOption{
		// branch:    "master",
		// devBranch: "develop",
		// buildMode: "default",
		configVer: GetDefaultConfigVer(), // helm/phpv1
	}
	for _, op := range options {
		op(c)
	}
	log.Printf("project: %v, options: %#v\n", project, c)

	config, err := ReadProjectConfig(project, SetConfigRepo(configrepo))
	// if err != nil && !c.configMustExist {
	// 	log.Println("read project config err, will using default config for ", project)
	// 	// defaultConfig, e := readTemplateConfig(configrepo, c.configVer) // using default config, can we get configver now?
	// 	// if e != nil {
	// 	// 	err = fmt.Errorf("get defaultConfig from config-repo err: %v", e)
	// 	// 	return
	// 	// }
	// 	// config = defaultConfig
	// 	// err = nil
	// }
	if err != nil {
		err = fmt.Errorf("read config failed, config may not exist, err: %v", err)
		return
	}
	prefix := fmt.Sprintf("%v branch: %v projectconfig", project, c.branch)
	pretty(prefix, config)
	if !config.S.Enable && !c.noenablecheck {
		err = fmt.Errorf("project disabled, try set selfrelease=enabled")
		return
	}

	log.Printf("using configver: %v, devbranch: %v, buildmode: %v", config.S.ConfigVer, config.S.DevBranch, config.S.BuildMode)

	p = &Project{
		Project: project,
		Branch:  c.branch,
		Config:  config, // how to persist this config?
	}

	p.op = *c
	p.configrepo = configrepo
	log.Printf("create project: %q ok\n", project)
	return
}

func pretty(prefix string, a interface{}) {
	out, _ := prettyjson.Marshal(a)
	fmt.Printf("%v: %s\n", prefix, out)
}

func (p *Project) GetRepo() (repo *git.Repo, err error) {
	if p.repo_ != nil {
		repo = p.repo_
		return
	}
	c := p.op
	repo, err = getRepo(p.Project, p.Branch, c.nopull)
	if err != nil {
		err = fmt.Errorf("clone or open project: %v, err: %v, configver: %v", p.Project, err, c.configVer)
		return
	}
	p.repo_ = repo
	return
}

func (p *Project) GetWorkDir() (workdir string, err error) {
	repo, err := p.GetRepo()
	if err != nil {
		return
	}
	workdir = repo.GetWorkDir()
	return
}

func (p *Project) GetPreviousTag() (tag string, err error) {
	repo, err := p.GetRepo()
	if err != nil {
		return
	}
	tag, err = repo.GetPreviousTag()
	return
}

func getRepo(project, branch string, nopull bool) (repo *git.Repo, err error) {

	if nopull {
		repo, err = git.New(project, git.SetBranch(branch), git.SetForce())
	} else {
		repo, err = git.NewWithPull(project, git.SetBranch(branch), git.SetForce())
	}
	if err != nil {
		err = fmt.Errorf("new git repo for: %v, err: %v", project, err)
		return
	}
	return
}
