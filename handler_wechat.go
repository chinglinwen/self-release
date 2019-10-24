package main

import (
	"encoding/json"
	"net/http"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

// work with commander service and wechat-receiver service.
func wechatHandler(c echo.Context) error {
	r := c.Request()
	ip := r.RemoteAddr

	cmd := r.FormValue("cmd")
	from := r.FormValue("from")
	if cmd == "" {
		cmd = "empty"
	}
	log.Printf("from %v(ip: %v), cmd: %v", from, ip, cmd)

	out, err := doAction(from, cmd)

	data, err := encode(out, err)
	if err != nil {
		log.Printf("encode wechat response err: %v, out: %v, err: %v\n", err, out, err)
	}

	return c.String(http.StatusOK, data)
}

func encode(data string, err error) (string, error) {
	var errtext string
	if err != nil {
		errtext = err.Error()
	}
	b, err := json.MarshalIndent(&struct {
		Type  string `json:"type"`
		Data  string `json:"data"`
		Error string `json:"error"`
	}{
		Type:  "text",
		Data:  data,
		Error: errtext,
	}, "", "  ")
	return string(b), err
}
