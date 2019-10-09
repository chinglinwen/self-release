package project

import (
	"encoding/json"
	"fmt"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
	"gopkg.in/yaml.v2"
)

const (
	configYaml = "config.yaml"
)

// provide values.yaml read and write

// frontend is using json, helm charts use yaml

type ProjectConfigRepo struct {
	project    string
	configrepo *git.Repo
}

type projectConfigOption func(v *ProjectConfigRepo)

func SetConfigRepo(r *git.Repo) projectConfigOption {
	return func(v *ProjectConfigRepo) {
		v.configrepo = r
	}
}

// new will fetch new repo change too, only if it's nil?
func NewProjectConfigRepo(project string, options ...projectConfigOption) (v *ProjectConfigRepo, err error) {
	v = &ProjectConfigRepo{
		project: project,
	}
	for _, op := range options {
		op(v)
	}
	var r *git.Repo
	if v.configrepo == nil {
		r, err = GetConfigRepo()
		if err != nil {
			err = fmt.Errorf("get config repo err: %v", err)
			return
		}
		v.configrepo = r
	}
	if !v.configrepo.IsExist(project) {
		// should we create it? only if it's write?
		// let create for the init?
		err = fmt.Errorf("project does not exist in config-repo")
		return
	}
	return
}

// main read project config
func ReadProjectConfig(project string, options ...projectConfigOption) (config ProjectConfig, err error) {
	r, err := NewProjectConfigRepo(project, options...)
	if err != nil {
		return
	}
	return r.ReadConfig()
}

// main config write
func ConfigFileWrite(project string, config ProjectConfig, options ...projectConfigOption) (err error) {
	log.Println("writted config", config, "for", project)
	return nil
	r, err := NewProjectConfigRepo(project, options...)
	if err != nil {
		return
	}
	return r.ConfigFileWrite(config)
}

func ParseProjectConfigJson(body string) (c ProjectConfig, err error) {
	err = json.Unmarshal([]byte(body), &c)
	if err != nil {
		return
	}
	return
}

func getConfigFileName(project string) (filename string) {
	return fmt.Sprintf("%v/%v", project, configYaml)
}

func (v *ProjectConfigRepo) ReadConfig() (config ProjectConfig, err error) {
	f := getConfigFileName(v.project)
	return v.readfile(f)
}

func (v *ProjectConfigRepo) readfile(filename string) (config ProjectConfig, err error) {
	b, err := v.configrepo.GetFile(filename)
	return ParseProjectConfigYaml(string(b))
}

// support force flag?
func (v *ProjectConfigRepo) ConfigFileWrite(config ProjectConfig) (err error) {
	_, err = v.configFileWrite(config)
	if err != nil {
		return
	}
	commit := fmt.Sprintf("update config.yaml for: %v", v.project)
	log.Println(commit)
	return v.configrepo.CommitAndPush(commit)
}

func (v *ProjectConfigRepo) configFileWrite(config ProjectConfig) (updated bool, err error) {
	// if no values, no create
	if !checkConfigHasValue(config) {
		updated = false
		log.Printf("no need to update config file for %v\n", v.project)
		return
	}
	log.Printf("start to write config filefor %v\n", v.project)
	d, err := yaml.Marshal(&config)
	if err != nil {
		err = fmt.Errorf("marshal yaml err: %v", err)
		return
	}
	body := string(d)
	f := getConfigFileName(v.project)
	err = v.configrepo.Add(f, body)
	if err != nil {
		err = fmt.Errorf("create file err: %v", err)
		return
	}
	updated = true
	return
}

func checkConfigHasValue(v ProjectConfig) bool {
	if len(v.S.DevBranch) != 0 {
		return true
	}
	if len(v.S.BuildMode) != 0 {
		return true
	}
	if len(v.S.ConfigVer) != 0 {
		return true
	}
	if len(v.S.Version) != 0 {
		return true
	}
	return false
}

func ParseProjectConfigYaml(body string) (config ProjectConfig, err error) {
	err = yaml.Unmarshal([]byte(body), &config)
	if err != nil {
		return
	}
	return
}
