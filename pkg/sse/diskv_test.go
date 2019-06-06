package sse

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestReadFile(t *testing.T) {
	b := &Broker{}
	// b.ExistMsg = []string{}
	err := json.Unmarshal([]byte(examplejson), b)
	if err != nil {
		t.Error("unmarshal err", err)
		return
	}
	if len(b.ExistMsg) == 0 {
		t.Error("unmarshal ExistMsg is empty")
		return
	}
	spew.Dump(b)
}

var examplejson = `
{
	"Project": "wenzhenglin/project-example",
	"Key": "wenzhenglin-project-example-2019-6-6_19:38:46",
	"Branch": "develop",
	"ExistMsg": [
	  "\u003ch1\u003ecreated log for project: wenzhenglin/project-example\u003c/h1\u003e",
	  "starting logs",
	  "start build for project wenzhenglin/project-example, branch: develop, env: test\n",
	  "end."
	],
	"CreateTime": "2019-6-6_19:38:46",
	"Stored": true
  }`
