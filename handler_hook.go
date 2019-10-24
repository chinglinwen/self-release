package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/chinglinwen/log"

	"github.com/davecgh/go-spew/spew"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	"github.com/tidwall/gjson"
	gitlab "github.com/xanzy/go-gitlab"
)

// save log to file, for later access?
// https://github.com/google/logger

func hookHandler(c echo.Context) (err error) {
	log.Debug.Println("start hook handler")

	payload, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		err = fmt.Errorf("read body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	a := make(map[string]interface{})
	err = json.Unmarshal(payload, &a)
	if err != nil {
		err = fmt.Errorf("unmarshal body err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	out, err := prettyjson.Marshal(a)
	if err != nil {
		err = fmt.Errorf("marshal a err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}

	projectName := gjson.GetBytes(payload, "project.name").String()
	if projectName == "config-deploy" || projectName == "self-release" {
		log.Println("ignore config-deploy projects")
		c.JSONPretty(http.StatusOK, E(0, "ignore config-deploy", "ok"), " ")
		return
	}

	ns := gjson.GetBytes(payload, "project.namespace").String()

	log.Printf("got gitlab event for project: %v/%v\n", ns, projectName)

	if ns != "wenzhenglin" && ns != "donglintong" && ns != "yuzongwei" && ns != "robot" {
		log.Println("ignore non-test projects")
		c.JSONPretty(http.StatusOK, E(0, "ignore non-test projects", "ok"), " ")
		return
	}

	// eventName, _ := a["event_name"].(string)
	eventName := gjson.GetBytes(payload, "event_name").String()
	data, err := ParseEvent(eventName, payload)
	if err != nil {
		err = fmt.Errorf("parse event err: %v", err)
		log.Println(err)
		c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
		return
	}
	if eventName == "repository_update" {
		log.Println("ignore repository_update event")
		c.JSONPretty(http.StatusOK, E(0, "ignore repository_update event", "ok"), " ")
		return
	}
	log.Printf("out: %s\n", out)
	log.Debug.Println("got event: ", eventName)

	event1, ok := data.(*PushEvent)
	if ok {
		if event1.TotalCommitsCount == 0 {
			log.Println("ignore 0 commits event")
			return c.JSONPretty(http.StatusOK, E(0, "zero commits", "ok"), " ")
		}

		if len(event1.Commits) >= 1 {
			if strings.Contains(event1.Commits[0].Message, "by self-release") {
				log.Println("ignore project init commits event")
				return c.JSONPretty(http.StatusOK, E(0, "project init commits", "ok"), " ")
			}
		}
		log.Println("got push event")
		log.Printf("commits: %v\n", len(event1.Commits))

		err = handlePush(event1)
		if err != nil {
			err = fmt.Errorf("push release err: %v", err)
			log.Println(err)
			c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
			return
		}
	}

	// tag push event need to remove messge=empty or commitcount=0(except include force keyword? )
	event2, ok := data.(*TagPushEvent)
	if ok {
		// if event2.TotalCommitsCount == 0 {
		// 	log.Println("ignore 0 commits event")
		// 	return c.JSONPretty(http.StatusOK, E(0, "zero commits", "ok"), " ")
		// }

		// TODO(wen): need to recheck?
		// if event2.Message == "" && event2.TotalCommitsCount == 0 {
		// 	log.Println("ignore empty message for tag event")
		// 	return c.JSONPretty(http.StatusOK, E(0, "empty message for tag event", "ok"), " ")
		// }

		if strings.Contains(event2.Message, "by self-release") {
			log.Println("ignore project init commits event")
			return c.JSONPretty(http.StatusOK, E(0, "project init commits", "ok"), " ")
		}

		log.Println("got tag push event")
		for _, v := range event2.Commits {
			pp.Print("modified", v.Modified)
		}
		log.Printf("commits: %v\n", len(event2.Commits))

		// if event2.Project.Namespace == "wenzhenglin" || event2.Project.Namespace == "donglintong" {
		err = handleRelease(event2)
		if err != nil {
			err = fmt.Errorf("tag release err: %v", err)
			log.Println(err)
			c.JSONPretty(http.StatusBadRequest, E(0, err.Error(), "failed"), " ")
			return
		}
	}

	return c.JSONPretty(http.StatusOK, E(0, "endok", "ok"), " ")
}

func processPushEvent(e *gitlab.PushEvent) {
	spew.Dump("event", e)
}

func E(code int, msg, status string) map[string]interface{} {
	// log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
	}
}

func EData(code int, msg, status string, data interface{}) map[string]interface{} {
	// log.Println(msg)
	return map[string]interface{}{
		"code":    code,
		"message": msg,
		"status":  status,
		"data":    data,
	}
}
