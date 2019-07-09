package project

import (
	"errors"
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
	branch     string // do we need this?
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

// mostly branch is develop
func SetInitBranch(branch string) func(*initOption) {
	return func(o *initOption) {
		o.branch = branch
	}
}

type errlist map[string]error

func (errs *errlist) Error() (s string) {
	for k, v := range *errs {
		s = fmt.Sprintf("%v\nname: %v, init err: %v", s, k, v)
	}
	return
}

// we use no err to signal update
// we need to filter out nochange err (but not all)
//
// no err update
// single no change update
// all nochange no update
// single err? just return?
func (errs *errlist) Nochange() bool {
	if errs != nil && len(*errs) == 0 {
		return false
	}
	for _, v := range *errs {
		if v != ErrNoChange {
			return false
		}
	}
	return true // all err is ErrNoChange
}

// init can reading from repo's config, or assume have project name only(using default config version)?
//
// init template file, config.yaml and repotemplate files
// if repo config.yaml exist, it will affect init process?
func (p *Project) Init(options ...func(*initOption)) (err error) {
	c := initOption{branch: "develop"}
	for _, op := range options {
		op(&c)
	}
	log.Printf("got init option: %#v", c)

	if p.Inited() && !c.force {
		// it should be by tag? text to force
		err = fmt.Errorf("project %v already inited, you can try force init", p.Project)
		return
	}
	err = p.initAll(c)
	return
}

// errs := make(errlist)
// found := false

// var updateconfigrepo bool

// // not inited, using template config
// tp, e := readTemplateConfig(p.configrepo, c.configVer)
// if e != nil {
// 	err = fmt.Errorf("readTemplateConfig for project: %v, err: %v, configver: %v", p.Project, e, c.configVer)
// 	return
// }

// log.Printf("read template main config to init project %q\n", p.Project)

// /*
//   - name: config.yaml
//     template: php.v1/config.yaml
//     final: config:self-release/config.yaml
//   - name: config.env
//     template: php.v1/config.env
//     final: config:self-release/config.env
// */

// // we add item here, so remove two config items, to simplify
// files := []File{
// 	{Name: "config.yaml", Template: GetDefaultConfigVer() + "/config.yaml", Final: "config:self-release/config.yaml"},
// 	{Name: "config.env", Template: GetDefaultConfigVer() + "/config.env", Final: "config:self-release/config.env"},
// }
// files = append(files, tp.Files...)

// // copy from template to project repo, later to customize it? generate final by setting
// for _, v := range files {

// 	// // init should only concern with config.yaml? init need includes repotemplate
// 	// if v.Name != "config.yaml" {
// 	// 	continue
// 	// }

// 	if c.singleName != "" {
// 		if c.singleName != v.Name { // try support filename match?
// 			// mostly specify file to init single file, so continue
// 			continue
// 		}
// 	}

// 	if v.RepoTemplate == "" && v.Final == "" {
// 		err = fmt.Errorf("repotemplate and final file not specified for %v", v.Name)
// 		errs[v.Name] = err
// 		continue
// 	}
// 	found = true

// 	// init only init final or repotemplate, not both

// 	// === generate repo template parts( if not ovewwrite, custom setting will be keeped)
// 	updateconfig, e := p.initRepoTemplateOrFinal(p.configrepo, c.force, v, envMap)
// 	if e != nil {
// 		err = fmt.Errorf("initRepoTemplateOrFinal project: %v file: %v err: %v", p.Project, v.RepoTemplate, e)
// 		errs[v.Name] = err
// 		continue
// 	}
// 	if updateconfig {
// 		// if one item update exist, commit it
// 		updateconfigrepo = true
// 	}
// 	// if p.repo.IsExist(v.Final) && !v.Overwrite && !p.Force {
// 	// 	err = fmt.Errorf("final file: %v exist and force or overwrite not set, skip", v.Final)
// 	// 	errs[v.Name] = err
// 	// 	continue
// 	// }

// 	// // check file setting format is valid? say v.template is empty
// 	// if v.Template == "" {
// 	// 	err = fmt.Errorf("template file not specified for %v", v.Name)
// 	// 	errs[v.Name] = err
// 	// 	continue
// 	// }

// 	// f := filepath.Join("template", v.Template) // prefix template for template
// 	// tfile, e := configrepo.GetFile(f)
// 	// if e != nil {
// 	// 	err = fmt.Errorf("get template file: %v err: %v", f, e)
// 	// 	errs[v.Name] = err
// 	// 	continue
// 	// }

// 	// // if no variable to replace or no custom setting, no need to init repotemplate?
// 	// if v.RepoTemplate == "" {
// 	// 	// no need init empty template, repo template is for customize
// 	// 	// nontheless, put one there? put _ops/template/
// 	// 	log.Println("RepoTemplate is empty, skip init for file:", v.Name)
// 	// 	continue
// 	// }

// 	// log.Println("creating init file:", v.Name)
// 	// if v.Perm == 0 {
// 	// 	err = p.repo.Add(v.RepoTemplate, string(tfile))
// 	// } else {
// 	// 	err = p.repo.Add(v.RepoTemplate, string(tfile), git.SetPerm(v.Perm))
// 	// }

// 	// if v.perm == 0 {
// 	// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate)
// 	// } else {
// 	// 	err = p.repo.AddAndPush(v.RepoTemplate, string(tfile), "init "+v.RepoTemplate, git.SetPerm(v.perm))
// 	// }

// 	// // how to init final?, we don't init final, we generate final in later steps
// }
// if !found {
// 	err = fmt.Errorf("init for %v, err: not found item for the init in config", p.Project)
// 	return
// }
// if len(errs) != 0 {
// 	err = &errs
// 	return
// }
// err = p.CommitAndPush("init config.yaml for " + p.Project)
// if err != nil {
// 	err = fmt.Errorf("init push err: %v, project: %v", err, p.Project)
// 	return
// }
// if updateconfigrepo {
// 	p.configrepo.CommitAndPush("init config.yaml for " + p.Project)
// 	if err != nil {
// 		err = fmt.Errorf("init push err: %v, project: %v", err, p.Project)
// 		return
// 	}
// }

// return
// }

/*
  # php
  - name: php.ini
    template: php.v1/php.ini
    final: _ops/php.ini
    #overwrite: true
  - name: nginx.conf
    template: php.v1/nginx.conf
    final: _ops/nginx.conf
    #overwrite: true
  # docker
  - name: dockerfile
    template: php.v1/Dockerfile
    final: Dockerfile
    #overwrite: true
  - name: build-docker.sh
    template: php.v1/build-docker.sh
    final: build-docker.sh
    overwrite: true
  # support existing ci
  - name: gitlab-ci.yml
    template: php.v1/.gitlab-ci.yml
    final: .gitlab-ci.yml
    #overwrite: true
*/
// copy to config
// copy to repo

// separate initall for easier operate, init docker only ( aka project repo only )
// human can easily intercept and fix if there's error ( since it's only about docker image )
//
// we still do init k8s relate, but it's optional, it can skip
func (p *Project) initAll(c initOption) (err error) {

	// we currently ignore autoenv, only config env is working for init
	envMap, err := p.readEnvs(nil)
	if err != nil {
		err = fmt.Errorf("readenvs err: %v", err)
	}
	changed1, err := p.initDocker(envMap, c)
	if err != nil {
		err = fmt.Errorf("initDocker err: %v", err)
		return
	}
	changed2, err := p.initK8s(envMap, c)
	if err != nil {
		err = fmt.Errorf("initK8s err: %v", err)
		return
	}
	if !changed1 && !changed2 {
		log.Println("both repo and configrepo have no change")
		return nil
	}
	return
}

// init docker just an optional steps
// require build-docker.sh exist, if using self-release to build image
func (p *Project) initDocker(envMap map[string]string, c initOption) (update bool, err error) {
	items := []struct {
		src, dst string
	}{
		{src: "php.ini", dst: "ops/php.ini"},
		{src: "nginx.conf", dst: "ops/nginx.conf"},
		{src: "Dockerfile", dst: "Dockerfile"},
		{src: "build-docker.sh", dst: "build-docker.sh"},
		// {src: ".gitlab-ci.yml", dst: ".gitlab-ci.yml"},  // so much other files to generate too
	}
	var changed bool
	for _, v := range items {
		src := filepath.Join("template", p.Config.ConfigVer, v.src)
		changed, err = p.CopyToRepo(src, v.dst, envMap)
		if err != nil {
			err = fmt.Errorf("copytoconfig err: %v", err)
		}
		if changed {
			log.Printf("file: %v will be updated", v.dst)
			update = true
		}
	}
	if !update {
		log.Println("docker init have no change")
		return
	}
	err = commitandpush(p.repo, "init docker files by self-release")
	return
}

/*
  - name: k8s-online
    template: php.v1/k8s/template-online.yaml
    repotemplate: config:self-release/template/template-online.yaml
    final: config:self-release/k8s-online.yaml
    validatefinalyaml: true
  - name: k8s-pre
    template: php.v1/k8s/template-pre.yaml
    repotemplate: config:self-release/template/template-pre.yaml
    final: config:self-release/k8s-pre.yaml
    validatefinalyaml: true
  - name: k8s-test
    template: php.v1/k8s/template-test.yaml
    repotemplate: config:self-release/template/template-test.yaml
    final: config:self-release/k8s-test.yaml
	validatefinalyaml: true
*/
func (p *Project) initK8s(envMap map[string]string, c initOption) (update bool, err error) {
	items := []struct {
		src, dst string
	}{
		{src: "k8s/template-online.yaml", dst: "self-release/template/template-online.yaml"},
		{src: "k8s/template-pre.yaml", dst: "self-release/template/template-pre.yaml"},
		{src: "k8s/template-test.yaml", dst: "self-release/template/template-test.yaml"},
		{src: "config.env", dst: "self-release/config.env"},
		{src: "config.yaml", dst: "self-release/config.yaml"}, // should we add this?
	}
	var changed bool
	for _, v := range items {
		src := filepath.Join("template", p.Config.ConfigVer, v.src)
		dst := filepath.Join(p.Project, v.dst)
		if c.force {
			changed, err = p.CopyToConfigNoGenForce(src, dst, envMap)
		} else {
			changed, err = p.CopyToConfigNoGen(src, dst, envMap)
		}
		if err != nil {
			err = fmt.Errorf("copytoconfig err: %v", err)
			return
		}
		if changed {
			log.Printf("file: %v will be updated", v.dst)
			update = true
		}
	}
	if !update {
		log.Println("init k8s yaml have no change")
		return
	}

	// 'by self-release' is used to filter out init webhook later
	err = commitandpush(p.configrepo, fmt.Sprintf("init for project %v:%v", p.Project, p.Branch))
	return
}

// let gen k8s, to decide if it need init again?
// can we make this optional?
//
func (p *Project) genK8s(c genOption) (target string, err error) {
	if p.envMap == nil {
		err = fmt.Errorf("no any env specified, likely can't generate yaml")
		return
	}
	items := []struct {
		src, dst, env string
	}{
		{src: "self-release/template/template-online.yaml", dst: "self-release/k8s-online.yaml", env: ONLINE},
		{src: "self-release/template/template-pre.yaml", dst: "self-release/k8s-pre.yaml", env: PRE},
		{src: "self-release/template/template-test.yaml", dst: "self-release/k8s-test.yaml", env: TEST},
	}
	// should we use p.Init()? using config.yaml to detect?
	needinit := true
	for _, v := range items {
		// if c.singleName != "" && !strings.Contains(v.src, c.singleName) {
		// 	continue
		// }
		src := filepath.Join(p.Project, v.src)
		if p.configrepo.IsExist(src) {
			needinit = false
			break
		}
	}

	if needinit {
		log.Printf("doing initk8s...")
		co := initOption{force: true} // try generate everytime, no need to check force?
		_, e := p.initK8s(p.envMap, co)
		if e != nil {
			err = fmt.Errorf("initK8s err: %v", e)
			return
		}
	}

	var updatedst string
	var update, changed bool
	for _, v := range items {
		if v.env != c.env {
			continue
		}
		src := filepath.Join(p.Project, v.src) // template is in project-path/ template in config repo
		dst := filepath.Join(p.Project, v.dst)
		changed, err = p.CopyToConfigWithVerify(src, dst, p.envMap)
		if err != nil {
			err = fmt.Errorf("copytoconfig err: %v", err)
			return
		}
		target = filepath.Join(p.configrepo.GetWorkDir(), p.Project, v.dst)
		if changed {
			log.Printf("file: %v will be updated", v.dst)
			update = true
		}
		updatedst = dst
	}
	if !update {
		log.Println("generated k8s yaml have no change")
		return
	}
	err = commitandpush(p.configrepo, fmt.Sprintf("generated %v for %v", updatedst, p.Project))
	return
}

func commitandpush(repo *git.Repo, text string) (err error) {
	err = repo.CommitAndPush(text)
	if err != nil {
		err = fmt.Errorf("push change to repo %v:%v,err: %v", repo.Project, repo.Branch, err)
	}
	return
}

// copy content to any repo
// copytoconfig is init? init to two repos?
//
// assume all src come from config-repo
func (p *Project) CopyToConfigWithVerify(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetVerify())
}

func (p *Project) CopyToConfigNoGen(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetNoGen())
}
func (p *Project) CopyToConfigNoGenForce(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetNoGen(), SetForce())
}

func (p *Project) CopyToConfig(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap)
}

func (p *Project) CopyToRepo(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.repo, src, dst, envMap)
}

var ErrNoChange = errors.New("have no change")

type copyto struct {
	verify    bool
	nogen     bool
	force     bool
	finalbody *string
}

type copytOption func(c *copyto)

func SetVerify() copytOption {
	return func(c *copyto) {
		c.verify = true
	}
}
func SetNoGen() copytOption {
	return func(c *copyto) {
		c.nogen = true
	}
}

func SetForce() copytOption {
	return func(c *copyto) {
		c.force = true
	}
}

func SetFinalBody(body *string) copytOption {
	return func(c *copyto) {
		c.finalbody = body
	}
}

func CopyTo(repo, torepo *git.Repo, src, dst string, envMap map[string]string, options ...copytOption) (changed bool, err error) {
	o := &copyto{}
	for _, op := range options {
		op(o)
	}
	c, err := getcontent(repo, src)
	if err != nil {
		return
	}
	var body string
	if !o.nogen {
		// fmt.Println("convert", convertToSubst(c))  // for test
		// return
		body, err = generateByMap(convertToSubst(c), envMap)
		if err != nil {
			err = fmt.Errorf("generate with map err: %v", err)
			return
		}
	} else {
		body = c
	}
	if body == "" {
		err = fmt.Errorf("try to write empty content")
		return
	}

	var exist bool
	exist, changed, err = checkChanged(torepo, dst, body)
	if err != nil {
		err = fmt.Errorf("check if changed err: %v", err)
		return
	}
	if !changed {
		// err = ErrNoChange
		return
	}
	var note string
	if exist {
		if !o.force {
			log.Printf("%v should changed, it's already exist and no force provided, skip", dst)
			return
		} else {
			note = "(overwrite)"
		}
	}
	if o.verify {
		_, err = ValidateByKubectl(body, dst)
		if err != nil {
			log.Printf("validate body: %v\n", body)
			err = fmt.Errorf("validate by kubectl err: %v", err)
			return
		}
	}
	// if o.finalbody != nil {
	// 	target := filepath.Join(repo.GetWorkDir(), dst)
	// 	o.finalbody = &target
	// }
	log.Printf("writing file: %v %v to %v:%v\n", dst, note, torepo.Project, torepo.Branch)
	err = putcontent(torepo, dst, body)
	if err != nil {
		err = fmt.Errorf("putcontent err: %v", err)
		return
	}

	changed = true
	return
}

func getcontent(repo *git.Repo, path string) (content string, err error) {
	// f := filepath.Join("template", v.Template) // prefix template for template
	// f := filepath.Join(project, path)
	b, err := repo.GetFile(path)
	if err != nil {
		err = fmt.Errorf("get file: %v err: %v", path, err)
		return
	}
	content = string(b)
	return
}

// we just overwrite it
func putcontent(repo *git.Repo, path, content string) (err error) {
	// exist := repo.IsExist(path)
	// if exist {
	// 	log.Printf("file: %v exist, will be overwrite", path)
	// }
	err = repo.Add(path, content)
	return
}
func checkChanged(repo *git.Repo, path, content string) (exist, changed bool, err error) {
	oldfinal, err := repo.GetFile(path)
	if err != nil {
		changed = true // take it as no file exist
		err = nil      // we don't need this err
		return
	} else {
		exist = true
	}
	sum1, err := getHash(string(oldfinal))
	if err != nil {
		err = fmt.Errorf("gethash1 err: %v", err)
		return
	}
	sum2, err := getHash(content)
	if err != nil {
		err = fmt.Errorf("gethash2 err: %v", err)
		return
	}
	if sum1 != sum2 {
		changed = true
		return
	}
	return
}

// // if no variable to replace or no custom setting, no need to init repotemplate?
// // we generate for init, so it will easier to custom later
// // gen or add to git?  // why not generate once
// func (p *Project) initRepoTemplateOrFinal(configrepo *git.Repo, force bool, v File, envMap map[string]string) (updateconfigrepo bool, err error) {
// 	if v.RepoTemplate == "" && v.Final == "" {
// 		err = fmt.Errorf("nothing toinit for project: %v, file: %v, skip", p.Project, v.Name)
// 		return
// 	}
// 	// store repotemplate to configrepo if prefixed with config:
// 	var (
// 		repo = p.repo // for repotemplate only?
// 		// updateprojectrepo bool  // we always update project repo for init phase
// 		// updateconfigrepo bool
// 		rtmplfile string
// 		// rtmplconfig       bool // repotemplate flag store to config
// 	)

// 	projectName := p.Project

// 	if v.RepoTemplate != "" {
// 		rtmpl := strings.Split(v.RepoTemplate, ":")
// 		if len(rtmpl) == 1 {
// 			// rrepo = p.repo
// 			// updateprojectrepo = true
// 			rtmplfile = rtmpl[0] // store to project repo
// 		} else if len(rtmpl) == 2 {
// 			repo = configrepo
// 			updateconfigrepo = true
// 			rtmplfile = filepath.Join(projectName, rtmpl[1])
// 			log.Printf("will update config for %v\n", v.Name)
// 			// rtmplconfig = true // will store to config repo
// 		} else {
// 			err = fmt.Errorf("repotemplate value incorrect, should be \"path\" or \"config:path\" for %v", v.Name)
// 			return
// 		}
// 	}

// 	var initfile string
// 	var evaltemplate bool
// 	var exist bool
// 	if v.RepoTemplate != "" {
// 		initfile = rtmplfile // init repotemplate, later generate final
// 	} else {
// 		initfile = v.Final
// 		evaltemplate = true
// 	}

// 	exist = repo.IsExist(initfile)

// 	// force should only for config.yaml and repo, force for redo of  init
// 	// if exist && v.Name == "config.yaml" && !p.InitForce {
// 	// 	log.Printf("init file: %v exist and force have not set, skip", v.Final)
// 	// 	return
// 	// }
// 	if exist && !v.Overwrite && !force {
// 		log.Printf("init file: %v exist and force or overwrite have not set, skip", v.Final)
// 		return
// 	}

// 	// get config template
// 	f := filepath.Join("template", v.Template) // prefix template for template
// 	tfile, e := configrepo.GetFile(f)
// 	if e != nil {
// 		err = fmt.Errorf("get configtemplate file: %v err: %v", f, e)
// 		return
// 	}
// 	var tbody string
// 	var note string
// 	if evaltemplate {
// 		tbody, err = generateByMap(string(tfile), envMap)
// 		if err != nil {
// 			err = fmt.Errorf("get configtemplate file: %v err: %v", f, err)
// 			return
// 		}
// 		note = "(generated)"
// 	} else {
// 		tbody = string(tfile)
// 	}
// 	if updateconfigrepo {
// 		note += "(init in configrepo)"
// 	}

// 	if v.Perm == 0 {
// 		err = repo.Add(initfile, string(tbody))
// 	} else {
// 		err = repo.Add(initfile, string(tbody), git.SetPerm(v.Perm))
// 	}

// 	log.Printf("inited file: %v%v, project: %v\n", initfile, note, p.Project)

// 	return
// }

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
