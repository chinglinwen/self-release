package project

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// provide values.yaml read and write

// frontend is using json, helm charts use yaml

type Values struct {
	Envs  map[string]string `json:"envs,omitempty"`
	Mysql []struct {
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
		MountPath string `json:"mountPath,omitempty"`
	} `json:"nfs,omitempty"`
}

func ValuesJsonToYaml(body string) (values string, err error) {
	v, err := parseValuesJson(body)
	if err != nil {
		return
	}
	d, err := yaml.Marshal(&v)
	if err != nil {
		return
	}
	values = string(d)
	return
}

func ValuesYamlToJson(body string) (values string, err error) {
	v, err := parseValuesYaml(body)
	if err != nil {
		return
	}
	d, err := json.Marshal(&v)
	if err != nil {
		return
	}
	values = string(d)
	return
}

func parseValuesJson(body string) (values Values, err error) {
	err = json.Unmarshal([]byte(body), &values)
	if err != nil {
		return
	}
	return
}

func parseValuesYaml(body string) (values Values, err error) {
	err = yaml.Unmarshal([]byte(body), &values)
	if err != nil {
		return
	}
	return
}

func ValuesFileWrite(project, env string, v Values) (err error) {
	d, err := yaml.Marshal(&v)
	if err != nil {
		err = fmt.Errorf("marshal yaml err: %v", err)
		return
	}
	body := string(d)

	// skip project fetch, we fetch config repo directly
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	// how to handle multi env resource
	if env != "" {
		env += "-" + env
	}
	f := fmt.Sprintf("%v/values%v.yaml", project, env)
	return configrepo.AddAndPush(f, body, fmt.Sprintf("add %v", f))
}
func ValuesFileWriteFromJson(project, env, body string) (err error) {
	v, err := ValuesJsonToYaml(body)
	if err != nil {
		err = fmt.Errorf("convert json to yaml err: %v", err)
		return
	}

	// skip project fetch, we fetch config repo directly
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	// how to handle multi env resource
	if env != "" {
		env += "-" + env
	}
	f := fmt.Sprintf("%v/values%v.yaml", project, env)
	return configrepo.AddAndPush(f, v, fmt.Sprintf("add %v", f))
}

func ValuesFileRead(project, env string) (body string, err error) {
	var fetchdefault bool
	// how to handle multi env resource
	if env != "" {
		env += "-" + env
		f := fmt.Sprintf("%v/values%v.yaml", project, env)
		body, err = valuesFileRead(f)
		if err == nil {
			return
		}
		fetchdefault = true
	}
	if env == "" || fetchdefault {
		f := fmt.Sprintf("%v/values.yaml", project)
		return valuesFileRead(f)
	}
	return
}

func valuesFileRead(filename string) (body string, err error) {
	// skip project fetch, we fetch config repo directly
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	b, err := configrepo.GetFile(filename)
	body, err = ValuesYamlToJson(string(b))
	if err != nil {
		err = fmt.Errorf("convert yaml to json err: %v", err)
		return
	}
	return
}
