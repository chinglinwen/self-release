package template

import (
	"fmt"
	"io/ioutil"
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
  - name: aa
    template: aa
    final: aa1
  - name: bb
    template: bb
    final: bb1

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
files:
  - name: aa
    template: aa
    final: aa1
  - name: bb
    template: bb
    final: bb1
noupdate: true
`

func TestProjectInit(t *testing.T) {

	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	// spew.Dump(p)
	files, err := ioutil.ReadDir(p.repo.GetWorkDir())
	if err != nil {
		t.Error("readdir err", err)
		return
	}
	for _, v := range files {
		fmt.Println(v.Name(), v.Mode(), v.Size())
	}
}
