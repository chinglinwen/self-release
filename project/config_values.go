package project

import (
	"encoding/json"
	"fmt"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
	"gopkg.in/yaml.v2"
)

const (
	EnvOnline    = "online"
	EnvPreOnline = "pre"
	EnvTest      = "test"
)

// provide values.yaml read and write

// frontend is using json, helm charts use yaml

type ValuesRepo struct {
	project    string
	configrepo *git.Repo

	create bool
}

type ValuesOption func(*ValuesRepo)

func SetValuesCreate() ValuesOption {
	return func(v *ValuesRepo) {
		v.create = true
	}
}
func NewValuesRepo(project string, options ...ValuesOption) (v *ValuesRepo, err error) {
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	v = &ValuesRepo{
		project:    project,
		configrepo: configrepo,
	}
	for _, op := range options {
		op(v)
	}
	if !v.create && !configrepo.IsExist(project) {
		// should we create it? only if it's write?
		// let create for the init?
		err = fmt.Errorf("project does not exist in config-repo")
		return
	}
	return
}

// twoword key need yaml tag
type ValuesConfig struct {
	NodePort int    `json:"nodePort,omitempty" yaml:"nodePort,omitempty"`
	Domain   string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Deploy   struct {
		Replicas int `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	} `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Monitor struct {
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"monitor,omitempty" yaml:"monitor,omitempty"`
}

type Values struct {
	Config ValuesConfig      `json:"config"`
	Envs   map[string]string `json:"envs,omitempty"`
	Mysql  []struct {
		Name     string `json:"name,omitempty"`
		Host     string `json:"host,omitempty"`
		Port     string `json:"port,omitempty"`
		Database string `json:"database,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"mysql,omitempty"`
	Codis map[string]string `json:"codis,omitempty"`
	Nfs   []struct {
		Name      string `json:"name,omitempty"`
		Path      string `json:"path,omitempty"`
		Server    string `json:"server,omitempty"`
		MountPath string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	} `json:"nfs,omitempty"`
}

type ValuesAll struct {
	Online Values `json:"online,omitempty"`
	Pre    Values `json:"pre,omitempty"`
	Test   Values `json:"test,omitempty"`
}

// we just read all env, no need fallback read online for pre?
func (v *ValuesRepo) ValuesFileReadAll() (all ValuesAll, err error) {
	log.Printf("reading online values for %v\n", v.project)
	onlinev, err := v.ValuesFileRead(ONLINE)
	if err != nil {
		err = fmt.Errorf("get online values err: %v", err)
		return
	}
	log.Printf("reading pre values for %v\n", v.project)
	prev, err := v.ValuesFileRead(PRE)
	if err != nil {
		err = fmt.Errorf("get pre values err: %v", err)
		return
	}
	log.Printf("reading test values for %v\n", v.project)
	testv, err := v.ValuesFileRead(TEST)
	if err != nil {
		err = fmt.Errorf("get test values err: %v", err)
		return
	}
	all = ValuesAll{
		Online: onlinev,
		Pre:    prev,
		Test:   testv,
	}
	return
}

// func (v *ValuesRepo) ValuesFileReadWithFallback(env string) (values Values, err error) {
// 	var fetchdefault bool
// 	// how to handle multi env resource
// 	if env != ONLINE {
// 		values, err = v.ValuesFileRead(env)
// 		if err == nil {
// 			return
// 		}
// 		fetchdefault = true
// 	}
// 	if env == ONLINE || fetchdefault {
// 		return v.ValuesFileRead(ONLINE)
// 	}
// 	return
// }
func (v *ValuesRepo) ValuesFileRead(env string) (values Values, err error) {
	f := getValueFileName(v.project, env)
	return v.readfile(f)
}

func (v *ValuesRepo) readfile(filename string) (values Values, err error) {
	b, err := v.configrepo.GetFile(filename)
	return ParseValuesYaml(string(b))
}

func (v *ValuesRepo) ValuesFileWriteAll(all ValuesAll) (err error) {
	update1, err := v.ValuesFileWrite(ONLINE, all.Online)
	if err != nil {
		return
	}
	update2, err := v.ValuesFileWrite(PRE, all.Pre)
	if err != nil {
		return
	}
	update3, err := v.ValuesFileWrite(TEST, all.Test)
	if err != nil {
		return
	}
	if !update1 && !update2 && !update3 {
		err = fmt.Errorf("all values are empty")
		return
	}
	var a string
	if update1 {
		a += " " + ONLINE
	}
	if update2 {
		a += " " + PRE
	}
	if update3 {
		a += " " + TEST
	}
	commit := fmt.Sprintf("update values.yaml(%v) for: %v", a, v.project)
	log.Println(commit)
	return v.configrepo.CommitAndPush(commit)
}

func checkHasConfig(v ValuesConfig) bool {
	if v.NodePort != 0 {
		return true
	}
	if v.Domain != "" {
		return true
	}
	if v.Deploy.Replicas != 0 {
		return true
	}
	if v.Monitor.Address != "" {
		return true
	}
	return false
}

func checkHasValue(v Values) bool {
	if len(v.Codis) != 0 {
		return true
	}
	if len(v.Envs) != 0 {
		return true
	}
	if len(v.Mysql) != 0 {
		return true
	}
	if len(v.Nfs) != 0 {
		return true
	}
	return checkHasConfig(v.Config)
}
func (v *ValuesRepo) ValuesFileWrite(env string, value Values) (updated bool, err error) {
	// if no values, no create
	if !checkHasValue(value) {
		updated = false
		log.Printf("no need to update values file env: %v for %v\n", env, v.project)
		return
	}
	log.Printf("start to write values file env: %v for %v\n", env, v.project)
	d, err := yaml.Marshal(&value)
	if err != nil {
		err = fmt.Errorf("marshal yaml err: %v", err)
		return
	}
	body := string(d)
	f := getValueFileName(v.project, env)
	err = v.configrepo.Add(f, body)
	if err != nil {
		err = fmt.Errorf("create file err: %v", err)
		return
	}
	updated = true
	return
}

func getValueFileName(project, env string) (filename string) {
	if env == ONLINE {
		return fmt.Sprintf("%v/values.yaml", project)
	}
	return fmt.Sprintf("%v/values-%v.yaml", project, env)

}

// func (v *ValuesRepo) valuesFileWrite(filename, body string) (values Values, err error) {
// 	b, err := v.configrepo.AddAndPush(filename, body, fmt.Sprintf("add values file %v", filename))
// 	return ParseValuesYaml(string(b))
// }

// func ValuesJsonToYaml(body string) (values string, err error) {
// 	v, err := ParseValuesJson(body)
// 	if err != nil {
// 		return
// 	}
// 	d, err := yaml.Marshal(&v)
// 	if err != nil {
// 		return
// 	}
// 	values = string(d)
// 	return
// }

// func ValuesYamlToJson(body string) (values string, err error) {
// 	v, err := ParseValuesYaml(body)
// 	if err != nil {
// 		return
// 	}
// 	d, err := json.Marshal(&v)
// 	if err != nil {
// 		return
// 	}
// 	values = string(d)
// 	return
// }

// for api endpoint
func ParseAllValuesJson(body string) (values ValuesAll, err error) {
	err = json.Unmarshal([]byte(body), &values)
	if err != nil {
		return
	}
	return
}

func ParseValuesJson(body string) (values Values, err error) {
	err = json.Unmarshal([]byte(body), &values)
	if err != nil {
		return
	}
	return
}

func ParseValuesYaml(body string) (values Values, err error) {
	err = yaml.Unmarshal([]byte(body), &values)
	if err != nil {
		return
	}
	return
}

// func ValuesFileWriteFromJson(project, env, body string) (err error) {
// 	v, err := ValuesJsonToYaml(body)
// 	if err != nil {
// 		err = fmt.Errorf("convert json to yaml err: %v", err)
// 		return
// 	}

// 	// skip project fetch, we fetch config repo directly
// 	configrepo, err := GetConfigRepo()
// 	if err != nil {
// 		err = fmt.Errorf("get config repo err: %v", err)
// 		return
// 	}
// 	// how to handle multi env resource
// 	if env != "" {
// 		env += "-" + env
// 	}
// 	f := fmt.Sprintf("%v/values%v.yaml", project, env)
// 	return configrepo.AddAndPush(f, v, fmt.Sprintf("add %v", f))
// }
