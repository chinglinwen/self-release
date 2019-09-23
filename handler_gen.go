package main

import (
	"fmt"
	"net/http"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

func genYAMLHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	env := c.Param("env")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("gen yaml for project: %v, env: %v\n ", project, env)

	y, err := projectpkg.HelmGenPrint(project, env)
	if err != nil {
		err = fmt.Errorf("gen yaml for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, EData(0, "read values ok", "ok", y), " ")
}
