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
	log.Printf("do apply for project: %v\n ", project)

	// getinfo
	body, err := readbody(c.Request())
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}

	// parse info
	info, err := sse.ParseEventInfoJson(body)
	if err != nil {
		err = fmt.Errorf("parse apply body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}

	envmap, err := EventInfoToMap(info)
	if err != nil {
		err = fmt.Errorf("convert info to map err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(3, err.Error(), "failed"), " ")
		return
	}
	log.Printf("got apply: %v, env: %v\n", project, env)
	log.Printf("envinfo: \n")
	for k, v := range envmap {
		log.Printf("%v: %v\n", k, v)
	}
	// return c.JSONPretty(http.StatusOK, EData(0, "apply ok", "ok", nil), " ")

	var out string
	if op == applyOp {
		// do apply
		out, err = projectpkg.Apply(project, env, envmap)
		if err != nil {
			err = fmt.Errorf("apply for project: %v, err: %v", project, err)
			log.Println(err)
			c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
			return
		}
		return c.JSONPretty(http.StatusOK, EData(0, "apply ok", "ok", out), " ")
	}

	// do delete
	out, err = projectpkg.Delete(project, env, envmap)
	if err != nil {
		err = fmt.Errorf("delete for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, EData(0, "delete ok", "ok", out), " ")
}
