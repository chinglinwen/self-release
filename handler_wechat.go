package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

// what command can support?

// ops can pick project
// dev will remember last project

// ops can set any project?

// test can do extra release? ( they just dev? or extra person , test group? )

// curl localhost:4000 -F from=me -F cmd=help
func wechatHandler(c echo.Context) error {
	r := c.Request()
	ip := r.RemoteAddr
	// fmt.Printf("r: %#v\n", r)
	cmd := r.FormValue("cmd")
	from := r.FormValue("from")
	if cmd == "" {
		cmd = "empty"
	}
	log.Printf("from %v(ip: %v), cmd: %v", from, ip, cmd)

	cmd = strings.TrimPrefix(cmd, "/")

	var out string
	var err error
	switch cmd {
	case "demo":
		out, err = demo(from)
	case "retry":
		out, err = retry(from)
	default:
		err = fmt.Errorf("no cmd to try")
	}

	data, err := encode(out, err)
	if err != nil {
		log.Printf("encode wechat response err: %v, out: %v, err: %v\n", err, out, err)
	}
	// reply, err := NewAsk(from, cmd).Reply()
	// if err != nil {
	// 	fmt.Fprintln(w, "internal error: ", err.Error())
	// 	log.Println("error: ", err.Error())
	// 	return
	// }

	// replyType := gjson.Get(string(reply), "type").String()
	// replyData := gjson.Get(string(reply), "data").String()
	// replyErr := gjson.Get(string(reply), "error").String()
	// var n int
	// if len(replyData) < 10 {
	// 	n = len(replyData)
	// }
	// log.Printf("results type: %v, len: %v, data: %v, err: %v\n",
	// 	replyType, len(replyData), replyData[0:n], replyErr)

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
