package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"wen/self-release/cmd/nfssvc/nfs"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

type exports struct {
	server string
	path   string
	body   string
}

func newNfs(server, path string) *exports {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("read exports file err ", err)
	}
	return &exports{
		server: server,
		path:   path,
		body:   string(b),
	}
}

func (e *exports) listNfsHandler(c echo.Context) (err error) {
	results, err := nfs.Parse(e.body, e.server)
	if err != nil {
		err = fmt.Errorf("parse exports error: ", err)
		log.Println(err)
		c.JSONPretty(http.StatusOK, E(1, err.Error(), "failed"), " ")
		return
	}
	return c.JSONPretty(http.StatusOK, EData(0, "list nfs ok", "ok", results), "")
}

func E(code int, msg, status string) map[string]interface{} {
	log.Println(msg)
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
