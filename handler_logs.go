package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"wen/self-release/pkg/sse"

	"github.com/labstack/echo"
)

func logsHandler(c echo.Context) (err error) {

	project := c.FormValue("project")
	branch := c.FormValue("branch") // branch includes tag
	key := c.FormValue("key")

	var note string

	if project != "" && branch == "" {
		branch = defaultDevBranch
		note = "default"
	}

	var list bool
	var stored bool
	var existmsg string

	type Item struct {
		Key     string
		Project string
		Branch  string
	}
	items := []Item{}

	// if project and key both not specified, list all keys
	if project == "" && key == "" {
		log.Println("getting logs list")
		brokers, e := sse.GetBrokers()
		if err != nil {
			err = fmt.Errorf("GetBrokersFromDisk err: %v", e)
			log.Println(err)
			c.JSONPretty(http.StatusInternalServerError, E(0, err.Error(), "failed"), " ")
			return
		}
		list = true
		for _, v := range brokers {
			if v.Key == "" || v.Project == "" {
				continue
			}
			item := Item{Key: v.Key, Project: v.Project, Branch: v.Branch}
			items = append(items, item)
		}
	}

	if key != "" {
		b, e := sse.GetBrokerFromKey(key)
		if err != nil {
			err = fmt.Errorf("GetBrokerFromKey err: %v", e)
			log.Println(err)
			c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		}
		if b != nil {
			project = b.Project
			branch = b.Branch
			if b.Stored {
				stored = true
				existmsg = b.GetExistMsg()
			}
		}
	}

	p := struct {
		Key      string
		Project  string
		Branch   string
		Note     string
		List     bool
		Stored   bool
		ExistMsg string

		Items []Item
	}{
		Key:     key,
		Project: project,
		Branch:  branch,
		Note:    note,

		List:     list,
		Stored:   stored,
		ExistMsg: existmsg,

		Items: items,
	}

	// Read in the template with our SSE JavaScript code.
	t, err := template.New("logs").Parse(box.MustString("logs.html"))
	// t, err := template.ParseFiles("web/logs.html")
	if err != nil {
		log.Fatal("WTF dude, error parsing your template.")
	}
	// log.Println("parsed template")

	// Render the template, writing to `w`.
	t.Execute(c.Response(), p)

	return
}
