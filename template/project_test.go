package template

import (
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestMain(m *testing.M) {
	Init()
	code := m.Run()
	os.Exit(code)
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

func TestNewProject(t *testing.T) {
	p, err := NewProject(pyaml)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	spew.Dump(p)
}

// use this as project model http://g.haodai.net/wenzhenglin/project-example

var exampleproject = `
project: wenzhenglin/project-example
env: master
#files:
#  - name: config.yaml
#    template: php.v1/config.yaml
#    final: _ops/config.yaml
#  - name: dockerfile
#    template: php.v1/Dockerfile
#    final: Dockerfile
#    overwrite: true
#  - name: k8s-online
#    template: php.v1/k8s/k8s-online.yaml
#    repoTemplate: _ops/template/k8s-online.yaml
#    final: _ops/k8s-online.yaml
#nopull: true
`

func TestNoPull(t *testing.T) {

	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	spew.Dump(p.Files)
	// spew.Dump(p)
	if !p.NoPull {
		t.Errorf("got nopull %v, want %v", p.NoPull, true)
		return
	}
}

func TestProjectInit(t *testing.T) {

	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	for _, v := range p.Files {
		fmt.Printf("file: %#v\n", v)
	}
	err = p.Init(SetInitForce())
	if err != nil {
		t.Error("newproject init err", err)
		return
	}
}
