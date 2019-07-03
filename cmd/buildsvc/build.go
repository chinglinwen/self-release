package main

import (
	"fmt"
	"net/http"
	buildpkg "wen/self-release/cmd/buildsvc/build"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

// curl localhost:8005/api/build -F project=wenzhenglin/project-example -F env=test

// can we deploy after gen? it's need bonded?
// if we can gen, we can deploy
// with build and deploy flag to trigger it
func buildAPIHandler(c echo.Context) (err error) {

	project := c.FormValue("project")
	branch := c.FormValue("branch") // branch can be tag
	env := c.FormValue("env")       // parse branch to get it?
	if branch == "" {
		branch = "develop" // default to test build
	}

	if project == "" {
		err = fmt.Errorf("project not provided")
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	}

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		return
	}

	log.Printf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)
	out, e := buildpkg.Build(p.WorkDir, project, branch, env)
	// e := p.Build(project, branch, env, out)
	if e != nil {
		err = fmt.Errorf("build err: %v", e)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	log.Printf("docker build outputs: %v", out)
	// scanner := bufio.NewScanner(strings.NewReader(out))
	// // scanner.Split(bufio.ScanLines)
	// for scanner.Scan() {
	// 	log.Println(scanner.Text())
	// }

	return c.JSONPretty(http.StatusOK, EData(200, "build ok", "ok", out), "")
}

func E(code int, msg, status string) map[string]interface{} {
	log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
	}
}

func EData(code int, msg, status string, data string) map[string]interface{} {
	// log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
		"data":    data,
	}
}
