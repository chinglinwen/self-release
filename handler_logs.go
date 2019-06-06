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

	// if project == "" {
	// 	// if no project
	// 	err = fmt.Errorf("project parameter value is empty")
	// 	log.Println(err)
	// 	c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// 	return
	// }

	if project != "" && branch == "" {
		branch = "develop"
		note = "default"
		// err = fmt.Errorf("branch parameter value is empty")
		// log.Println(err)
		// c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		// return
	}

	// var brokers = make(map[string]*sse.Broker{})
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
		// list existing build logs
		// brokers := sse.GetBrokers()
		brokers, e := sse.GetBrokers()
		if err != nil {
			err = fmt.Errorf("GetBrokersFromDisk err: %v", e)
			log.Println(err)
			c.JSONPretty(http.StatusInternalServerError, E(0, err.Error(), "failed"), " ")
			return
		}
		list = true
		// spew.Dump("brokers", brokers)
		for _, v := range brokers {
			if v.Key == "" || v.Project == "" {
				continue
			}
			item := Item{Key: v.Key, Project: v.Project, Branch: v.Branch}
			items = append(items, item)
		}
		// items = append(items, Item{Project: "test", Branch: "dev"})
		// spew.Dump("items", items)
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
		// Projects []*sse.Broker
		Items []Item
	}{
		Key:     key,
		Project: project,
		Branch:  branch,
		Note:    note,

		List:     list,
		Stored:   stored,
		ExistMsg: existmsg,
		// Projects: brokers,
		Items: items,
	}

	// Read in the template with our SSE JavaScript code.
	t, err := template.ParseFiles("web/logs.html")
	if err != nil {
		log.Fatal("WTF dude, error parsing your template.")
	}
	// log.Println("parsed template")

	// Render the template, writing to `w`.
	t.Execute(c.Response(), p)

	// Done.
	// log.Println("finished HTTP request for", project)
	return
}
