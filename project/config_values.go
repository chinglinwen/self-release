package project

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// provide values.yaml read and write

// frontend is using json, helm charts use yaml

type Values struct {
	Envs  map[string]string `json:"envs"`
	Mysql []struct {
		Name     string `json:"name"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"mysql"`
	Codis map[string]string `json:"codis"`
	Nfs   []struct {
		Name      string `json:"name"`
		Path      string `json:"path"`
		Server    string `json:"server"`
		MountPath string `json:"mountPath"`
	} `json:"nfs"`
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

func ValuesFileWrite(project, env, body string) (err error) {
	// skip project fetch, we fetch config repo directly
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}

	v, err := ValuesJsonToYaml(body)
	if err != nil {
		err = fmt.Errorf("convert json to yaml err: %v", err)
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
