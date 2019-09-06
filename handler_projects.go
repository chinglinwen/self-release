package main

import (
	"encoding/json"
	"net/http"
	"wen/self-release/git"
	"wen/self-release/pkg/resource"

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