package main

import (
	"fmt"
	"net/http"
	"wen/self-release/git"
	"wen/self-release/pkg/resource"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/labstack/echo"
)

// var UserToken = "JQBLUdNq9twWbCbdg6m-"

type Project struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Git   string `json:"git"`
	State bool   `json:"state"`
}

// check if image exist
func projectImageCheckHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	tag := c.FormValue("tag")
	log.Printf("check image tag for project: %v:%v\n ", project, tag)

	exist, err := projectpkg.ImageIsExist(project, tag)
	if err != nil {
		err = fmt.Errorf("check image tag for project: %v:%v, err: %v", project, tag, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	out := map[string]bool{
		"exist": exist,
	}
	return c.JSONPretty(http.StatusOK, EData(0, "check image tag ok", "ok", out), "")
}

// read values file
func projectConfigGetHandler(c echo.Context) (err error) {
	user := c.Request().Header.Get("X-Auth-User")

	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("get values for project: %v, by user %v\n ", project, user)

	out, err := projectpkg.ReadProjectConfig(project)
	if err != nil {
		err = fmt.Errorf("read config file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}

	// out = projectpkg.ProjectConfig{
	// 	devBranch: "test",
	// }
	// out.S.DevBranch = "test"
	return c.JSONPretty(http.StatusOK, EData(0, "read values ok", "ok", out), "")
}

// save values file
func projectConfigUpdateHandler(c echo.Context) (err error) {
	user := c.Request().Header.Get("X-Auth-User")

	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("write values for project: %v, by user %v\n ", project, user)

	r := c.Request()
	body, err := readbody(r)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	// log.Printf("body: %v", body)

	v, err := projectpkg.ParseProjectConfigJson(body)
	if err != nil {
		err = fmt.Errorf("parse body for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}

	err = projectpkg.ConfigFileWrite(project, v, projectpkg.SetConfigUser(user))
	if err != nil {
		err = fmt.Errorf("write config file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, E(0, "saved ok", "ok"), " ")
}

// read values file
func projectValuesGetHandler(c echo.Context) (err error) {
	user := c.Request().Header.Get("X-Auth-User")

	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("get values for project: %v, by user: %v\n ", project, user)

	repo, err := projectpkg.NewValuesRepo(project)
	if err != nil {
		err = fmt.Errorf("read values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}

	out, err := repo.ValuesFileReadAll()
	if err != nil {
		err = fmt.Errorf("read values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, EData(0, "read values ok", "ok", out), "")
}

// save values file
func projectValuesUpdateHandler(c echo.Context) (err error) {
	user := c.Request().Header.Get("X-Auth-User") // set by middlerware

	ns := c.Param("ns")
	project := fmt.Sprintf("%v/%v", ns, c.Param("project"))
	log.Printf("write values for project: %v by user %v\n ", project, user)

	r := c.Request()
	body, err := readbody(r)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	// log.Printf("body: %v", body)

	v, err := projectpkg.ParseAllValuesJson(body)
	if err != nil {
		err = fmt.Errorf("parse body for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(2, err.Error(), "failed"), " ")
		return
	}
	repo, err := projectpkg.NewValuesRepo(project, projectpkg.SetValuesCreate(), projectpkg.SetValuesUser(user))
	if err != nil {
		err = fmt.Errorf("git fetch project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(3, err.Error(), "failed"), " ")
		return
	}
	err = repo.ValuesFileWriteAll(v)
	if err != nil {
		err = fmt.Errorf("write values file for project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(4, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, E(0, "saved ok", "ok"), " ")
}

func pretty(prefix string, a interface{}) {
	out, _ := prettyjson.Marshal(a)
	fmt.Printf("%v: %s\n", prefix, out)
}

func projectUpdateHandler(c echo.Context) (err error) {
	// r, err := c.Request().GetBody()
	// b, _ := ioutil.ReadAll(r)
	// fmt.Println("update body: ", b)
	return c.String(http.StatusOK, `{"result_code":"0","status":"ok"}`)
}

// how to know when to update the cache? manual refresh?
// var projectsCache []*gitlab.Project

func getUserHandler(c echo.Context) (err error) {
	r := c.Request()
	user := r.Header.Get("X-Auth-User")
	// usertoken := r.Header.Get("X-Secret")
	d := map[string]string{
		"user": user,
	}
	return c.JSONPretty(http.StatusOK, EData(0, "get user ok", "ok", d), "")
}
func projectListHandler(c echo.Context) (err error) {
	refresh := c.FormValue("refresh")

	r := c.Request()
	user := r.Header.Get("X-Auth-User")
	usertoken := r.Header.Get("X-Secret")
	// if user == "" || usertoken == "" {
	// 	err := fmt.Errorf("login required")
	// 	log.Println(err)
	// 	return c.JSONPretty(http.StatusOK, E(-1, err.Error(), "failed"), " ")
	// }
	log.Printf("got user: %v, refresh: %v\n", user, refresh)

	// var pss []*gitlab.Project
	pss, err := git.GetProjects(usertoken, refresh)
	if err != nil {
		err = fmt.Errorf("get project err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	var ps []Project
	for _, v := range pss {
		p := Project{
			ID:   v.ID,
			Name: v.PathWithNamespace,
			Git:  v.WebURL,
			// to know all project status takes too long?
			// State: false,  // no easy way to get ?
		}
		ps = append(ps, p)
	}
	log.Printf("got %v projects for user: %v\n", len(ps), user)
	return c.JSONPretty(http.StatusOK, EData(0, "list projects ok", "ok", ps), "")
}
func projectResourceListHandler(c echo.Context) (err error) {
	ns := c.Param("ns")
	if ns == "" {
		return echo.NewHTTPError(http.StatusOK, "empty ns")
	}
	log.Printf("try get resource for %v\n", ns)
	r, err := resource.GetResource(ns)
	if err != nil {
		return echo.NewHTTPError(http.StatusOK, "get resource err:", err)
	}
	log.Printf("get resource for %v ok\n", ns)
	return c.JSONPretty(http.StatusOK, EData(0, "get resource ok", "ok", r), "")
}
