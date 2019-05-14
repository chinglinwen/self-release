// template ops relate files
package template

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"wen/self-release/git"

	"github.com/drone/envsubst"
	"github.com/joho/godotenv"
	yaml "gopkg.in/yaml.v2"
)

type File struct {
	Name         string
	Template     string
	Final        string // generated yaml final put into config-deploy?
	RepoTemplate string

	overwrite bool
	perm      os.FileMode
}

// this will be the project config for customizing
type Project struct {
	Project    string
	Env        string // branch
	ConfigFile string // _ops/config.yaml  //set for every env? what's the difference
	Files      []File
	EnvFiles   []string // for setting of template, env.sh ?  no need export

	repo    *git.Repo
	workDir string

	Force    bool
	noUpdate bool
}

// let template store inside repo( rather than config-deploy? )
var defaultFiles = []File{
	{
		Name:         "config",
		Template:     "config.yaml",      //make env specific suffix? or inside
		RepoTemplate: "",                 // it can exist, default no need
		Final:        "_ops/config.yaml", //Project special, just store inside repo
	},
	{
		Name:         "build-docker.sh",
		Template:     "php.v1/build-docker.sh",      //make env specific suffix? or inside
		RepoTemplate: "",                            // it can exist, default no need  // most of the time, it's just template one
		Final:        "projectPath/build-docker.sh", //Project special, just store inside repo
	},
	{
		Name:         "k8s.yaml",
		Template:     "php.v1/k8s.yaml",
		RepoTemplate: "_ops/template/k8s.yaml", // it can exist, default no need
		Final:        "projectPath/k8s.yaml",   // why not _ops/template/k8s.yaml
	},
}

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

// init template file
func (p *Project) Init() (err error) {
	if p.Inited() && !p.Force {

		// it should be by tag? text to force
		return fmt.Errorf("project %v already inited, you can try force init by setting force in the config.yaml", p.Project)
	}

	// copy from template to project repo, later to customize it? generate final by setting
	for _, v := range p.Files {
		if p.repo.IsExist(v.Final) && !v.overwrite {
			err = fmt.Errorf("final file: %v exist", v.Final)
			continue
		}
		tfile, e := configrepo.GetFile(v.Template)
		if e != nil {
			err = fmt.Errorf("get template file: %v err: %v", v.Template, e)
			continue
		}
		if v.RepoTemplate == "" {
			// no need init empty template, repo template is for customize
			// nontheless, put one there? put _ops/template/
			continue
		}
		if v.perm == 0 {
			err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate)
		} else {
			err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate, git.SetPerm(v.perm))
		}
	}
	return
}

// an api call to test?
// using curl? or webpage

// a webpage to trigger the release, manual release

// a webpage to trigger the test

// generate by env setting
func (p *Project) Generate() (err error) {
	// read env
	err = readEnvs(p.EnvFiles)
	if err != nil {
		err = fmt.Errorf("readenvs err: %v", err)
		return
	}

	for _, v := range p.Files {

		template := v.Template
		if v.RepoTemplate != "" {
			template = v.RepoTemplate
		}
		_ = template
		// how to get from template to final

		// https://github.com/drone/envsubst
		// can generate block?
		// use env to overwrite
		finalbody, e := generate(v.Final)
		if err != nil {
			err = fmt.Errorf("generate %v err: %v", v.Final, e)
			// continue
			return
		}
		// write final
		_ = finalbody
	}
	return

}

func generate(file string) (finalbody string, err error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		err = fmt.Errorf("read file: %v,err: %v", file, err)
		return
	}
	return envsubst.EvalEnv(string(b))
}

// https://github.com/joho/godotenv
func readEnvs(files []string) (err error) {
	return godotenv.Load(files...)
}

// // init template
// func Init() {

// }

// fetch config-deploy, no need fetch, let it a pkg call

var (
	configbase = "yunwei/config-deploy"
)

// func SetOverwrite(overwrite bool) func(*Project) {
// 	return func(d *Project) {
// 		d.overwrite = overwrite
// 	}
// }

// template: php.v1/docker/online.yaml  // the name can be anything
// template: php.v1/docker/pre.yaml
// config: _ops/config/templatename.config
// config: _ops/config/config.yaml  //specify which template and which config file?
func NewProject(configyaml string) (p *Project, err error) {
	// final := filepath.Join(project, "Project."+env) // should put into project repo? let them build?
	// p = &Project{
	// 	Project: project,
	// 	Env:     env,
	// 	// Template: template,
	// 	// Config:   config,
	// 	// Final:    final,
	// }

	// we don't want fixed project config template here?
	// let it be template too?
	p = &Project{
		Project: "demo",
	}

	err = yaml.Unmarshal([]byte(configyaml), p)
	if err != nil {
		err = fmt.Errorf("unmarshal project %v, from %v, err: %v", p.Project, configyaml, err)
		return
	}

	for _, v := range defaultFiles {
		if configed(p.Files, v.Name) {
			continue
		}
		p.Files = append(p.Files, v)
	}

	// clone project repo
	if p.noUpdate {
		p.repo, err = git.New(p.Project, git.SetNoPull())
	} else {
		p.repo, err = git.New(p.Project)
	}
	if err != nil {
		log.Println("new err:", err)
		return
	}
	p.workDir = p.repo.GetWorkDir()
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
