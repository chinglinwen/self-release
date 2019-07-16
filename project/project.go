// template ops relate files
package project

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"wen/self-release/git"
	"wen/self-release/pkg/harbor"

	yaml "gopkg.in/yaml.v2"
)

var (
	defaultConfigBase = "yunwei/config-deploy"
	defaultAppName    = "self-release"

	defaultConfigName = "config.env" // later will prefix with default or customize version
	defaultConfigYAML = "config.yaml"
	// opsDir         = "_ops"
	defaultRepoConfigPath = "_ops" // configpath becomes project path in config-deploy
)

// type File struct {
// 	Name         string
// 	Template     string
// 	Final        string // generated yaml final put into config-deploy?
// 	RepoTemplate string

// 	Overwrite         bool
// 	Perm              os.FileMode // set final file perm
// 	ValidateFinalYaml bool        // `yaml:'validateFinalYaml'`
// }

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
	DevBranch   string `yaml:"devbranch"`   // default dev branch name
	BuildMode   string `yaml:"buildmode"`   // used to disable auto build [default, auto, disabled]
	ConfigVer   string `yaml:"configver"`   // specify different version
	SelfRelease string `yaml:"selfrelease"` // enable or not for self-release on this project
	Version     string `yaml:"version"`     // currently don't need?
}

func (c ProjectConfig) String() string {
	return fmt.Sprintf("devbranch: %v\nbuildmode: %v\nconfigver: %v\nselfrelease: %v\nversion: %v\n",
		c.DevBranch, c.BuildMode, c.ConfigVer, c.SelfRelease, c.Version)
}

// type buildmode string

const (
	buildmodeOn       = "on"
	buildmodeAuto     = "auto"
	buildmodeDisabled = "disabled"
	buildmodeManual   = "manual" // for manual build
)

func (p *Project) NeedBuild() bool {
	switch p.Config.BuildMode {
	case buildmodeAuto:
		if p.Branch == p.Config.DevBranch {
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
	return p.Config.BuildMode == buildmodeManual
}

func (p *Project) ImageIsExist() bool {
	exist, err := harbor.RepoTagIsExist(p.Project, p.Branch)
	if err != nil {
		log.Printf("check if image: %v:%v exist err: %v", p.Project, p.Branch, err)
		return false
	}
	return exist
}

// // let template store inside repo( rather than config-deploy? )
// var defaultFiles = []File{
// 	{
// 		Name:         "config",
// 		Template:     "config.yaml",      //make env specific suffix? or inside
// 		RepoTemplate: "",                 // it can exist, default no need
// 		Final:        "_ops/config.yaml", //Project special, just store inside repo
// 	},
// 	{
// 		Name:         "build-docker.sh",
// 		Template:     "php.v1/build-docker.sh",      //make env specific suffix? or inside
// 		RepoTemplate: "",                            // it can exist, default no need  // most of the time, it's just template one
// 		Final:        "projectPath/build-docker.sh", //Project special, just store inside repo
// 	},
// 	{
// 		Name:         "k8s.yaml",
// 		Template:     "php.v1/k8s.yaml",
// 		RepoTemplate: "_ops/template/k8s.yaml", // it can exist, default no need
// 		Final:        "projectPath/k8s.yaml",   // why not _ops/template/k8s.yaml
// 	},
// }

// func configed(files []File, name string) bool {
// 	for _, v := range files {
// 		if v.Name == name {
// 			return true
// 		}
// 	}
// 	return false
// }

func BranchIsTag(branch string) bool {
	return git.BranchIsTag(branch)
}

// func (p *Project) Inited() bool {
// 	if p != nil {
// 		if p.repo != nil {
// 	return p.repo.IsExist("_ops/config.yaml")
// 		}
// 	}
// 	return false
// }
func (p *Project) Inited() bool {
	if p != nil {
		config := filepath.Join(p.Project, "self-release", defaultConfigName) // relate to initk8s path
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

// func (p *Project) GetRepo() *git.Repo {
// 	return p.repo
// }

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

// type initOption struct {
// 	autoenv map[string]string
// }

// func SetAutoEnv(autoenv map[string]string) func(*Project) {
// 	return func(o *Project) {
// 		o.autoenv = autoenv
// 	}
// }

// let people replace with block?

// they need manual edit?

// we just generate one final (and may never change, unless overwrite(have a backup though)
//this way they can customize the final?  using diff?

// template: php.v1/docker/online.yaml  // the name can be anything
// template: php.v1/docker/pre.yaml
// config: _ops/config/templatename.config
// config: _ops/config/config.yaml  //specify which template and which config file?
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

	}
	c := &projectOption{
		// branch:    "master",
		// devBranch: "develop",
		// buildMode: "default",
		configVer: GetDefaultConfigVer(),
	}
	for _, op := range options {
		op(c)
	}
	log.Printf("project: %v, options: %#v\n", project, c)

	// p = &Project{
	// 	Project: project, // "template-before-create",
	// 	// Branch:    "master", // TODO: default to master?
	// 	// ConfigVer: GetDefaultConfigVer(),
	// 	// DevBranch: "develop", // default dev branch
	// }
	// // log.Printf("before options apply for repo: %q ok\n", p.Project)

	// if p.Branch == "" {
	// 	p.Branch = "master"
	// }
	// if p.ConfigVer == "" {
	// 	p.ConfigVer = GetDefaultConfigVer()
	// }
	// if p.DevBranch == "" {
	// 	p.DevBranch = "develop"
	// }

	// // // p variable will change multiple times, save the variable here
	// // autoenv := p.autoenv

	// // log.Printf("after options apply for repo: %q ok\n", p.Project)

	// branch := p.Branch
	// configVer := p.ConfigVer
	// // force := p.InitForce

	// // normal repo config take first
	// repo, e := getRepo(project, c.branch, c.nopull)
	// if e != nil {
	// 	err = fmt.Errorf("clone or open project: %v, err: %v, configver: %v", project, e, c.configVer)
	// 	return
	// }

	// try get config, to overwrite default config
	config, err := readProjectConfig(configrepo, project)
	if err != nil && !c.configMustExist {
		log.Println("read project config err, will using default config for ", project)
		defaultConfig, e := readTemplateConfig(configrepo, c.configVer) // using default config, can we get configver now?
		if e != nil {
			err = fmt.Errorf("get defaultConfig from config-repo err: %v", e)
			return
		}
		config = defaultConfig
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("read config failed, config may not exist, err: %v", err)
		return
	}
	if config.SelfRelease != "enabled" && !c.noenablecheck {
		err = fmt.Errorf("project disabled, try do init, if inited, try set selfrelease=enabled")
		return
	}
	// two way to provide config
	// by option setting
	// by project config.yaml
	log.Printf("using configver: %v, devbranch: %v, buildmode: %v", config.ConfigVer, config.DevBranch, config.BuildMode)

	p = &Project{
		Project: project,
		Branch:  c.branch,
		Config:  config, // how to persist this config?
	}

	// we don't need config.yaml anymore
	// if !c.noreadconfig {
	// 	p, err = readProjectConfig(configrepo, project)
	// 	if err != nil {
	// 		// // not inited, using template config? or just return error,since it not inited?
	// 		// tp, e := readTemplateConfig(configVer)
	// 		// if e != nil {
	// 		// 	err = fmt.Errorf("readTemplateConfig for project: %v, err: %v, configver: %v", project, e, configVer)
	// 		// 	return
	// 		// }
	// 		// p = tp

	// 		// // only except we don't write files to git?

	// 		// // it can't be, project name have issues too?
	// 		// // what others setting will be overwrite by template?
	// 		// p.Project = project
	// 		// p.Branch = branch

	// 		// log.Printf("set to default config for project %q\n", project)
	// 		// err = fmt.Errorf("project %v not inited, for branch: %v", project, branch)
	// 		log.Printf("project %v not inited, for branch: %v", project, c.branch)
	// 		return
	// 	}

	// 	log.Printf("reading project config for repo: %v, branch: %v ok\n", project, c.branch)
	// } else {
	// 	p = &Project{
	// 		Project: project,
	// 	}
	// }

	// var tp *Project

	// spew.Dump("template config:", tp.Files)

	// log.Printf("try read templateconfig for repo: %q ok\n", p.Project)

	// this will overwrite option setting?
	// if force {
	// 	// force ignore repo config
	// 	p = tp
	// } else {

	// }
	// // repo config exist, merge config, is this needed?
	// if we all come from init, it's likely that files is appending
	// for _, v := range tp.Files {
	// 	if configed(p.Files, v.Name) {
	// 		continue
	// 	}
	// 	p.Files = append(p.Files, v)
	// }

	// clone project repo
	// if p.NoPull {
	// 	p.repo, err = git.New(p.Project, git.SetBranch(p.Branch), git.SetForce())
	// } else {
	// 	p.repo, err = git.NewWithPull(p.Project, git.SetBranch(p.Branch), git.SetForce())
	// }
	// if err != nil {
	// 	err = fmt.Errorf("git clone err: %v for project: %v", err, p.Project)
	// 	return
	// }

	p.op = *c
	// p.ConfigVer = c.configVer
	// p.DevBranch = c.devBranch

	p.configrepo = configrepo
	// p.repo = repo
	// p.WorkDir = p.repo.GetWorkDir()

	p.configConfigPath = filepath.Join(defaultAppName, p.Project)

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
	// p = &Project{
	// 	Project: project,
	// }
	// log.Printf("try gitnew for repo: %q ok\n", p.Project)

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

func readTemplateConfig(configrepo *git.Repo, configVer string) (p ProjectConfig, err error) {
	if configVer == "" {
		configVer = GetDefaultConfigVer()
	}
	f := filepath.Join("template", configVer, defaultConfigYAML)
	tyaml, err := configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("read configrepo templateconfig: %v, err: %v", f, err)
		return
	}
	return decodeConfig(tyaml)
}

// func readTemplateConfig(configrepo *git.Repo, configVer string) (p ProjectConfig, err error) {
// 	if configVer == "" {
// 		configVer = GetDefaultConfigVer()
// 	}
// 	f := filepath.Join("template", configVer, defaultConfigName)
// 	tyaml, err := configrepo.GetFile(f)
// 	if err != nil {
// 		err = fmt.Errorf("read configrepo templateconfig: %v, err: %v", f, err)
// 		return
// 	}
// 	return parseConfig(tyaml)
// }

func readProjectConfig(configrepo *git.Repo, project string) (c ProjectConfig, err error) {
	f := filepath.Join(project, defaultAppName, defaultConfigYAML)
	cyaml, err := configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("read config file: %v, err: %v", f, err)
		return
	}
	return decodeConfig(cyaml)
}

func writeProjectConfig(configrepo *git.Repo, project string, c ProjectConfig) (err error) {
	body, err := encodeConfig(c)
	if err != nil {
		return
	}
	f := filepath.Join(project, "self-release/config.yaml")
	err = configrepo.Add(f, body)
	if err != nil {
		return
	}
	text := fmt.Sprintf("setting config.yaml for %v", project)
	return configrepo.CommitAndPush(text)
}

// let's pass configrepo?
// func readProjectConfig(configrepo *git.Repo, project string) (p *Project, err error) {
// 	// configrepo, err := GetConfigRepo()
// 	// if err != nil || configrepo == nil {
// 	// 	err = fmt.Errorf("read config from configrepo err: %v", err)
// 	// 	return
// 	// }
// 	f := filepath.Join(project, defaultConfigName)
// 	cyaml, err := configrepo.GetFile(f)
// 	if err != nil {
// 		err = fmt.Errorf("read config file: %v, err: %v", f, err)
// 		return
// 	}
// 	return parseConfig(cyaml)
// }

// func readRepoConfig(repo *git.Repo) (p *Project, err error) {
// 	if repo == nil {
// 		err = fmt.Errorf("read config for err: repo not clone or open yet")
// 		return
// 	}
// 	f := filepath.Join(defaultRepoConfigPath, defaultConfigName)
// 	cyaml, err := repo.GetFile(f)
// 	if err != nil {
// 		err = fmt.Errorf("read config file: %v, err: %v", f, err)
// 		return
// 	}
// 	return parseConfig(cyaml)
// }

// unmarshal template config
func decodeConfig(cyaml []byte) (c ProjectConfig, err error) {
	c = ProjectConfig{}
	err = yaml.Unmarshal(cyaml, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal config yaml: %v, err: %v", string(cyaml), err)
		return
	}
	return
}

func encodeConfig(c ProjectConfig) (body string, err error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		err = fmt.Errorf("config yaml marshal err: %v", err)
		return
	}
	body = string(b)
	return
}
