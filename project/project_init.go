package project

import (
	"fmt"
	"path/filepath"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
)

// a webpage to trigger the test
type initOption struct {
	singleName string
	force      bool
	configVer  string
}

// force is used to re-init config.yaml
func SetInitForce() func(*initOption) {
	return func(p *initOption) {
		p.force = true
	}
}

func SetInitName(name string) func(*initOption) {
	return func(o *initOption) {
		o.singleName = name
	}
}

func SetInitVersion(ver string) func(*initOption) {
	return func(o *initOption) {
		o.configVer = ver
	}
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
func (p *Project) Init(options ...func(*initOption)) (err error) {
	c := &initOption{}
	for _, op := range options {
		op(c)
	}

	if p.Inited() && !c.force {
		// it should be by tag? text to force
		return fmt.Errorf("project %v already inited, you can try force init by setting force in the config.yaml", p.Project)
	}

	// can we fix first init issue? need two times of init?
	// let's read env config first, it seems we read config first, no need two time of read?
	// only if there have no config before init

	// set envs
	// init file using template cause re-init much problem? changes maybe lost?
	// say build-docker should not using env, as init(static config.env) have no projects info?

	// we currently ignore autoenv, only config env is working for init
	envMap, err := p.readEnvs(nil) // only re-init is working, otherwise it's just not exist
	if err != nil {
		err = fmt.Errorf("readenvs err: %v", err)
	}

	errs := make(initErr)
	found := false

	// copy from template to project repo, later to customize it? generate final by setting
	for _, v := range p.Files {

		// // init should only concern with config.yaml? init need includes repotemplate
		// if v.Name != "config.yaml" {
		// 	continue
		// }

		if c.singleName != "" {
			if c.singleName != v.Name { // try support filename match?
				// mostly specify file to init single file, so continue
				continue
			}
		}

		if v.RepoTemplate == "" && v.Final == "" {
			err = fmt.Errorf("repotemplate and final file not specified for %v", v.Name)
			errs[v.Name] = err
			continue
		}
		found = true

		// init only init final or repotemplate, not both

		// === generate repo template parts( if not ovewwrite, custom setting will be keeped)
		err = p.initRepoTemplateOrFinal(c.force, v, envMap)
		if err != nil {
			err = fmt.Errorf("initRepoTemplateOrFinal project: %v file: %v err: %v", p.Project, v.RepoTemplate, err)
			errs[v.Name] = err
			continue
		}

		// if p.repo.IsExist(v.Final) && !v.Overwrite && !p.Force {
		// 	err = fmt.Errorf("final file: %v exist and force or overwrite not set, skip", v.Final)
		// 	errs[v.Name] = err
		// 	continue
		// }

		// // check file setting format is valid? say v.template is empty
		// if v.Template == "" {
		// 	err = fmt.Errorf("template file not specified for %v", v.Name)
		// 	errs[v.Name] = err
		// 	continue
		// }

		// f := filepath.Join("template", v.Template) // prefix template for template
		// tfile, e := configrepo.GetFile(f)
		// if e != nil {
		// 	err = fmt.Errorf("get template file: %v err: %v", f, e)
		// 	errs[v.Name] = err
		// 	continue
		// }

		// // if no variable to replace or no custom setting, no need to init repotemplate?
		// if v.RepoTemplate == "" {
		// 	// no need init empty template, repo template is for customize
		// 	// nontheless, put one there? put _ops/template/
		// 	log.Println("RepoTemplate is empty, skip init for file:", v.Name)
		// 	continue
		// }

		// log.Println("creating init file:", v.Name)
		// if v.Perm == 0 {
		// 	err = p.repo.Add(v.RepoTemplate, string(tfile))
		// } else {
		// 	err = p.repo.Add(v.RepoTemplate, string(tfile), git.SetPerm(v.Perm))
		// }

		// if v.perm == 0 {
		// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate)
		// } else {
		// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate, git.SetPerm(v.perm))
		// }

		// // how to init final?, we don't init final, we generate final in later steps
	}
	if !found {
		err = fmt.Errorf("init for %v, err: not found item for the init in config", p.Project)
		return
	}
	if len(errs) != 0 {
		return &errs
	}
	err = p.CommitAndPush("init config.yaml")
	if err != nil {
		err = fmt.Errorf("init push err: %v, project: %v", err, p.Project)
		return
	}

	return
}

// if no variable to replace or no custom setting, no need to init repotemplate?
// gen or add to git?  // why not generate once
func (p *Project) initRepoTemplateOrFinal(force bool, v File, envMap map[string]string) (err error) {
	var initfile string
	var evaltemplate bool
	var exist bool
	if v.RepoTemplate != "" {
		initfile = v.RepoTemplate // init repotemplate, later generate final
	} else {
		initfile = v.Final
		evaltemplate = true
	}
	exist = p.repo.IsExist(initfile)

	// force should only for config.yaml and repo, force for redo of  init
	// if exist && v.Name == "config.yaml" && !p.InitForce {
	// 	log.Printf("init file: %v exist and force have not set, skip", v.Final)
	// 	return
	// }
	if exist && !v.Overwrite && !force {
		log.Printf("init file: %v exist and force or overwrite have not set, skip", v.Final)
		return
	}

	// get config template
	f := filepath.Join("template", v.Template) // prefix template for template
	tfile, e := configrepo.GetFile(f)
	if e != nil {
		err = fmt.Errorf("get configtemplate file: %v err: %v", f, e)
		return
	}
	var tbody string
	var note string
	if evaltemplate {
		tbody, err = generateByMap(string(tfile), envMap)
		if err != nil {
			err = fmt.Errorf("get configtemplate file: %v err: %v", f, err)
			return
		}
		note = "(generated)"
	} else {
		tbody = string(tfile)
	}

	if v.Perm == 0 {
		err = p.repo.Add(initfile, string(tbody))
	} else {
		err = p.repo.Add(initfile, string(tbody), git.SetPerm(v.Perm))
	}

	log.Printf("inited file: %v%v, project: %v\n", initfile, note, p.Project)

	return
}

// func (p *Project) initFinal(v File) (err error) {

// }

// fetch config-deploy, no need fetch, let it a pkg call

// func SetOverwrite(overwrite bool) func(*Project) {
// 	return func(d *Project) {
// 		d.overwrite = overwrite
// 	}
// }

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
