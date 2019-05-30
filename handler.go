package main

import (
	"fmt"
	"net/http"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"

	"github.com/labstack/echo"
)

func homeHandler(c echo.Context) error {
	//may do redirect later?
	return c.String(http.StatusOK, "home page")
}

// func initPageHandler(c echo.Context) error {
// 	//may do redirect later?
// 	page := `
// 	<!DOCTYPE html>
// 	<html>

// 	<body>

// 		<h2>Init Project</h2>

// 		<form action="/api/init">
// 			First name:<br>
// 			<input type="text" name="project" placeholder="gitlab-namespace/repo-name">
// 			<br> Last name:<br>
// 			<input type="text" name="branch" placeholder="branch or tag">
// 			<br><br>
// 			<input type="checkbox" name="force" value="true"> force init<br><br>
// 			<input type="submit" value="Submit">
// 		</form>

// 	</body>

// 	</html>

// 	`
// 	return c.String(http.StatusOK, page)
// }
// func genPageHandler(c echo.Context) error {
// 	page := `
// <!DOCTYPE html>
// <html>

// <body>

//     <h2>Generate Project</h2>

//     <form action="/api/gen">
//         First name:<br>
//         <input type="text" name="project" placeholder="gitlab-namespace/repo-name">
//         <br> Last name:<br>
//         <input type="text" name="branch" placeholder="branch or tag">
//         <br><br>
//         <input type="submit" value="Submit">
//     </form>

// </body>

// </html>
// `
// 	return c.String(http.StatusOK, page)
// }

func initAPIHandler(c echo.Context) error {
	// records client ip?
	project := c.FormValue("project")
	branch := c.FormValue("branch")
	if branch == "" {
		branch = "develop"
	}
	force := c.FormValue("force")

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))

	if err != nil {
		err = fmt.Errorf("new project: %v, err: %v", project, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	}
	if force == "true" {
		err = p.Init(projectpkg.SetInitForce())
	} else {
		err = p.Init()
	}
	if err != nil {
		err = fmt.Errorf("init api err: %v", err)
		log.Println(err)
		return c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	}

	return c.String(http.StatusOK, "init ok")
}

// can we deploy after gen? it's need bonded?
// if we can gen, we can deploy
// with build and deploy flag to trigger it
func genAPIHandler(c echo.Context) (err error) {
	r := c.Request()
	r.ParseForm()
	booptions := r.Form["booptions"]

	bo := &buildOption{
		gen:    contains(booptions, "gen"),
		build:  contains(booptions, "build"),
		deploy: contains(booptions, "deploy"),
	}

	project := c.FormValue("project")
	branch := c.FormValue("branch")
	env := c.FormValue("env")
	// file := c.FormValue("file")
	if branch == "" {
		branch = "develop"
	}

	username, useremail, msg := getUserInfo(c)

	e := &EventInfo{
		Project:   project,
		Branch:    branch,
		Env:       env, // default derive from branch
		UserName:  username,
		UserEmail: useremail,
		Message:   msg,
	}

	log.Println("event", e)
	log.Println("option", bo)
	return

	err = startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("startBuild for project: %v, branch: %v, err: %v", project, branch, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	// if err != nil {
	// 	err = fmt.Errorf("new project: %v, err: %v", project, err)
	// 	log.Println(err)
	// 	c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// }
	// if file != "" {
	// 	_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv), projectpkg.SetGenerateName(file))
	// } else {
	// 	_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv))
	// }

	// if err != nil {
	// 	err = fmt.Errorf("gen api err: %v", err)
	// 	log.Println(err)
	// 	return c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// }

	return c.String(http.StatusOK, "generate ok")
}

func getUserInfo(c echo.Context) (username, useremail, msg string) {
	username = c.FormValue("username")
	useremail = c.FormValue("useremail")
	msg = c.FormValue("msg")

	if username == "" {
		username = "unknownUser"
	}
	if useremail == "" {
		username = "unknownUserEmail"
	}
	if msg == "" {
		username = "emptyMessage"
	}
	return
}
func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func rollbackAPIHandler(c echo.Context) (err error) {
	// r := c.Request()
	// r.ParseForm()
	// booptions := r.Form["booptions"]

	bo := &buildOption{
		rollback: true,
	}

	project := c.FormValue("project")
	tag := c.FormValue("tag") // optional
	env := c.FormValue("env") // optional
	// file := c.FormValue("file")
	// if branch == "" {
	// 	branch =
	// }

	username, useremail, msg := getUserInfo(c)

	e := &EventInfo{
		Project:   project,
		Branch:    tag,
		Env:       env, // default derive from branch
		UserName:  username,
		UserEmail: useremail,
		Message:   msg,
	}

	log.Println("event", e)
	log.Println("option", bo)
	return

	err = startBuild(e, bo)
	if err != nil {
		err = fmt.Errorf("startBuild for project: %v, branch: %v, err: %v", project, tag, err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	// p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	// if err != nil {
	// 	err = fmt.Errorf("new project: %v, err: %v", project, err)
	// 	log.Println(err)
	// 	c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// }
	// if file != "" {
	// 	_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv), projectpkg.SetGenerateName(file))
	// } else {
	// 	_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv))
	// }

	// if err != nil {
	// 	err = fmt.Errorf("gen api err: %v", err)
	// 	log.Println(err)
	// 	return c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// }

	return c.String(http.StatusOK, "generate ok")
}

// func deployAPIHandler(c echo.Context) error {
// 	project := c.FormValue("project")
// 	branch := c.FormValue("branch")
// 	file := c.FormValue("file")
// 	if branch == "" {
// 		branch = "develop"
// 	}

// 	username := c.FormValue("username")
// 	useremail := c.FormValue("useremail")
// 	msg := c.FormValue("msg")

// 	autoenv := make(map[string]string)
// 	autoenv["PROJECTPATH"] = project
// 	autoenv["BRANCH"] = branch
// 	autoenv["USERNAME"] = username
// 	autoenv["USEREMAIL"] = useremail
// 	autoenv["MSG"] = msg
// 	log.Println("autoenv:", autoenv)

// 	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
// 	if err != nil {
// 		err = fmt.Errorf("new project: %v, err: %v", project, err)
// 		log.Println(err)
// 		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
// 	}
// 	if file != "" {
// 		_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv), projectpkg.SetGenerateName(file))
// 	} else {
// 		_, err = p.Generate(projectpkg.SetGenAutoEnv(autoenv))
// 	}

// 	if err != nil {
// 		err = fmt.Errorf("gen api err: %v", err)
// 		log.Println(err)
// 		return c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
// 	}

// 	return c.String(http.StatusOK, "generate ok")
// }
