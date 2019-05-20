package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"wen/self-release/git"

	"github.com/chinglinwen/log"

	"github.com/drone/envsubst"
	"github.com/joho/godotenv"
)

// an api call to test?
// using curl? or webpage

// a webpage to trigger the release, manual release

// a webpage to trigger the test
type genOption struct {
	generateName string
}

func SetGenerateName(name string) func(*genOption) {
	return func(o *genOption) {
		o.generateName = name
	}
}

// for init
func (p *Project) GenerateAndPush(options ...func(*genOption)) (err error) {
	err = p.Generate(options...)
	if err != nil {
		return
	}
	return p.Push()
}

// generate by env setting
// generate to develop branch, not master
// let's generate to local first, later if needed ( upload to remote ), say trigger by init?
func (p *Project) Generate(options ...func(*genOption)) (err error) {
	c := &genOption{}
	for _, op := range options {
		op(c)
	}

	// checkout to branch first, how do they release?

	// clone specific branch?

	// for the filter
	envFiles := []string{}

	// read env
	if len(p.EnvFiles) == 0 {
		log.Printf("no env specified, setting default to %v/config.env", repoConfigPath)
		envFiles = append(envFiles, fmt.Sprintf("%v/config.env", repoConfigPath))
	}

	for _, v := range p.EnvFiles {
		log.Println("got env file setting:", v)
		f := filepath.Join(p.repo.GetWorkDir(), v)
		if isExist(f) {
			envFiles = append(envFiles, f)
		} else {
			log.Printf("env file: %v, setted but not exist\n", f)
		}
	}

	envMap, err := godotenv.Read(envFiles...)
	if err != nil {
		err = fmt.Errorf("readenvfiles err: %v", err)
	}

	// append build envs, but when?
	// for k, v := range p.envMap {

	// }

	// set envs
	// err = readEnvs(envFiles)
	// if err != nil {
	// 	err = fmt.Errorf("readenvs err: %v", err)
	// }

	found := false
	for _, v := range p.Files {
		if c.generateName != "" {
			if c.generateName != v.Name {
				continue
			}
		}
		found = true

		var templateBody []byte
		if v.RepoTemplate != "" {
			// read from repo if specified, which need init first (later human can customize it)
			templateBody, err = p.repo.GetFile(v.RepoTemplate)
			if err != nil {
				err = fmt.Errorf("get template file: %v err: %v", v.RepoTemplate, err)
				continue
			}
		} else {
			// read from config repo by default
			f := filepath.Join("template", v.Template) // prefix template for template
			templateBody, err = configrepo.GetFile(f)
			if err != nil {
				err = fmt.Errorf("get template file: %v err: %v", f, err)
				continue
			}
		}

		// how to get from template to final

		// https://github.com/drone/envsubst
		// can generate block?
		// use env to overwrite
		var finalbody string
		// finalbody, err = generateByEnv(string(templateBody))
		finalbody, err = generateByMap(string(templateBody), envMap)
		if err != nil {
			err = fmt.Errorf("generate finalbody for: %v, err: %v", v.Name, err)
			// continue
			return
		}
		// write finalbody to project? validate first

		// fmt.Println("finalbody:", finalbody)

		var repo *git.Repo
		var file string

		projectName := p.Project

		finals := strings.Split(v.Final, ":")
		if len(finals) == 1 {
			repo = p.repo
			file = finals[0]
		} else if len(finals) == 2 {
			repo = configrepo
			file = filepath.Join(projectName, finals[1])
		} else {
			err = fmt.Errorf("final value incorrect, should be \"path\" or \"config:path\" for %v", v.Name)
			return
		}

		if v.Perm == 0 {
			err = repo.Add(file, finalbody)
		} else {
			err = repo.Add(file, finalbody, git.SetPerm(v.Perm))
		}
		log.Printf("generated final file: %v/%v\n", p.repo.GetWorkDir(), file)
	}
	if c.generateName != "" && !found {
		err = fmt.Errorf("generate finalbody for: %v, err: not found item in config", c.generateName)
	}
	return
}

func (p *Project) Push() (err error) {
	return p.repo.Push()
}

func generateByMap(templateBody string, envMap map[string]string) (string, error) {
	return envsubst.Eval(templateBody, func(k string) string {
		return envMap[k]
	})
}

func generateByEnv(templateBody string) (string, error) {
	return envsubst.EvalEnv(templateBody)
}

// https://github.com/joho/godotenv
func readEnvs(files []string) (err error) {
	// default to .env
	log.Println("reading envfiles ", files)
	return godotenv.Overload(files...)
}

func isExist(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}
