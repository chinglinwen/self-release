package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/k0kubun/pp"

	"github.com/davecgh/go-spew/spew"
	"github.com/tidwall/gjson"

	"github.com/chinglinwen/log"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/labstack/echo"
)

func homeHandler(c echo.Context) error {
	//may do redirect later?
	return c.String(http.StatusOK, "home page")
}

func hookHandler(c echo.Context) (err error) {
	// spew.Dump("c.header", c.Request().Header)
	// header: X-Gitlab-Event: "System Hook"
	payload, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	log.Println("readbody ok")

	a := make(map[string]interface{})
	err = json.Unmarshal(payload, &a)
	if err != nil {
		err = fmt.Errorf("unmarshal body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	log.Println("unmarshal ok")

	out, err := prettyjson.Marshal(a)
	if err != nil {
		err = fmt.Errorf("marshal a err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	log.Println("marshal ok")
	fmt.Printf("out: %s\n", out)

	// log.Printf("===event_name: %v\n", a["event_name"])
	// log.Printf("===message: %v\n", a["message"])

	// eventName, _ := a["event_name"].(string)
	eventName := gjson.GetBytes(payload, "event_name").String()
	data, err := ParseEvent(eventName, payload)
	if err != nil {
		err = fmt.Errorf("parse event err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	// spew.Dump("event:", data)
	// pp.Print("data", data)

	event1, ok := data.(*PushEvent)
	if ok {
		if event1.TotalCommitsCount == 0 {
			log.Println("ignore 0 commits event")
			return c.JSONPretty(http.StatusOK, E(0, "zero commits", "ok"), " ")
		}

		fmt.Println("got push event")
		for _, v := range event1.Commits {
			pp.Print("modified", v.Modified)
		}
		fmt.Printf("commits: %v\n", len(event1.Commits))
		// spew.Dump("details:", event1.Commits)

		// PathWithNamespace is better, name or namespace maybe chinese chars
		if event1.Project.Name == "test" || event1.Project.Name == "project-example" {
			err = handlePush(event1)
			if err != nil {
				err = fmt.Errorf("handle push event err: %v", err)
				log.Println(err)
				c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
				return
			}
		} else {
			log.Println("ignore non-test projects")
		}
	}

	// tag push event need to remove messge=empty or commitcount=0(except include force keyword? )
	event2, ok := data.(*TagPushEvent)
	if ok {
		if event2.TotalCommitsCount == 0 {
			log.Println("ignore 0 commits event")
			return c.JSONPretty(http.StatusOK, E(0, "zero commits", "ok"), " ")
		}
		fmt.Println("got tag push event")
		for _, v := range event2.Commits {
			pp.Print("modified", v.Modified)
		}
		fmt.Printf("commits: %v\n", len(event2.Commits))
		// spew.Dump("details:", event2.Commits)

		if event2.Project.Name == "test" || event2.Project.Name == "project-example" {
			err = handleRelease(event2)
			if err != nil {
				err = fmt.Errorf("handle release event err: %v", err)
				log.Println(err)
				c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
				return
			}

		} else {
			log.Println("ignore non-test projects")
		}
	}

	// var eventType string
	// eventName := a["event_name"]
	// switch eventName {
	// case "push":
	// 	eventType = "Push Hook"
	// case "tag_push":
	// 	eventType = "Tag Push Hook"
	// default:
	// 	t, _ := a["event_name"].(string)
	// 	eventType = t
	// }
	// unexpected event type: System Hook
	// event, err := gitlab.ParseWebhook(gitlab.WebhookEventType(c.Request()), payload)

	// event, err := gitlab.ParseWebhook(gitlab.EventType(eventType), payload)
	// if err != nil {
	// 	err = fmt.Errorf("parse event err: %v", err)
	// 	log.Println(err)
	// 	c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
	// 	return
	// }
	// switch event := event.(type) {
	// case *gitlab.PushEvent:
	// 	processPushEvent(event)
	// // case *gitlab.MergeEvent:
	// // 	processMergeEvent(event)
	// default:
	// 	// msg := fmt.Sprintf("ignore event: %v\n", event)
	// 	return c.JSONPretty(http.StatusOK, E(0, "ignore event", "ok"), " ")
	// }

	return c.JSONPretty(http.StatusOK, E(0, "endok", "ok"), " ")
}
func processPushEvent(e *gitlab.PushEvent) {
	spew.Dump("event", e)
}
func E(code int, msg, status string) map[string]interface{} {
	log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
	}
}

func EData(code int, msg, status string, data []map[string]interface{}) map[string]interface{} {
	log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
		"data":    data,
	}
}

func notifyHandler(c echo.Context) error { return nil }
