package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wen/self-release/git"

	"github.com/chinglinwen/log"

	"github.com/drone/envsubst"
)

// to match specific k8s yaml
const (
	DEV    = "develop"
	TEST   = "test"
	PRE    = "pre" // TODO: pre branch is same as master branch?
	ONLINE = "online"
)

// an api call to test?
// using curl? or webpage

// a webpage to trigger the release, manual release

// a webpage to trigger the test
type genOption struct {
	singleName string
	autoenv    map[string]string
	env        string
}

func SetGenerateName(name string) func(*genOption) {
	return func(o *genOption) {
		o.singleName = name
	}
}

func SetGenAutoEnv(autoenv map[string]string) func(*genOption) {
	return func(o *genOption) {
		o.autoenv = autoenv
	}
}
func SetGenEnv(env string) func(*genOption) {
	return func(o *genOption) {
		o.env = env
	}
}

// for init
// func (p *Project) GenerateAndPush(options ...func(*genOption)) (err error) {
// 	_, err = p.Generate(options...)
// 	if err != nil {
// 		return
// 	}
// 	return p.CommitAndPush(fmt.Sprintf("generate for %v", p.Project))
// }

// get env by parse tag?
// default is test env
// func GetEnvFromBranch(branch string) string {
// 	if git.BranchIsOnline(branch) {
// 		return ONLINE
// 	}
// 	if git.BranchIsPre(branch) {
// 		return PRE
// 	}
// 	return TEST
// }

func GetEnvFromBranchOrCommitID(project, branch string) string {
	if git.BranchIsTag(branch) {
		if git.BranchIsOnline(branch) {
			return ONLINE
		}
		if git.BranchIsPre(branch) {
			return PRE
		}
	} else {
		// check if branch is commitid
		o, p, err := git.GetLastTagCommitID(project)
		if err == nil {
			if o == branch {
				return ONLINE
			}
			if p == branch {
				return PRE
			}
		}
	}
	return TEST
}

func GetProjectName(project string) (namespace, name string, err error) {
	if project == "" {
		err = fmt.Errorf("parse project empty err")
		return
	}
	s := strings.Split(project, "/")
	if len(s) != 2 {
		err = fmt.Errorf("parse project err, invalid project: %v(expect namespace/repo-name)", project)
		return
	}
	return s[0], s[1], nil
}

// func (p *Project) Generate(options ...func(*genOption)) (target string, err error) {
// 	// if !p.Inited() {
// 	// 	err = fmt.Errorf("project %v have not init", p.Project)
// 	// 	return
// 	// }

// 	c := genOption{}
// 	for _, op := range options {
// 		op(&c)
// 	}
// 	if c.autoenv == nil {
// 		err = fmt.Errorf("autoenv is empty")
// 		return
// 	}

// 	// // var envMap map[string]string
// 	// envMap, err := p.readEnvs(c.autoenv)
// 	// if err != nil {
// 	// 	// err = fmt.Errorf("readenvs err: %v", err)
// 	// 	log.Printf("readenvs err: %v, will ignore\n", err)
// 	// 	// envMap = make(map[string]string)
// 	// }
// 	p.envMap = c.autoenv // already merged envs
// 	// p.genOption = c

// 	target, err = p.genK8s(c)
// 	return
// }

// // generate config only? generate build files? not generate repotemplate
// //
// // generate by env setting
// // generate to config-repo only
// // let's generate to local first, later if needed ( upload to remote ), say trigger by init?
// func (p *Project) Generate(options ...func(*genOption)) (target string, err error) {
// 	// configrepo, err := GetConfigRepo()
// 	// if err != nil {
// 	// 	err = fmt.Errorf("get configrepo err: %v", err)
// 	// 	return
// 	// }
// 	if !p.Inited() {
// 		err = fmt.Errorf("project %v have not init", p.Project)
// 		return
// 	}
// 	c := &genOption{}
// 	for _, op := range options {
// 		op(c)
// 	}

// 	// default to test?
// 	env := GetEnvFromBranch(p.Branch)
// 	log.Printf("generate for project: %v, env: %v", p.Project, env)

// 	// set envs with autoenv together
// 	envMap, err := p.readEnvs(c.autoenv)
// 	if err != nil {
// 		err = fmt.Errorf("readenvs err: %v", err)
// 	}

// 	var (
// 		errs              = make(errlist)
// 		found             = false
// 		updateprojectrepo bool
// 		updateconfigrepo  bool

// 		generatedFiles string
// 	)

// 	// we do ignore template for static files?
// 	for _, v := range p.Files {
// 		if c.singleName != "" {
// 			if c.singleName != v.Name { // try support filename match?
// 				// mostly specify file to generate, so continue
// 				continue
// 			}
// 		}
// 		found = true

// 		// repotemplate is generate by init for one time

// 		// check file setting format is valid? say v.template is empty
// 		if v.Template == "" && v.RepoTemplate == "" {
// 			err = fmt.Errorf("template and repotemplate file not specified for %v", v.Name)
// 			errs[v.Name] = err
// 			continue
// 		}

// 		// skip inited final files
// 		if v.RepoTemplate == "" {
// 			continue
// 		}

// 		// match k8s yaml to env only
// 		if !strings.Contains(v.Name, env) {
// 			log.Printf("skip non-env generate for name: %v, env: %v\n", v.Name, env)
// 			continue
// 		}

// 		// // === generate repo template parts( if not ovewwrite, custom setting will be keeped)
// 		// err = p.genRepoTemplate(v)
// 		// if err != nil {
// 		// 	err = fmt.Errorf("genRepoTemplate project: %v file: %v err: %v", p.Project, v.RepoTemplate, err)
// 		// 	errs[v.Name] = err
// 		// 	continue
// 		// }

// 		// === generate final parts

// 		// get repotemplate first, if it exist
// 		var templateBody string
// 		if v.RepoTemplate != "" {
// 			// // read from repo if specified, which need init first (later human can customize it)

// 			// tbody, err := p.repo.GetFile(v.RepoTemplate)
// 			// if err != nil {
// 			// 	log.Printf("get repo template file: %v err: %v, will ignore", v.RepoTemplate, err)
// 			// 	// err = fmt.Errorf("get repo template file: %v err: %v", v.RepoTemplate, err)
// 			// 	// errs[v.Name] = err
// 			// 	// continue // we should hanlde things gracefully at here
// 			// } else {
// 			// 	templateBody = string(tbody)
// 			// }

// 			// change to read from config-repo
// 			tbody, err := p.readRepoTemplate(configrepo, v)
// 			if err != nil {
// 				log.Printf("get repotemplate file: %v err: %v, will ignore", v.RepoTemplate, err)
// 				// err = fmt.Errorf("get repo template file: %v err: %v", v.RepoTemplate, err)
// 				// errs[v.Name] = err
// 				// continue // we should hanlde things gracefully at here
// 			} else {
// 				templateBody = string(tbody)
// 			}
// 		}

// 		// if empty, or not exist, using default template
// 		if templateBody == "" {
// 			// read from config repo by default
// 			f := filepath.Join("template", v.Template) // prefix template for template
// 			tobdy, e := configrepo.GetFile(f)
// 			if e != nil {
// 				err = fmt.Errorf("get configrepo template file: %v err: %v", f, e)
// 				errs[v.Name] = err
// 				continue
// 			}
// 			templateBody = string(tobdy)
// 			if templateBody == "" {
// 				err = fmt.Errorf("get template file: %v err: it's empty", f)
// 				errs[v.Name] = err
// 				continue
// 			}
// 		}

// 		// how to get from template to final

// 		// https://github.com/drone/envsubst
// 		// can generate block?
// 		// use env to overwrite
// 		var finalbody string
// 		// finalbody, err = generateByEnv(string(templateBody))
// 		finalbody, err = generateByMap(convertToSubst(templateBody), envMap)
// 		if err != nil {
// 			err = fmt.Errorf("generate finalbody for: %v, err: %v", v.Name, err)
// 			errs[v.Name] = err
// 			continue
// 			// return
// 		}
// 		// log.Println(string(templateBody[:30]), finalbody[:30], envMap)

// 		// write finalbody to project? validate first
// 		// final body often auto generate, though for k8s, we may still validate first

// 		// fmt.Println("finalbody:", finalbody)

// 		var repo *git.Repo
// 		var file string

// 		projectName := p.Project

// 		finals := strings.Split(v.Final, ":")
// 		if len(finals) == 1 {
// 			repo = p.repo
// 			updateprojectrepo = true
// 			file = finals[0]
// 		} else if len(finals) == 2 {
// 			repo = configrepo
// 			updateconfigrepo = true
// 			file = filepath.Join(projectName, finals[1])
// 		} else {
// 			err = fmt.Errorf("final value incorrect, should be \"path\" or \"config:path\" for %v", v.Name)
// 			errs[v.Name] = err
// 			continue
// 		}

// 		new := ""
// 		oldfinal, err := p.repo.GetFile(file)
// 		if err != nil {
// 			new = "(new)"
// 			// log.Printf("gethash1 err: %v, will move on", err)
// 		}

// 		sum1, err := getHash(string(oldfinal))
// 		if err != nil {
// 			log.Printf("gethash2 err: %v, will move on", err)
// 		}
// 		sum2, err := getHash(finalbody)
// 		if err != nil {
// 			err = fmt.Errorf("gethash2 err: %v", err)
// 			errs[v.Name] = err
// 			continue
// 		}
// 		if sum1 == sum2 {
// 			// got no change
// 			// log.Printf("ignore file: %v, there's no change for the final", file)
// 			continue
// 		}

// 		if v.Perm == 0 {
// 			err = repo.Add(file, finalbody)
// 		} else {
// 			err = repo.Add(file, finalbody, git.SetPerm(v.Perm))
// 		}
// 		log.Printf("generated final file%v: %v/%v\n", new, repo.GetWorkDir(), file)

// 		// validate before write final? no where to lookout what's wrong?
// 		if v.ValidateFinalYaml {
// 			_, e := ValidateByKubectl(finalbody, file)
// 			if err != nil {
// 				log.Printf("validate finalbody for: %v, err: %v", file, e)
// 				// err = fmt.Errorf("validate finalbody for: %v, err: %v", file, e)
// 				// continue or just logs
// 			}
// 			log.Printf("validate finalbody for: %v ok", file)
// 		}

// 		// only one target? for now it is
// 		target = filepath.Join(repo.GetWorkDir(), file)

// 		generatedFiles = fmt.Sprintf("%v%v ", generatedFiles, v.Name)

// 	}
// 	if c.singleName != "" && !found {
// 		err = fmt.Errorf("generate finalbody for: %v, err: not found item in config", c.singleName)
// 		return
// 	}
// 	if updateconfigrepo {
// 		err = configrepo.CommitAndPush(fmt.Sprintf("generate final(%v) files for %v", generatedFiles, p.Project))
// 		if err != nil {
// 			err = fmt.Errorf("configrepo push err: %v, project: %v", err, p.Project)
// 			return
// 		}
// 		log.Println("configrepo commit and pushed")
// 	}

// 	// currently, this is no need, since we don't update project after init, we update config deploy
// 	if updateprojectrepo {
// 		err = p.CommitAndPush(fmt.Sprintf("generate final files for %v", p.Project))
// 		if err != nil {
// 			err = fmt.Errorf("repo push err: %v, project: %v", err, p.Project)
// 			return
// 		}
// 		log.Println("projectrepo commit and pushed")
// 	}
// 	// log.Println("done generate final files for", p.Project)
// 	return
// }

// func (p *Project) readRepoTemplate(configrepo *git.Repo, v File) (tbody []byte, err error) {
// 	// store repotemplate to configrepo if prefixed with config:
// 	var (
// 		repo = p.repo // for repotemplate only?
// 		// updateprojectrepo bool  // we always update project repo for init phase
// 		// updateconfigrepo bool
// 		rtmplfile string
// 		// rtmplconfig       bool // repotemplate flag store to config
// 	)
// 	projectName := p.Project
// 	rtmpl := strings.Split(v.RepoTemplate, ":")
// 	if len(rtmpl) == 1 {
// 		// rrepo = p.repo
// 		// updateprojectrepo = true
// 		rtmplfile = rtmpl[0] // store to project repo
// 	} else if len(rtmpl) == 2 {
// 		repo = configrepo
// 		// updateconfigrepo = true
// 		rtmplfile = filepath.Join(projectName, rtmpl[1])
// 		// log.Printf("will read repotemplate from config for %v\n", v.Name)
// 		// rtmplconfig = true // will store to config repo
// 	} else {
// 		err = fmt.Errorf("repotemplate value incorrect, should be \"path\" or \"config:path\" for %v", v.Name)
// 		return
// 	}

// 	return repo.GetFile(rtmplfile)
// }

// func (p *Project) CommitAndPush(commitText string) (err error) {
// 	return p.repo.CommitAndPush(commitText)
// }

func getHashByFile(file string) (sum string, err error) {
	s, err := ioutil.ReadFile(file)
	if err != nil {
		err = fmt.Errorf("gethashbyfile err: %v", err)
		return
	}
	sum, err = getHash(string(s))
	return
}

func getHash(filebody string) (sum string, err error) {
	h := sha256.New()
	// if file

	_, err = h.Write([]byte(filebody))
	if err != nil {
		err = fmt.Errorf("gethash err: %v", err)
		return
	}
	sum = hex.EncodeToString(h.Sum(nil))
	return
}

func convertToSubst(templateBody string) string {
	// s := strings.ReplaceAll(templateBody, "{{", "")
	// s = strings.ReplaceAll(s, "}}", "")

	ciEnvPattern := regexp.MustCompile(`({{\s+\$(\w+)\s+}})`)
	s := ciEnvPattern.ReplaceAll([]byte(templateBody), []byte("${${2}}"))
	return string(s)
}

func generateByMap(templateBody string, envMap map[string]string) (string, error) {
	// for compatibilities
	templateBody = convertToSubst(templateBody)

	// inject resource here
	templateBody = injectResource(templateBody, envMap)

	return envsubst.Eval(templateBody, func(k string) string {
		if v, ok := envMap[k]; !ok {
			log.Printf("got unknown env config name: %v\n", k)
			return fmt.Sprintf("UNKNOWN-%v", k)
		} else {
			return v
		}
	})
}

func generateByEnv(templateBody string) (string, error) {
	return envsubst.EvalEnv(templateBody)
}

func injectResource(templateBody string, envMap map[string]string) string {
	if envMap["ENV-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "env:", fmt.Sprintf("env:\n%v", envMap["ENV-RESOURCE"]), -1)
	}
	if envMap["VOLUME-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "volumes:", fmt.Sprintf("volumes:\n%v", envMap["VOLUME-RESOURCE"]), -1)
	}
	if envMap["MOUNT-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "volumeMounts::", fmt.Sprintf("volumeMounts::\n%v", envMap["MOUNT-RESOURCE"]), -1)
	}

	return templateBody
}

// https://github.com/joho/godotenv
// func readEnvs(files []string) (err error) {
// 	// default to .env
// 	log.Println("reading envfiles ", files)
// 	return godotenv.Overload(files...)
// }

// // autoenv can be nil, if no env settings
// func (p *Project) ReadEnvs(autoenv map[string]string) (mergeNote []string, envMap map[string]string, err error) {
// 	// for the filter
// 	// envFiles := []string{}

// 	// read env
// 	// if p.EnvFiles != nil {
// 	// if len(p.EnvFiles) == 0 {
// 	// defaultEnv := fmt.Sprintf("%v/config.env", p.configConfigPath)
// 	// log.Printf("no env specified, setting default to %v\n", defaultEnv)
// 	// envFiles = append(envFiles, defaultEnv)
// 	// }

// 	// if after init, this should read from config.yaml?
// 	// for _, v := range envFiles {
// 	// log.Println("got env file setting:", v)
// 	f := filepath.Join(p.configrepo.GetWorkDir(), p.configConfigPath, "config.env") // make this configurable?
// 	if !isExist(f) {
// 		// envFiles = append(envFiles, f)
// 		err = fmt.Errorf("env file: %v not exist", f)
// 		// log.Printf("env file: %v not exist, ignore env\n", f)
// 		// envMap = autoenv
// 		return
// 	}
// 	// }
// 	// }

// 	envMap, err = godotenv.Read(f) // it seems we just read env first be fore init
// 	if err != nil {
// 		err = fmt.Errorf("readenvfiles err: %v", err)
// 		return // TODO: need ignore?
// 	}

// 	// mergeNote = "config envs:\n"
// 	for k, v := range envMap {
// 		mergeNote = append(mergeNote, fmt.Sprintf("configenv: %v=%v\n", k, v))
// 	}

// 	// for inject resources
// 	r, _ := p.readResources()
// 	envMap["ENV-RESOURCE"] = r.envs
// 	envMap["VOLUME-RESOURCE"] = r.volumes
// 	envMap["MOUNT-RESOURCE"] = r.mounts

// 	sort.Slice(mergeNote, func(i, j int) bool {
// 		return mergeNote[i] < mergeNote[j] // alphabetical order
// 	})
// 	mergeNote = append(mergeNote, "") // append newline split
// 	// for k, v := range envMap {
// 	// 	if k == "devBranch" {
// 	// 		p.DevBranch = v
// 	// 	}
// 	// 	if k == "buildMode" {
// 	// 		p.BuildMode = v // read this before build is ok, no need to read for every newproject
// 	// 	}
// 	// }

// 	mergeNote1 := make([]string, 0)
// 	split := "//----------//"
// 	if autoenv != nil {
// 		// append build envs
// 		for k, v := range autoenv {
// 			if configv, ok := envMap[k]; ok {
// 				a := fmt.Sprintf("autoenv: %v=%v %v overwrite by configenv: %v to %v\n", k, configv, split, v, configv)
// 				log.Printf(a)
// 				mergeNote1 = append(mergeNote1, a)
// 				continue
// 			} else {
// 				envMap[k] = v // support overwrite?
// 				mergeNote1 = append(mergeNote1, fmt.Sprintf("autoenv: %v=%v\n", k, v))
// 			}
// 		}
// 	}

// 	sort.Slice(mergeNote1, func(i, j int) bool {
// 		return mergeNote1[i] < mergeNote1[j] // alphabetical order
// 	})

// 	mergeNote = append(mergeNote, mergeNote1...)

// 	return
// }

func isExist(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

type resource struct {
	envs    string
	volumes string
	mounts  string
}

func (p *Project) readResources() (r resource, err error) {
	r.envs, err = p.readResource("envs.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	r.volumes, err = p.readResource("volumes.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	r.mounts, err = p.readResource("mounts.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	err = nil
	return
}

func (p *Project) readResource(name string) (s string, err error) {
	f := filepath.Join(p.configConfigPath, "template", name)
	if !p.configrepo.IsExist(f) {
		err = fmt.Errorf("%v is not exist", name)
		return
	}
	b, err := p.configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("try read %v err: %v", name, err)
		return
	}
	s = string(b)
	return
}
