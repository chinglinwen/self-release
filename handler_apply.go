package main

import (
	"fmt"
	"net/http"
	"wen/self-release/pkg/sse"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

const (
	applyOp = iota
	deleteOp
)

func applyYAMLHandler(c echo.Context) (err error) {
	// return c.JSONPretty(http.StatusOK, EData(0, "apply ok", "ok", nil), " ")
	return applyOrDelete(c, applyOp)
}
func deleteYAMLHandler(c echo.Context) (err error) {
	// return c.JSONPretty(http.StatusOK, EData(0, "delete ok", "ok", nil), " ")
	return applyOrDelete(c, deleteOp)
}

func applyOrDelete(c echo.Context, op int) (err error) {
	ns := c.Param("ns")
	// env := c.Param("env")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	env := c.Param("env")

	log.Printf("do apply for project: %v, env: %v\n ", project, env)

	projectenv := fmt.Sprintf("%v-%v", project, env)

	// getinfo
	body, err := readbody(c.Request())
	if err != nil {
		err = fmt.Errorf("read body for %v err: %v", projectenv, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}

	// parse info
	info, err := sse.ParseEventInfoJson(body)
	if err != nil {
		err = fmt.Errorf("parse apply body for %v err: %v", projectenv, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}

	pretty("got project: ", info)

	envmap, err := EventInfoToMap(info)
	if err != nil {
		err = fmt.Errorf("parse event for %v err: %v", projectenv, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(3, err.Error(), "failed"), " ")
		return
	}
	pretty("envinfo", envmap)

	var out string
	if op == applyOp {
		// do apply
		out, err = projectpkg.Apply(project, env, envmap)
		if err != nil {
			err = fmt.Errorf("apply for project: %v, err: %v", projectenv, err)
			log.Println(err)
			c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
			return
		}
		log.Printf("apply project: %v-%v ok\n", project, env)
		return c.JSONPretty(http.StatusOK, EData(0, "apply ok", "ok", out), " ")
	}

	// do delete
	out, err = projectpkg.Delete(project, env, envmap)
	if err != nil {
		err = fmt.Errorf("delete for project: %v, err: %v", projectenv, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
		return
	}
	log.Printf("delete project: %v, env: %v ok\n", projectenv, env)
	return c.JSONPretty(http.StatusOK, EData(0, "delete ok", "ok", out), " ")
}
