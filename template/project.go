// template ops relate files
package template

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"wen/self-release/git"

	yaml "gopkg.in/yaml.v2"
)

type File struct {
	Name         string
	Template     string
	Final        string // generated yaml final put into config-deploy?
	RepoTemplate string

	Overwrite bool
	Perm      os.FileMode
}

// this will be the project config for customizing
type Project struct {
	Project string
	Branch  string // build branch
	Env     string // branch, may derive from event's branch as env
	// not able to get branch? we can, but if it's a tag? init for develop branch only no tags
	DevBranch string // default dev branch name

	ConfigFile string // _ops/config.yaml  //set for every env? what's the difference
	Files      []File
	EnvFiles   []string // for setting of template, env.sh ?  no need export

	Force     bool   // force for re-init, file setting, we often setting it by tag msg
	NoPull    bool   // `yaml:"nopull"`
	ConfigVer string // specify different version

	// image ?
	// size of replicas?

	repo    *git.Repo
	workDir string
	envMap  map[string]string
	autoenv map[string]string // env from hook
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

func configed(files []File, name string) bool {
	for _, v := range files {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (p *Project) Inited() bool {
	return p.repo.IsExist("config.yaml")
}

func SetInitForce() func(*Project) {
	return func(p *Project) {
		p.Force = true
	}
}

func SetBranch(branch string) func(*Project) {
	return func(p *Project) {
		p.Branch = branch
	}
}

func SetAutoEnv(autoenv map[string]string) func(*Project) {
	return func(p *Project) {
		p.autoenv = autoenv
	}
}

func SetInitVersion(ver string) func(*Project) {
	return func(p *Project) {
		p.ConfigVer = ver
	}
}

// let people replace with block?

// they need manual edit?

// we just generate one final (and may never change, unless overwrite(have a backup though)
//this way they can customize the final?  using diff?

// template: php.v1/docker/online.yaml  // the name can be anything
// template: php.v1/docker/pre.yaml
// config: _ops/config/templatename.config
// config: _ops/config/config.yaml  //specify which template and which config file?
func NewProject(project string, options ...func(*Project)) (p *Project, err error) {
	if configrepo == nil {
		err = fmt.Errorf("configrepo not inited")
		return
	}
	p = &Project{
		Project:   project,  // "template-before-create",
		Branch:    "master", // TODO: default to master?
		ConfigVer: GetDefaultConfigVer(),
		DevBranch: "develop", // default dev branch
	}
	// log.Printf("before options apply for repo: %q ok\n", p.Project)
	for _, op := range options {
		op(p)
	}
	// log.Printf("after options apply for repo: %q ok\n", p.Project)

	branch := p.Branch
	env := p.Env
	configVer := p.ConfigVer
	force := p.Force

	log.Printf("try read templateconfig for repo: %q ok\n", p.Project)

	var tp *Project
	// read for later merge template files setting?
	tp, err = readTemplateConfig(configVer)
	if err != nil {
		err = fmt.Errorf("readTemplateConfig err: %v, configver: %v", err, configVer)
		return
	}
	// spew.Dump("template config:", tp.Files)

	log.Printf("after read templateconfig for repo: %q ok\n", p.Project)

	if force {
		// force ignore repo config
		p = tp
	} else {
		log.Printf("try read repoconfig for repo: %q ok\n", p.Project)

		// normal repo config take first
		p, err = readRepoConfig(project, p.Branch)
		if err != nil {
			p = tp // if not inited, using default project setting from default template

			// only except we don't write files to git?

			// it can't be, project name have issues too?
			// what others setting will be overwrite by template?
			p.Project = project
			p.Branch = branch
			p.Env = env
			log.Printf("set to default config for project %q\n", project)
		}
	}
	// // repo config exist, merge config, is this needed?
	// if we all come from init, it's likely that files is appending
	// for _, v := range tp.Files {
	// 	if configed(p.Files, v.Name) {
	// 		continue
	// 	}
	// 	p.Files = append(p.Files, v)
	// }
	log.Printf("create repo: %q ok\n", p.Project)

	// clone project repo
	if p.NoPull {
		p.repo, err = git.New(p.Project, git.SetBranch(p.Branch), git.SetForce(), git.SetNoPull())
	} else {
		p.repo, err = git.New(p.Project, git.SetBranch(p.Branch), git.SetForce())
	}
	if err != nil {
		err = fmt.Errorf("git clone err: %v for project: %v", err, p.Project)
		return
	}

	p.workDir = p.repo.GetWorkDir()
	return
}

type initErr map[string]error

func (errs *initErr) Error() (s string) {
	for k, v := range *errs {
		s = fmt.Sprintf("%v\nname: %v, init err: %v", s, k, v)
	}
	return
}

// init can reading from repo's config, or assume have project name only(using default config version)
//
// init template file, config.yaml and repotemplate files
func (p *Project) Init(options ...func(*Project)) (err error) {

	for _, op := range options {
		op(p)
	}

	if p.Inited() && !p.Force {

		// it should be by tag? text to force
		return fmt.Errorf("project %v already inited, you can try force init by setting force in the config.yaml", p.Project)
	}

	errs := make(initErr)

	// copy from template to project repo, later to customize it? generate final by setting
	for _, v := range p.Files {

		// init should only concern with config.yaml?
		if v.Name != "config.yaml" {
			continue
		}

		if p.repo.IsExist(v.Final) && !v.Overwrite && !p.Force {
			err = fmt.Errorf("final file: %v exist and force or overwrite not set, skip", v.Final)
			errs[v.Name] = err
			continue
		}

		// check file setting format is valid? say v.template is empty
		if v.Template == "" {
			err = fmt.Errorf("template file not specified for %v", v.Name)
			errs[v.Name] = err
			continue
		}

		f := filepath.Join("template", v.Template) // prefix template for template
		tfile, e := configrepo.GetFile(f)
		if e != nil {
			err = fmt.Errorf("get template file: %v err: %v", f, e)
			errs[v.Name] = err
			continue
		}

		// if no variable to replace or no custom setting, no need to init repotemplate?
		if v.RepoTemplate == "" {
			// no need init empty template, repo template is for customize
			// nontheless, put one there? put _ops/template/
			log.Println("RepoTemplate is empty, skip init for file:", v.Name)
			continue
		}

		log.Println("creating init file:", v.Name)
		if v.Perm == 0 {
			err = p.repo.Add(v.RepoTemplate, string(tfile))
		} else {
			err = p.repo.Add(v.RepoTemplate, string(tfile), git.SetPerm(v.Perm))
		}

		// if v.perm == 0 {
		// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate)
		// } else {
		// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate, git.SetPerm(v.perm))
		// }

		// // how to init final?, we don't init final, we generate final in later steps
	}
	if len(errs) != 0 {
		return &errs
	}

	return
}

// // init template
// func Init() {

// }

// fetch config-deploy, no need fetch, let it a pkg call

var (
	configBase = "yunwei/config-deploy"

	configName = "config.yaml" // later will prefix with default or customize version
	// opsDir         = "_ops"
	repoConfigPath = "_ops"
)

// func SetOverwrite(overwrite bool) func(*Project) {
// 	return func(d *Project) {
// 		d.overwrite = overwrite
// 	}
// }

func readTemplateConfig(configVer string) (p *Project, err error) {
	p = &Project{
		Project: "template-config",
		// ConfigVer: configVer,
	}

	f := filepath.Join("template", configVer, configName)
	tyaml, err := configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("read configrepo for project: %v, templateconfig: %v, err: %v", p.Project, f, err)
		return
	}
	// unmarshal template config
	err = yaml.Unmarshal(tyaml, p)
	if err != nil {
		err = fmt.Errorf("unmarshal config for project %v, from %v, err: %v", p.Project, string(tyaml), err)
		return
	}
	return
}

func readRepoConfig(project, branch string) (p *Project, err error) {
	p = &Project{
		Project: project,
	}
	log.Printf("try gitnew for repo: %q ok\n", p.Project)

	p.repo, err = git.New(p.Project, git.SetBranch(branch), git.SetNoPull()) // TODO: nopull?
	if err != nil {
		err = fmt.Errorf("clone git repo for: %v, err: %v", p.Project, err)
		return
	}

	f := filepath.Join(repoConfigPath, configName)
	cyaml, err := configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("read config for project: %v, config: %v, err: %v", p.Project, f, err)
		return
	}
	// unmarshal template config
	err = yaml.Unmarshal(cyaml, p)
	if err != nil {
		err = fmt.Errorf("unmarshal config for project %v, from %v, err: %v", p.Project, string(cyaml), err)
		return
	}
	return
}

// // see if project's path has Project
// func (d *Project) Generate() (err error) {
// 	if IsExist(d.Final) && !d.overwrite {
// 		return
// 	}
// 	Copy(d.Template, d.Final)

// 	// do we need some customization
// 	// do the customization and verify later?

// 	// only copy the template? later customze it

// 	return d.Push()
// }

// func (d *Project) Push() (err error) {
// 	err = repo.AddFileAndPush(d.Final, fmt.Sprintf("generate %v", d.Final))
// 	if err != nil {
// 		return fmt.Errorf("push file: %v, err: %v\n", d.Final, err)
// 	}
// 	return
// }

// copoy template to project path

// verify it's working
// final result store into _ops/final?  store in config-deploy only?

// no easy way to merge manual part? ( yaml can be )

// we only do generate once ( but may repeat many time, template is good enough, with overwrite setting )
