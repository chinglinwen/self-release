package template

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

// func SetAutoEnv(autoenv map[string]string) func(*genOption) {
// 	return func(o *genOption) {
// 		o.autoenv = autoenv
// 	}
// }

// for init
func (p *Project) GenerateAndPush(options ...func(*genOption)) (err error) {
	err = p.Generate(options...)
	if err != nil {
		return
	}
	return p.Push()
}

// generate config only? generate build files? generate repotemplate
//
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
		// log.Println("got env file setting:", v)
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

	// append build envs
	for k, v := range p.autoenv {
		envMap[k] = v // support overwrite?
	}

	// set envs
	// err = readEnvs(envFiles)
	// if err != nil {
	// 	err = fmt.Errorf("readenvs err: %v", err)
	// }

	errs := make(initErr)

	found := false
	for _, v := range p.Files {
		if c.generateName != "" {
			if c.generateName != v.Name {
				// mostly specify file to generate, so continue
				continue
			}
		}
		found = true

		// check file setting format is valid? say v.template is empty
		if v.Template == "" && v.RepoTemplate == "" {
			err = fmt.Errorf("template and repotemplate file not specified for %v", v.Name)
			errs[v.Name] = err
			continue
		}

		// === generate repo template parts( if not ovewwrite, custom setting will be keeped)
		err = p.genRepoTemplate(v)
		if err != nil {
			err = fmt.Errorf("genRepoTemplate project: %v file: %v err: %v", p.Project, v.RepoTemplate, err)
			errs[v.Name] = err
			continue
		}

		// === generate final parts
		var templateBody []byte
		if v.RepoTemplate != "" {
			// read from repo if specified, which need init first (later human can customize it)
			templateBody, err = p.repo.GetFile(v.RepoTemplate)
			if err != nil {
				err = fmt.Errorf("get repo template file: %v err: %v", v.RepoTemplate, err)
				errs[v.Name] = err
				continue
			}
		} else {
			// read from config repo by default
			f := filepath.Join("template", v.Template) // prefix template for template
			templateBody, err = configrepo.GetFile(f)
			if err != nil {
				err = fmt.Errorf("get configrepo template file: %v err: %v", f, err)
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
		finalbody, err = generateByMap(string(templateBody), envMap)
		if err != nil {
			err = fmt.Errorf("generate finalbody for: %v, err: %v", v.Name, err)
			errs[v.Name] = err
			continue
			// return
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
		log.Printf("generated final file%v: %v/%v\n", new, p.repo.GetWorkDir(), file)
	}
	if c.generateName != "" && !found {
		err = fmt.Errorf("generate finalbody for: %v, err: not found item in config", c.generateName)
	}
	return
}

// if no variable to replace or no custom setting, no need to init repotemplate?
// gen or add to git?  // why not generate once
func (p *Project) genRepoTemplate(v File) (err error) {
	if v.RepoTemplate == "" {
		return // nothing to do
	}

	if p.repo.IsExist(v.RepoTemplate) && !v.Overwrite && !p.Force {
		err = fmt.Errorf("repotemplate file: %v exist and force or overwrite not set, skip", v.Final)
		return
	}

	// get config template
	f := filepath.Join("template", v.Template) // prefix template for template
	tfile, e := configrepo.GetFile(f)
	if e != nil {
		err = fmt.Errorf("get configtemplate file: %v err: %v", f, e)
		return
	}

	if v.Perm == 0 {
		err = p.repo.Add(v.RepoTemplate, string(tfile))
	} else {
		err = p.repo.Add(v.RepoTemplate, string(tfile), git.SetPerm(v.Perm))
	}

	log.Printf("created repotemplate file: %v, project: %v\n", v.Name, p.Project)

	return
}

func (p *Project) Push() (err error) {
	return p.repo.Push()
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
