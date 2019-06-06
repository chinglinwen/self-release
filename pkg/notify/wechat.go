package notify

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
	"time"

	resty "gopkg.in/resty.v1"
)

var (
	wechatNotifyURL = flag.String("w", "http://localhost:8001", "wechat notify service url")
	receiver        = flag.String("r", "", "default wechat receiver")
	receiverParty   = flag.String("party", "", "default receiver party ( eg. 3 )")
	agentid         = flag.String("agentid", "", "default agentid ( eg. 1000003 )")
	secret          = flag.String("secret", "", "default secret ( eg. G5h7CTEqkBw-Fe3luf2JM8UNNJAcYTpbXvpveY7M3lg )")

	expire = flag.String("e", "10m", "default expire time duration")
)

type sendconfig struct {
	touser  string
	toparty string
}

type sendoption func(*sendconfig)

// both touser and toparty
func SetReceiver(receiver string) sendoption {
	return func(c *sendconfig) {
		if regexp.MustCompile(`^[0-9]+$`).MatchString(receiver) {
			c.toparty = receiver
			c.touser = ""
			return
		}
		c.touser = receiver
		c.toparty = ""
	}
}

func SendPerson(message, person string) (reply string, err error) {
	return Send(message, SetReceiver(person))
}

func Send(message string, options ...sendoption) (reply string, err error) {
	c := &sendconfig{
		touser:  *receiver,
		toparty: *receiverParty,
	}
	for _, option := range options {
		option(c)
	}
	now := time.Now().Format("2006-1-2 15:04:05")
	precontent := fmt.Sprintf("时间: %v\n", now)

	r := strings.NewReplacer("\"", " ", "{", "", "}", "")
	message = r.Replace(message)

	resp, e := resty.R().
		SetQueryParams(map[string]string{
			"user":       c.touser,
			"toparty":    c.toparty,
			"agentid":    *agentid,
			"secret":     *secret,
			"precontent": precontent,
			"content":    message,
			"expire":     *expire,
		}).
		Get(*wechatNotifyURL)

	if e != nil {
		err = e
		return
	}
	reply = string(resp.Body())
	return
}

// func checkandsend(message string, options ...sendoption) (reply string, err error) {
// 	if !startsend {
// 		if !time.Now().After(starttime.Add(5 * time.Second)) {
// 			return "skip send at start time", nil
// 		}
// 	}
// 	return send(message, options...)
// }
