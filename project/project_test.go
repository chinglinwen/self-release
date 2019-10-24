package project

import (
	"testing"
)

// func TestMain(m *testing.M) {
// 	Init()
// 	code := m.Run()
// 	os.Exit(code)
// }

func init() {
	Setting("", "", "wenzhenglin/config-deploy")
}

var pyaml = `
project: wenzhenglin/project-example
env: master
files:
  - name: config.yaml
    template: php.v1/config.yaml
    final: _ops/config.yaml
  - name: dockerfile
    template: php.v1/Dockerfile
    final: Dockerfile
  # - name: config.yaml
  #   template: php.v1/config.yaml
  #   final: _ops/config.yaml
nopull: true
`

// use this as project model http://g.haodai.net/wenzhenglin/project-example

var exampleproject = "wenzhenglin/project-example"

func TestNewProject(t *testing.T) {
	// p, err := NewProject(exampleproject, SetBranch("develop"))
	// if err != nil {
	// 	t.Error("newproject err", err)
	// 	return
	// }
	// pretty(p)

	p, err := NewProject("wenzhenglin/test", SetBranch("develop"))
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	pretty("p", p)
}

// `
// project: wenzhenglin/project-example
// env: master
// #files:
// #  - name: config.yaml
// #    template: php.v1/config.yaml
// #    final: _ops/config.yaml
// #  - name: dockerfile
// #    template: php.v1/Dockerfile
// #    final: Dockerfile
// #    overwrite: true
// #  - name: k8s-online
// #    template: php.v1/k8s/k8s-online.yaml
// #    repoTemplate: _ops/template/k8s-online.yaml
// #    final: _ops/k8s-online.yaml
// #nopull: true
// `

/*
example project json

{
  "Project": "wenzhenglin/project-example",
  "Branch": "develop",
  "DevBranch": "develop",
  "ConfigFile": "",
  "Files": [
    {
      "Name": "config.yaml",
      "Template": "php.v1/config.yaml",
      "Final": "_ops/config.yaml",
      "RepoTemplate": "",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "config.env",
      "Template": "php.v1/config.env",
      "Final": "_ops/config.env",
      "RepoTemplate": "",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "php.ini",
      "Template": "php.v1/php.ini",
      "Final": "_ops/php.ini",
      "RepoTemplate": "",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "nginx.conf",
      "Template": "php.v1/nginx.conf",
      "Final": "_ops/nginx.conf",
      "RepoTemplate": "",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "dockerfile",
      "Template": "php.v1/Dockerfile",
      "Final": "Dockerfile",
      "RepoTemplate": "",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "build-docker.sh",
      "Template": "php.v1/build-docker.sh",
      "Final": "build-docker.sh",
      "RepoTemplate": "",
      "Overwrite": true,
      "Perm": 0,
      "ValidateFinalYaml": false
    },
    {
      "Name": "k8s-online",
      "Template": "php.v1/k8s/k8s-online.yaml",
      "Final": "config:k8s-online.yaml",
      "RepoTemplate": "_ops/template/k8s-online.yaml",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": true
    },
    {
      "Name": "k8s-pre",
      "Template": "php.v1/k8s/k8s-pre.yaml",
      "Final": "config:k8s-pre.yaml",
      "RepoTemplate": "_ops/template/k8s-pre.yaml",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": true
    },
    {
      "Name": "k8s-test",
      "Template": "php.v1/k8s/k8s-test.yaml",
      "Final": "config:k8s-test.yaml",
      "RepoTemplate": "_ops/template/k8s-test.yaml",
      "Overwrite": false,
      "Perm": 0,
      "ValidateFinalYaml": true
    }
  ],
  "EnvFiles": [
    "_ops/config.env"
  ],
  "InitForce": false,
  "NoPull": false,
  "ConfigVer": ""
}
*/
func TestNoPull(t *testing.T) {

	p, err := NewProject(exampleproject) //
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	_ = p
	// spew.Dump(p.Files)
	// // spew.Dump(p)
	// if !p.NoPull {
	// 	t.Errorf("got nopull %v, want %v", p.NoPull, true)
	// 	return
	// }
}

func TestProjectInit(t *testing.T) {

	p, err := NewProject(exampleproject, SetBranch("develop"))
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	// for _, v := range p.Files {
	// 	fmt.Printf("file: %#v\n", v)
	// }
	// err = p.Init()
	// err = p.Init(SetInitForce(), SetInitName("build-docker.sh"))
	err = p.Init(SetInitForce())
	// _, err := ProjectInit(exampleproject)
	if err != nil {
		t.Error("newproject init err", err)
		return
	}
}

// func TestProjectSetting(t *testing.T) {
// 	p, err := NewProject(exampleproject, SetBranch("develop"))
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}

// 	_, err = p.Setting(ProjectConfig{
// 		S: SelfRelease{BuildMode: "auto"},
// 	})
// 	if err != nil {
// 		t.Error("project set config err", err)
// 		return
// 	}
// }

// func TestDecodeConfig(t *testing.T) {
// 	a := `
// # configver choose different version k8s template, etc.
// configver: php.v1
// # branch relate to test-env build frequency
// # devbranch automatic trigger test-env build and apply
// # change the branch to "test", if you want less time build for test-env
// devbranch: develop
// # support [on, auto, disabled]
// # on -> build everytime
// # auto -> build if no image exist in registry
// # disabled -> disable image build, but apply yaml to k8s
// buildmode: auto
// aa: new # new filed will ignore, we can set this to filter for test instance only
// `
// 	c, err := decodeConfig([]byte(a))
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	// spew.Dump("c", c)
// 	fmt.Printf("c: %v", c)
// }
