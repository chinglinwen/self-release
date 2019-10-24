package project

import (
	"errors"
	"fmt"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
)

// a webpage to trigger the test
type initOption struct {
	// singleName string
	force      bool
	dockeronly bool
	configonly bool
	// configVer string
	// branch    string
	// devbranch string // do we need this?
	// buildmode string
	config *ProjectConfig
}

// force is used to re-init config.yaml
func SetInitForce() func(*initOption) {
	return func(p *initOption) {
		p.force = true
	}
}

func SetInitDockerOnly(dockeronly bool) func(*initOption) {
	return func(p *initOption) {
		p.dockeronly = dockeronly
	}
}
func SetInitConfigOnly(configonly bool) func(*initOption) {
	return func(p *initOption) {
		p.configonly = configonly
	}
}

func SetInitConfig(config *ProjectConfig) func(*initOption) {
	return func(p *initOption) {
		p.config = config
	}
}

// init can reading from repo's config, or assume have project name only(using default config version)?
//
// init template file, config.yaml and repotemplate files
// if repo config.yaml exist, it will affect init process?
func (p *Project) Init(options ...func(*initOption)) (err error) {
	// c := initOption{branch: "develop"}
	c := initOption{}
	for _, op := range options {
		op(&c)
	}
	log.Printf("got init option: %#v", c)

	// default to init config only
	c.configonly = true

	// change if config provided, overwrite default
	if c.config != nil {
		if c.config.S.BuildMode != "" {
			p.Config.S.BuildMode = c.config.S.BuildMode
		}
		if c.config.S.ConfigVer != "" {
			p.Config.S.ConfigVer = c.config.S.ConfigVer
		}
		if c.config.S.DevBranch != "" {
			p.Config.S.DevBranch = c.config.S.DevBranch
		}
		// if c.config.S.Enable != "" {
		// 	p.Config.S.Enable = c.config.S.Enable
		// }
		p.Config.S.Enable = c.config.S.Enable
		if c.config.S.Version != "" {
			p.Config.S.Version = c.config.S.Version
		}
	}
	// err = p.initAll(c)
	return
}

// Setting set project config from wechat
// got diff too
func ProjectSetting(project string, c ProjectConfig, user string) (out string, err error) {
	if c.S.BuildMode == "" && c.S.DevBranch == "" && c.S.ConfigVer == "" && c.S.Enable && c.S.Version == "" {
		err = fmt.Errorf("no config item provided,so nothing to set\n%v",
			"expected setting [imagebuild=auto|disabled|on][devbranch=develop|test][configver=php.v1]")
		return
	}

	pc := c
	var exist bool
	p, e := NewProject(project)
	if e == nil {
		exist = true
		pc = p.Config
	}

	var update bool
	if exist {
		out = "changed configs are:\n"

		if c.S.BuildMode != "" {
			log.Printf("project: %v changed buildmode from: %v to: %v\n", p.Project, pc.S.BuildMode, c.S.BuildMode)
			if pc.S.BuildMode == c.S.BuildMode {
				out = fmt.Sprintf("%v  buildmode already set to %v\n", out, c.S.BuildMode)
			} else {
				out = fmt.Sprintf("%v  buildmode from: %v -> %v\n", out, pc.S.BuildMode, c.S.BuildMode)
				update = true
			}
			pc.S.BuildMode = c.S.BuildMode
		}
		if c.S.ConfigVer != "" {
			log.Printf("project: %v changed configver from: %v to: %v\n", p.Project, pc.S.ConfigVer, c.S.ConfigVer)
			if pc.S.ConfigVer == c.S.ConfigVer {
				out = fmt.Sprintf("%v  configver already set to %v\n", out, c.S.ConfigVer)
			} else {
				out = fmt.Sprintf("%v  configver from: %v -> %v\n", out, pc.S.ConfigVer, c.S.ConfigVer)
				update = true
			}
			pc.S.ConfigVer = c.S.ConfigVer
		}
		if c.S.DevBranch != "" {
			log.Printf("project: %v changed devbranch from: %v to: %v\n", p.Project, pc.S.DevBranch, c.S.DevBranch)

			if pc.S.DevBranch == c.S.DevBranch {
				out = fmt.Sprintf("%v  devbranch already set to %v\n", out, c.S.DevBranch)
			} else {
				out = fmt.Sprintf("%v  devbranch from: %v -> %v\n", out, pc.S.DevBranch, c.S.DevBranch)
				update = true
			}
			pc.S.DevBranch = c.S.DevBranch
		}
		if c.S.Enable {
			log.Printf("project: %v changed selfrelease from: %v to: %v\n", p.Project, pc.S.Enable, c.S.Enable)
			if pc.S.Enable == c.S.Enable {
				out = fmt.Sprintf("%v  selfrelease already set to %v\n", out, c.S.Enable)
			} else {
				out = fmt.Sprintf("%v  selfrelease from: %v -> %v\n", out, pc.S.Enable, c.S.Enable)
				update = true
			}
			pc.S.Enable = c.S.Enable
		}
		if c.S.Version != "" {
			log.Printf("project: %v changed version from: %v to: %v\n", p.Project, pc.S.Version, c.S.Version)
			if pc.S.Version == c.S.Version {
				out = fmt.Sprintf("%v  Version already set to %v\n", out, c.S.Version)
			} else {
				out = fmt.Sprintf("%v  version from: %v -> %v\n", out, pc.S.Version, c.S.Version)
				update = true
			}
			pc.S.Version = c.S.Version
		}
	} else {
		// create new
		update = true
	}

	if update {
		log.Printf("project: %v saving config\n", project)
		err = ConfigFileWrite(project, pc, SetConfigUser(user))
	}
	return
}

// separate initall for easier operate, init docker only ( aka project repo only )
// human can easily intercept and fix if there's error ( since it's only about docker image )
//
// we still do init k8s relate, but it's optional, it can skip
// func (p *Project) initAll(c initOption) (err error) {
// 	var needinit bool

// 	// we currently ignore autoenv, only config env is working for init
// 	// _, envMap, err := p.ReadEnvs(nil)
// 	// if err != nil {
// 	// 	err = fmt.Errorf("readenvs err: %v", err)
// 	// }
// 	var changed1, changed2 bool
// 	if !c.configonly {
// 		if !p.DockerInited() || c.force {
// 			changed1, err = p.initDocker()
// 			if err != nil {
// 				err = fmt.Errorf("initDocker err: %v", err)
// 				return
// 			}
// 			needinit = true
// 		}
// 	}
// 	// if !c.dockeronly {
// 	// 	if !p.Inited() || c.force {
// 	// 		changed2, err = p.initK8s(envMap, c.force)
// 	// 		if err != nil {
// 	// 			err = fmt.Errorf("initK8s err: %v", err)
// 	// 			return
// 	// 		}
// 	// 		needinit = true
// 	// 	}
// 	// }
// 	if !needinit {
// 		err = fmt.Errorf("both repo and configrepo inited or no force provided, you can try force init")
// 		return
// 	}
// 	if !changed1 && !changed2 {
// 		err = fmt.Errorf("both repo and configrepo have no file change")
// 		return
// 	}
// 	return
// }

// // init docker just an optional steps
// // require build-docker.sh exist, if using self-release to build image
// func (p *Project) initDocker(envMap map[string]string) (update bool, err error) {
// 	repo, err := p.GetRepo()
// 	if err != nil {
// 		return
// 	}
// 	items := []struct {
// 		src, dst string
// 	}{
// 		{src: "php.ini", dst: "ops/php.ini"},
// 		{src: "nginx.conf", dst: "ops/nginx.conf"},
// 		{src: "Dockerfile", dst: "Dockerfile"},
// 		{src: "build-docker.sh", dst: "build-docker.sh"},
// 		// {src: ".gitlab-ci.yml", dst: ".gitlab-ci.yml"},  // so much other files to generate too
// 	}
// 	var changed bool
// 	for _, v := range items {
// 		src := filepath.Join("template", p.Config.S.ConfigVer, v.src)
// 		changed, err = p.CopyToRepo(repo, src, v.dst, envMap)
// 		if err != nil {
// 			err = fmt.Errorf("copytoconfig err: %v", err)
// 		}
// 		if changed {
// 			log.Printf("file: %v will be updated", v.dst)
// 			update = true
// 		}
// 	}
// 	if !update {
// 		log.Println("docker init have no change")
// 		return
// 	}
// 	err = commitandpush(repo, "init docker files by self-release")
// 	return
// }

// /*
//   - name: k8s-online
//     template: php.v1/k8s/template-online.yaml
//     repotemplate: config:self-release/template/template-online.yaml
//     final: config:self-release/k8s-online.yaml
//     validatefinalyaml: true
//   - name: k8s-pre
//     template: php.v1/k8s/template-pre.yaml
//     repotemplate: config:self-release/template/template-pre.yaml
//     final: config:self-release/k8s-pre.yaml
//     validatefinalyaml: true
//   - name: k8s-test
//     template: php.v1/k8s/template-test.yaml
//     repotemplate: config:self-release/template/template-test.yaml
//     final: config:self-release/k8s-test.yaml
// 	validatefinalyaml: true
// */
// func (p *Project) initK8s(envMap map[string]string, force bool) (update bool, err error) {
// 	items := []struct {
// 		src, dst string
// 	}{
// 		{src: "k8s/template-online.yaml", dst: "self-release/template/template-online.yaml"},
// 		{src: "k8s/template-pre.yaml", dst: "self-release/template/template-pre.yaml"},
// 		{src: "k8s/template-test.yaml", dst: "self-release/template/template-test.yaml"},
// 		{src: "config.env", dst: "self-release/config.env"},
// 		// {src: "config.yaml", dst: "self-release/config.yaml"}, // should we add this? TODO:
// 	}
// 	var changed bool
// 	for _, v := range items {
// 		src := filepath.Join("template", p.Config.ConfigVer, v.src)
// 		dst := filepath.Join(p.Project, v.dst)
// 		if force {
// 			changed, err = p.CopyToConfigNoGenForce(src, dst, envMap)
// 		} else {
// 			changed, err = p.CopyToConfigNoGen(src, dst, envMap)
// 		}
// 		if err != nil {
// 			err = fmt.Errorf("copytoconfig err: %v", err)
// 			return
// 		}
// 		if changed {
// 			log.Printf("file: %v will be updated", v.dst)
// 			update = true
// 		}
// 	}
// 	update, err = p.initConfig(envMap, force)
// 	if err != nil {
// 		err = fmt.Errorf("init config.yaml err: %v", err)
// 		return
// 	}
// 	if !update {
// 		log.Println("init k8s yaml have no change")
// 		return
// 	}

// 	// 'by self-release' is used to filter out init webhook later
// 	err = commitandpush(p.configrepo, fmt.Sprintf("init for project %v:%v", p.Project, p.Branch))
// 	return
// }

// // let gen k8s, to decide if it need init again?
// // can we make this optional?
// //
// func (p *Project) genK8s(c genOption) (target string, err error) {
// 	if p.envMap == nil {
// 		err = fmt.Errorf("no any env specified, likely can't generate yaml")
// 		return
// 	}
// 	items := []struct {
// 		src, dst, env string
// 	}{
// 		{src: "self-release/template/template-online.yaml", dst: "self-release/k8s-online.yaml", env: ONLINE},
// 		{src: "self-release/template/template-pre.yaml", dst: "self-release/k8s-pre.yaml", env: PRE},
// 		{src: "self-release/template/template-test.yaml", dst: "self-release/k8s-test.yaml", env: TEST},
// 	}
// 	// should we use p.Init()? using config.yaml to detect?
// 	needinit := true
// 	for _, v := range items {
// 		// if c.singleName != "" && !strings.Contains(v.src, c.singleName) {
// 		// 	continue
// 		// }
// 		src := filepath.Join(p.Project, v.src)
// 		if p.configrepo.IsExist(src) {
// 			needinit = false
// 			break
// 		}
// 	}

// 	if needinit {
// 		log.Printf("doing initk8s...")
// 		// co := initOption{force: true} // try generate everytime, no need to check force?
// 		_, e := p.initK8s(p.envMap, true)
// 		if e != nil {
// 			err = fmt.Errorf("initK8s err: %v", e)
// 			return
// 		}
// 	}

// 	var updatedst string
// 	var update, changed bool
// 	for _, v := range items {
// 		if v.env != c.env {
// 			continue
// 		}
// 		src := filepath.Join(p.Project, v.src) // template is in project-path/ template in config repo
// 		dst := filepath.Join(p.Project, v.dst)
// 		changed, err = p.CopyToConfigWithVerify(src, dst, p.envMap)
// 		if err != nil {
// 			err = fmt.Errorf("copytoconfig err: %v", err)
// 			return
// 		}
// 		target = filepath.Join(p.configrepo.GetWorkDir(), p.Project, v.dst)
// 		if changed {
// 			log.Printf("file: %v will be updated", v.dst)
// 			update = true
// 		}
// 		updatedst = dst
// 	}
// 	if !update {
// 		log.Println("generated k8s yaml have no change")
// 		return
// 	}
// 	err = commitandpush(p.configrepo, fmt.Sprintf("generated %v for %v", updatedst, p.Project))
// 	return
// }

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
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetVerify(), SetForce())
}

func (p *Project) CopyToConfigNoGen(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetNoGen())
}

func (p *Project) CopyToConfigNoGenForce(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap, SetNoGen(), SetForce())
}

// for config.yaml
func (p *Project) CopyToConfigWithSrcNoGen(srcbody, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, "", dst, envMap, SetNoGen(), SetSrcBody(srcbody))
}

// for config.yaml
func (p *Project) CopyToConfigWithSrcNoGenForce(srcbody, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, "", dst, envMap, SetNoGen(), SetSrcBody(srcbody), SetForce())
}

func (p *Project) CopyToConfig(src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, p.configrepo, src, dst, envMap)
}

func (p *Project) CopyToRepoForce(torepo *git.Repo, src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, torepo, src, dst, envMap, SetForce())
}

func (p *Project) CopyToRepo(torepo *git.Repo, src, dst string, envMap map[string]string) (changed bool, err error) {
	return CopyTo(p.configrepo, torepo, src, dst, envMap)
}

var ErrNoChange = errors.New("have no change")

type copyto struct {
	verify    bool
	nogen     bool
	force     bool
	finalbody *string

	srcbody string
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

func SetSrcBody(body string) copytOption {
	return func(c *copyto) {
		c.srcbody = body
	}
}

func CopyTo(repo, torepo *git.Repo, src, dst string, envMap map[string]string, options ...copytOption) (changed bool, err error) {
	o := &copyto{}
	for _, op := range options {
		op(o)
	}
	var c string
	if o.srcbody == "" {
		c, err = getcontent(repo, src)
		if err != nil {
			return
		}
	} else {
		c = o.srcbody
	}

	var body string
	if !o.nogen {
		body, err = generateByMap(c, envMap)
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
		_, err = ValidateByKubectlWithString(body)
		if err != nil {
			log.Printf("validate body: %v\n", body)
			err = fmt.Errorf("validate by kubectl err: %v", err)
			return
		}
	}
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
