package project

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	dir, project, tag, env := "/home/wen/t/repos/wenzhenglin/project-example", "wenzhenglin/project-example", "develop", "test"

	// out := make(chan string, 10)

	// b.logf("start building for project: %v, branch: %v, env: %v\n", project, branch, env)
	// out, e := p.Build(project, branch, env)
	out, e := Build(dir, project, tag, env)
	if e != nil {
		t.Errorf("build err: %v", e)
		return
	}
	// for v := range out {
	fmt.Println("output:", out)
	// }
}

func TestGetDefaultBuildScript(t *testing.T) {
	project, commitid, env := "wenzhenglin/project-example", "af0dcab65", "test"
	image := GetImage(project, commitid)
	out := getDefaultBuildScript(image, env)
	fmt.Println("output:", out)
}

// go test -timeout 60s wen/self-release/project -run TestBuild -v -count=1
// func TestBuild2(t *testing.T) {
// 	dir, project, tag, env := "/home/wen/t/repos/wenzhenglin/project-example", "wenzhenglin/project-example", "develop", "test"

// 	out := make(chan string, 10)

// 	// b.logf("start building for project: %v, branch: %v, env: %v\n", project, branch, env)
// 	// out, e := p.Build(project, branch, env)
// 	e := Build2(dir, project, tag, env, out)
// 	if e != nil {
// 		t.Errorf("build err: %v", e)
// 		return
// 	}
// 	for v := range out {
// 		fmt.Println("output:", v)
// 	}
// }
