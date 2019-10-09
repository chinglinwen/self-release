package project

import "testing"

func TestParseProjectConfigJson(t *testing.T) {
	all, err := ParseProjectConfigJson(democonfigjson)
	if err != nil {
		t.Error("ParseProjectConfigJson err", err)
		return
	}

	pretty(all)
}

var democonfigjson = `
{"selfrelease":{"devBranch":"test","version":"v1.0.0","configVer":"phpv1","buildMode":"auto","enable":true}}
`
