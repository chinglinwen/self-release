package notify

import (
	"flag"
	"fmt"
	"log"

	resty "gopkg.in/resty.v1"
)

var (
	wechatURL = flag.String("wechat-receiver-url", "http://localhost:8002", "wechat-receiver-url")
)

func Send(name, content string) (reply string, err error) {
	if name == "" {
		err = fmt.Errorf("empty name, skip send")
		return
	}
	if content == "" {
		err = fmt.Errorf("empty content, skip send")
		return
	}
	resp, e := resty. //SetDebug(true).
				R().
				SetQueryParams(map[string]string{
			"name":    name,
			"content": content,
			"expire":  "0s",
		}).
		Get(*wechatURL + "/dev")
	if e != nil {
		err = e
		log.Printf("send notify for %v, content: %v err: %v\n", name, content, err)
		return
	}
	reply = string(resp.Body())
	return
}

// // make this into project config?
// func convert(name string) string {
// 	if name == "wenzhenglin" {
// 		return "wen"
// 	}
// 	return name
// }
