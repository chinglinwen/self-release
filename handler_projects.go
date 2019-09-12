package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wen/self-release/git"
	"wen/self-release/pkg/resource"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

var UserToken = "JQBLUdNq9twWbCbdg6m-"

type Project struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Git   string `json:"git"`
	State bool   `json:"state"`
}

func projectValuesGetHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("get values for project: %v\n ", project)

	repo, err := projectpkg.NewValuesRepo(project)
	if err != nil {
		err = fmt.Errorf("read values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	out, err := repo.ValuesFileReadAll()
	if err != nil {
		err = fmt.Errorf("read values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, EData(200, "read values ok", "ok", out), "")
}

func projectValuesUpdateHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("write values for project: %v\n ", project)

	v := projectpkg.ValuesAll{}
	if err = c.Bind(v); err != nil {
		err = fmt.Errorf("read body for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	repo, err := projectpkg.NewValuesRepo(project)
	if err != nil {
		err = fmt.Errorf("write values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	err = repo.ValuesFileWriteAll(v)
	if err != nil {
		err = fmt.Errorf("write values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, E(200, "saved ok", "ok"), " ")
}

func projectUpdateHandler(c echo.Context) (err error) {
	// r, err := c.Request().GetBody()
	// b, _ := ioutil.ReadAll(r)
	// fmt.Println("update body: ", b)
	return c.String(http.StatusOK, `{"result_code":"0","status":"ok"}`)
}
func projectListHandler(c echo.Context) (err error) {
	_, pss, err := git.GetProjects(UserToken)
	if err != nil {
		return
	}
	var ps []Project
	for _, v := range pss {
		p := Project{
			ID:    v.ID,
			Name:  v.PathWithNamespace,
			Git:   v.WebURL,
			State: false,
		}
		ps = append(ps, p)
	}
	b, err := json.Marshal(ps)
	if err != nil {
		return
	}
	return c.String(http.StatusOK, string(b))
}
func projectResourceListHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	if ns == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty ns")
	}
	log.Printf("try get resource for %v\n", ns)
	r, err := resource.GetResource(ns)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "get resource err:", err)
	}
	log.Printf("get resource for %v ok\n", ns)
	// b, _ := json.MarshalIndent(r, "", "  ")
	return c.JSON(http.StatusOK, r)
}
