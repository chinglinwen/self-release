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
	"github.com/joho/godotenv"
)

// to match specific k8s yaml
const (
	DEV    = "develop"
	TEST   = "test"
	PRE    = "pre"
	ONLINE = "online"
)

// an api call to test?
// using curl? or webpage

// a webpage to trigger the release, manual release

// a webpage to trigger the test
type genOption struct {
	singleName string
	autoenv    map[string]string
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

// for init
func (p *Project) GenerateAndPush(options ...func(*genOption)) (err error) {
	err = p.Generate(options...)
	if err != nil {
		return
	}
	return p.CommitAndPush(fmt.Sprintf("generate for %v", p.Project))
}

func GetEnvFromBranch(branch string) string {
	env := TEST
	switch branch {
	case PRE:
		env = PRE
	case ONLINE:
		env = ONLINE
	default:
	}
	return env
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

// generate config only? generate build files? not generate repotemplate
//
// generate by env setting
// generate to config-repo only
// let's generate to local first, later if needed ( upload to remote ), say trigger by init?
func (p *Project) Generate(options ...func(*genOption)) (err error) {
	if !p.Inited() {
		err = fmt.Errorf("project %v have not init", p.Project)
		return
	}
	c := &genOption{}
	for _, op := range options {
		op(c)
	}

	// default to test?
	env := GetEnvFromBranch(p.Branch)
	log.Printf("got project %v, env: %v to generate", p.Project, env)

	// set envs with autoenv together
	envMap, err := p.readEnvs(c.autoenv)
	if err != nil {
		err = fmt.Errorf("readenvs err: %v", err)
	}

	var (
		errs              = make(initErr)
		found             = false
		updateprojectrepo bool
		updateconfigrepo  bool
	)

	// we do ignore template for static files?
	for _, v := range p.Files {
		if c.singleName != "" {
			if c.singleName != v.Name { // try support filename match?
				// mostly specify file to generate, so continue
				continue
			}
		}
		found = true

		// repotemplate is generate by init for one time

		// check file setting format is valid? say v.template is empty
		if v.Template == "" && v.RepoTemplate == "" {
			err = fmt.Errorf("template and repotemplate file not specified for %v", v.Name)
			errs[v.Name] = err
			continue
		}

		// skip inited final files
		if v.RepoTemplate == "" {
			continue
		}

		// match k8s yaml to env only
		if !strings.Contains(v.Name, env) {
			log.Printf("skip non-env generate for name: %v, env: %v\n", v.Name, env)
			continue
		}

		// // === generate repo template parts( if not ovewwrite, custom setting will be keeped)
		// err = p.genRepoTemplate(v)
		// if err != nil {
		// 	err = fmt.Errorf("genRepoTemplate project: %v file: %v err: %v", p.Project, v.RepoTemplate, err)
		// 	errs[v.Name] = err
		// 	continue
		// }

		// === generate final parts

		// get repotemplate first, if it exist
		var templateBody string
		if v.RepoTemplate != "" {
			// read from repo if specified, which need init first (later human can customize it)
			tbody, err := p.repo.GetFile(v.RepoTemplate)
			if err != nil {
				log.Printf("get repo template file: %v err: %v, will ignore", v.RepoTemplate, err)
				// err = fmt.Errorf("get repo template file: %v err: %v", v.RepoTemplate, err)
				// errs[v.Name] = err
				// continue // we should hanlde things gracefully at here
			} else {
				templateBody = string(tbody)
			}
		}

		// if empty, or not exist, using default template
		if templateBody == "" {
			// read from config repo by default
			f := filepath.Join("template", v.Template) // prefix template for template
			tobdy, e := configrepo.GetFile(f)
			if e != nil {
				err = fmt.Errorf("get configrepo template file: %v err: %v", f, e)
				errs[v.Name] = err
				continue
			}
			templateBody = string(tobdy)
			if templateBody == "" {
				err = fmt.Errorf("get template file: %v err: it's empty", f)
				errs[v.Name] = err
				continue
			}
		}

		// how to get from template to final

		// https://github.com/drone/envsubst
		// can generate block?
		// use env to overwrite
		var finalbody string
		// finalbody, err = generateByEnv(string(templateBody))
		finalbody, err = generateByMap(convertToSubst(templateBody), envMap)
		if err != nil {
			err = fmt.Errorf("generate finalbody for: %v, err: %v", v.Name, err)
			errs[v.Name] = err
			continue
			// return
		}
		// log.Println(string(templateBody[:30]), finalbody[:30], envMap)

		// write finalbody to project? validate first
		// final body often auto generate, though for k8s, we may still validate first

		// fmt.Println("finalbody:", finalbody)

		var repo *git.Repo
		var file string

		projectName := p.Project

		finals := strings.Split(v.Final, ":")
		if len(finals) == 1 {
			repo = p.repo
			updateprojectrepo = true
			file = finals[0]
		} else if len(finals) == 2 {
			repo = configrepo
			updateconfigrepo = true
			file = filepath.Join(projectName, finals[1])
		} else {
			err = fmt.Errorf("final value incorrect, should be \"path\" or \"config:path\" for %v", v.Name)
			errs[v.Name] = err
			continue
		}

		new := ""
		oldfinal, err := p.repo.GetFile(file)
		if err != nil {
			new = "(new)"
			// log.Printf("gethash1 err: %v, will move on", err)
		}

		sum1, err := getHash(string(oldfinal))
		if err != nil {
			log.Printf("gethash2 err: %v, will move on", err)
		}
		sum2, err := getHash(finalbody)
		if err != nil {
			err = fmt.Errorf("gethash2 err: %v", err)
			errs[v.Name] = err
			continue
		}
		if sum1 == sum2 {
			// got no change
			// log.Printf("ignore file: %v, there's no change for the final", file)
			continue
		}

		if v.Perm == 0 {
			err = repo.Add(file, finalbody)
		} else {
			err = repo.Add(file, finalbody, git.SetPerm(v.Perm))
		}
		log.Printf("generated final file%v: %v/%v\n", new, repo.GetWorkDir(), file)

		// validate before write final? no where to lookout what's wrong?
		if v.ValidateFinalYaml {
			_, e := ValidateByKubectl(finalbody, file)
			if err != nil {
				log.Printf("validate finalbody for: %v, err: %v", file, e)
				// err = fmt.Errorf("validate finalbody for: %v, err: %v", file, e)
				// continue or just logs
			}
			log.Printf("validate finalbody for: %v ok", file)
		}
	}
	if c.singleName != "" && !found {
		err = fmt.Errorf("generate finalbody for: %v, err: not found item in config", c.singleName)
		return
	}
	if updateconfigrepo {
		err = configrepo.CommitAndPush(fmt.Sprintf("generate final files for %v", p.Project))
		if err != nil {
			err = fmt.Errorf("configrepo push err: %v, project: %v", err, p.Project)
			return
		}
		log.Println("configrepo commit and pushed")
	}
	if updateprojectrepo {
		err = p.CommitAndPush(fmt.Sprintf("generate final files for %v", p.Project))
		if err != nil {
			err = fmt.Errorf("repo push err: %v, project: %v", err, p.Project)
			return
		}
		log.Println("projectrepo commit and pushed")
	}
	// log.Println("done generate final files for", p.Project)
	return
}

func (p *Project) CommitAndPush(commitText string) (err error) {
	return p.repo.CommitAndPush(commitText)
}

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
	return envsubst.Eval(templateBody, func(k string) string {
		if v, ok := envMap[k]; !ok {
			log.Printf("got unknown env config name: %v\n", k)
			return fmt.Sprintf("UNKNOWN-%v", k)
		} else {
			return v
		}
		return ""
	})
}

func generateByEnv(templateBody string) (string, error) {
	return envsubst.EvalEnv(templateBody)
}

// https://github.com/joho/godotenv
// func readEnvs(files []string) (err error) {
// 	// default to .env
// 	log.Println("reading envfiles ", files)
// 	return godotenv.Overload(files...)
// }

// autoenv can be nil, if no env settings
func (p *Project) readEnvs(autoenv map[string]string) (envMap map[string]string, err error) {
	// for the filter
	envFiles := []string{}

	// read env
	if len(p.EnvFiles) == 0 {
		defaultEnv := fmt.Sprintf("%v/config.env", defaultRepoConfigPath)
		log.Printf("no env specified, setting default to %v\n", defaultEnv)
		envFiles = append(envFiles, defaultEnv)
	}

	for _, v := range p.EnvFiles {
		// log.Println("got env file setting:", v)
		f := filepath.Join(p.repo.GetWorkDir(), v)
		if isExist(f) {
			envFiles = append(envFiles, f)
		} else {
			log.Printf("env file: %v, setted but not exist, usually for the firsttime init\n", f)
		}
	}

	envMap, err = godotenv.Read(envFiles...) // it seems we just read env first be fore init
	if err != nil {
		err = fmt.Errorf("readenvfiles err: %v", err)
		return // TODO: need ignore?
	}

	if autoenv == nil {
		return
	}

	// append build envs
	for k, v := range autoenv {
		envMap[k] = v // support overwrite?
	}

	return
}

func isExist(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}
