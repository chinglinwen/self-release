// template ops relate files
package project

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"wen/self-release/git"
	"wen/self-release/pkg/harbor"
)

const (
	defaultConfigBase = "yunwei/config-deploy"
	// defaultAppName    = "self-release"

	// defaultConfigName = "config.env" // later will prefix with default or customize version
	// defaultConfigYAML = "config.yaml"
	// opsDir         = "_ops"
	defaultRepoConfigPath = "_ops" // configpath becomes project path in config-deploy
)

// this will be the project config for customizing
type Project struct {
	Project string
	Branch  string // build branch
	// Env     string // branch, may derive from event's branch as env
	// not able to get branch? we can, but if it's a tag? init for develop branch only no tags
	Config ProjectConfig

	// ConfigFile string // _ops/config.yaml  //set for every env? what's the difference
	// Files []File
	// EnvFiles []string // for setting of template, env.sh ?  no need export

	// GitForce  bool   // git pull force  default is force
	// InitForce bool   // init project config force, force for re-init, file setting, we often setting it by tag msg
	// NoPull bool // `yaml:"nopull"`

	// image ?
	// size of replicas?
	configrepo *git.Repo
	repo_      *git.Repo // not for directly use, used for avoid re-clone for same project and branch?
	// WorkDir    string // git local path
	envMap map[string]string
	// autoenv map[string]string // env from hook

	op projectOption
	// init            initOption
	// genOption        genOption
	configConfigPath string // configpath in config-deploy
	// env              map[string]string // store config.env values, only init need this
}

type ProjectConfig struct {
	S SelfRelease `yaml:"selfrelease" json:"selfrelease,omitempty"` // enable or not for self-release on this project
}

type SelfRelease struct {
	Enable    bool   `yaml:"enable" json:"enable,omitempty"`       // flag to enable
	DevBranch string `yaml:"devbranch" json:"devBranch,omitempty"` // default dev branch name
	BuildMode string `yaml:"buildmode" json:"buildMode,omitempty"` // used to disable auto build [default, auto, disabled]
	ConfigVer string `yaml:"configver" json:"configVer,omitempty"` // specify different version
	Version   string `yaml:"version" json:"version,omitempty"`     // upgrade concern, different logic based on version?
}

func (c ProjectConfig) String() string {
	return fmt.Sprintf("devbranch: %v\nbuildmode: %v\nconfigver: %v\nenable: %v\nversion: %v\n",
		c.S.DevBranch, c.S.BuildMode, c.S.ConfigVer, c.S.Enable, c.S.Version)
}

// type buildmode string

const (
	buildmodeOn       = "on"
	buildmodeAuto     = "auto"
	buildmodeDisabled = "disabled"
	buildmodeManual   = "manual" // for manual build
)

func (p *Project) NeedBuild() bool {
	switch p.Config.S.BuildMode {
	case buildmodeAuto:
		if p.Branch == p.Config.S.DevBranch {
			return true
		}
		return !p.ImageIsExist()
	case buildmodeDisabled:
		return false
	case buildmodeManual:
		return false
	default:
		return true
	}
}
func (p *Project) IsManual() bool {
	return p.Config.S.BuildMode == buildmodeManual
}

// func CheckImageExist() (exist bool, err error) {
// 	tag := p.Branch
// 	if projectpkg.GetEnvFromBranch(p.Branch) == projectpkg.TEST {
// 		tag = p.CommitId
// 	}
// 	return harbor.RepoTagIsExist(p.getprojectpath(), tag)
// }

// so, let's not pass extra commitid, but let branch be commitid
func (p *Project) ImageIsExist() bool {
	exist, err := harbor.RepoTagIsExist(p.Project, p.Branch)
	if err != nil {
		log.Printf("check if image: %v:%v exist err: %v", p.Project, p.Branch, err)
		return false
	}
	return exist
}

func ImageIsExist(project, tag string) (exist bool, err error) {
	return harbor.RepoTagIsExist(project, tag)
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
	if !config.S.Enable && !c.noenablecheck {
		err = fmt.Errorf("project disabled, try do init, if inited, try set selfrelease=enabled")
		return
	}
	// two way to provide config
	// by option setting
	// by project config.yaml
	log.Printf("using configver: %v, devbranch: %v, buildmode: %v", config.S.ConfigVer, config.S.DevBranch, config.S.BuildMode)

	p = &Project{
		Project: project,
		Branch:  c.branch,
		Config:  config, // how to persist this config?
	}

	p.op = *c
	// p.ConfigVer = c.configVer
	// p.DevBranch = c.devBranch

	p.configrepo = configrepo
	// p.repo = repo
	// p.WorkDir = p.repo.GetWorkDir()

	// p.configConfigPath = filepath.Join(p.Project, defaultAppName)

	log.Printf("create project: %q ok\n", project)

	return
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

// func readTemplateConfig(configrepo *git.Repo, configVer string) (p ProjectConfig, err error) {
// 	if configVer == "" {
// 		configVer = GetDefaultConfigVer()
// 	}
// 	f := filepath.Join("template", configVer, defaultConfigYAML)
// 	tyaml, err := configrepo.GetFile(f)
// 	if err != nil {
// 		err = fmt.Errorf("read configrepo templateconfig: %v, err: %v", f, err)
// 		return
// 	}
// 	return decodeConfig(tyaml)
// }

// func readProjectConfig(configrepo *git.Repo, project string) (c ProjectConfig, err error) {
// 	f := getConfigFileName(project)
// 	cyaml, err := configrepo.GetFile(f)
// 	if err != nil {
// 		err = fmt.Errorf("read config file: %v, err: %v", f, err)
// 		return
// 	}
// 	return decodeConfig(cyaml)
// }

// func writeProjectConfig(configrepo *git.Repo, project string, c ProjectConfig) (err error) {
// 	body, err := encodeConfig(c)
// 	if err != nil {
// 		return
// 	}
// 	f := filepath.Join(project, "self-release/config.yaml")
// 	err = configrepo.Add(f, body)
// 	if err != nil {
// 		return
// 	}
// 	text := fmt.Sprintf("setting config.yaml for %v", project)
// 	return configrepo.CommitAndPush(text)
// }

// // unmarshal template config
// func decodeConfig(cyaml []byte) (c ProjectConfig, err error) {
// 	c = ProjectConfig{}
// 	err = yaml.Unmarshal(cyaml, &c)
// 	if err != nil {
// 		err = fmt.Errorf("unmarshal config yaml: %v, err: %v", string(cyaml), err)
// 		return
// 	}
// 	return
// }

// func encodeConfig(c ProjectConfig) (body string, err error) {
// 	b, err := yaml.Marshal(c)
// 	if err != nil {
// 		err = fmt.Errorf("config yaml marshal err: %v", err)
// 		return
// 	}
// 	body = string(b)
// 	return
// }
