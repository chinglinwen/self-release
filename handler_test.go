package main

import (
	"encoding/json"
	"testing"
)

func TestParseRefs(t *testing.T) {
	if b := parseBranch("refs/tags/v1.0.2"); b != "v1.0.2" {
		t.Errorf("parse tag err, got %v, want %v\n", b, "v1.0.2")
	}
	if b := parseBranch("refs/heads/feature1"); b != "feature1" {
		t.Errorf("parse tag err, got %v, want %v\n", b, "feature1")
	}
}

func TestHandlePush(t *testing.T) {
	e := &PushEvent{}
	err := json.Unmarshal([]byte(examplePush), e)
	if err != nil {
		t.Error("unmarshal pushevent err", err)
		return
	}
	err = handlePush(e)
	if err != nil {
		t.Error("handlePush err", err)
		return
	}
}

func TestHandleTagPush(t *testing.T) {
	e := &TagPushEvent{}
	err := json.Unmarshal([]byte(exampleTagPush), e)
	if err != nil {
		t.Error("unmarshal pushevent err", err)
		return
	}
	err = handleRelease(e)
	if err != nil {
		t.Error("handlePush err", err)
		return
	}
}

var examplePush = `
{
	"after": "cb381a534cfcb6a90e421159ac2ea383f2de7f25",
	"before": "2adb55715b6e8b2e1fae1feb64f93d7fd572b672",
	"checkout_sha": "cb381a534cfcb6a90e421159ac2ea383f2de7f25",
	"commits": [
	  {
		"added": [
		  "devtest.txt"
		],
		"author": {
		  "email": "wenzhenglin@haodai.net",
		  "name": "wenzhenglin"
		},
		"id": "cb381a534cfcb6a90e421159ac2ea383f2de7f25",
		"message": "Add devfile",
		"modified": [],
		"removed": [],
		"timestamp": "2019-05-20T15:47:00+08:00",
		"url": "http://g.haodai.net/wenzhenglin/project-example/commit/cb381a534cfcb6a90e421159ac2ea383f2de7f25"
	  }
	],
	"event_name": "push",
	"message": null,
	"object_kind": "push",
	"project": {
	  "avatar_url": null,
	  "ci_config_path": null,
	  "default_branch": "master",
	  "description": "main-new as project example for test",
	  "git_http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
	  "git_ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
	  "homepage": "http://g.haodai.net/wenzhenglin/project-example",
	  "http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
	  "id": 308,
	  "name": "project-example",
	  "namespace": "wenzhenglin",
	  "path_with_namespace": "wenzhenglin/project-example",
	  "ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
	  "url": "git@g.haodai.net:wenzhenglin/project-example.git",
	  "visibility_level": 0,
	  "web_url": "http://g.haodai.net/wenzhenglin/project-example"
	},
	"project_id": 308,
	"ref": "refs/heads/develop",
	"repository": {
	  "description": "main-new as project example for test",
	  "git_http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
	  "git_ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
	  "homepage": "http://g.haodai.net/wenzhenglin/project-example",
	  "name": "project-example",
	  "url": "git@g.haodai.net:wenzhenglin/project-example.git",
	  "visibility_level": 0
	},
	"total_commits_count": 1,
	"user_avatar": "http://g.haodai.net/uploads/-/system/user/avatar/75/avatar.png",
	"user_email": "wenzhenglin@haodai.net",
	"user_id": 75,
	"user_name": "wenzhenglin",
	"user_username": "wenzhenglin"
  }
  `

var exampleTagPush = `
	{
    "after": "3ac0156a0821c5bba3ae68b2daea2a7dc0b119a8",
    "before": "0000000000000000000000000000000000000000",
    "checkout_sha": "3d04f23fb490b12def8bcdbe5be623970f9efae2",
    "commits": [{
        "added": [
            "_ops/config.env",
            "_ops/config.yaml",
            "_ops/nginx.conf",
            "build-docker.sh"
        ],
        "author": {
            "email": "john@doe.org",
            "name": "robot"
        },
        "id": "3d04f23fb490b12def8bcdbe5be623970f9efae2",
        "message": "generate final files for wenzhenglin/project-example",
        "modified": [],
        "removed": [],
        "timestamp": "2019-05-21T15:41:07+08:00",
        "url": "http://g.haodai.net/wenzhenglin/project-example/commit/3d04f23fb490b12def8bcdbe5be623970f9efae2"
    }],
    "event_name": "tag_push",
    "message": "v1.0.1 msg here",
    "object_kind": "tag_push",
    "project": {
        "avatar_url": null,
        "ci_config_path": null,
        "default_branch": "master",
        "description": "main-new as project example for test",
        "git_http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
        "git_ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
        "homepage": "http://g.haodai.net/wenzhenglin/project-example",
        "http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
        "id": 308,
        "name": "project-example",
        "namespace": "wenzhenglin",
        "path_with_namespace": "wenzhenglin/project-example",
        "ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
        "url": "git@g.haodai.net:wenzhenglin/project-example.git",
        "visibility_level": 0,
        "web_url": "http://g.haodai.net/wenzhenglin/project-example"
    },
    "project_id": 308,
    "ref": "refs/tags/v1.0.1",
    "repository": {
        "description": "main-new as project example for test",
        "git_http_url": "http://g.haodai.net/wenzhenglin/project-example.git",
        "git_ssh_url": "git@g.haodai.net:wenzhenglin/project-example.git",
        "homepage": "http://g.haodai.net/wenzhenglin/project-example",
        "name": "project-example",
        "url": "git@g.haodai.net:wenzhenglin/project-example.git",
        "visibility_level": 0
    },
    "total_commits_count": 1,
    "user_avatar": "http://g.haodai.net/uploads/-/system/user/avatar/75/avatar.png",
    "user_email": "wenzhenglin@haodai.net",
    "user_id": 75,
    "user_name": "wenzhenglin",
    "user_username": "wenzhenglin"
}
	`
