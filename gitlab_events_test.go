package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParseEvent(t *testing.T) {
	data, err := ParseEvent("push", []byte(pushEvent))
	if err != nil {
		t.Error("parse err", err)
		return
	}
	e, ok := data.(*PushEvent)
	if !ok {
		t.Error("cast to push event err")
		return
	}
	spew.Dump("event", e)
}

var pushEvent = ` {
  "after": "1151ff0d392edec7ba8091a0bd3456f8c4110095",
  "before": "adeb26a4a024541729e7d7626ca183dda55f5c39",
  "checkout_sha": "1151ff0d392edec7ba8091a0bd3456f8c4110095",
  "commits": [
    {
      "added": [],
      "author": {
        "email": "wenzhenglin@haodai.net",
        "name": "wenzhenglin"
      },
      "id": "1151ff0d392edec7ba8091a0bd3456f8c4110095",
      "message": "Update helo-test1",
      "modified": [
        "helo-test1"
      ],
      "removed": [],
      "timestamp": "2019-05-13T14:28:14+08:00",
      "url": "http://g.haodai.net/wenzhenglin/test/commit/1151ff0d392edec7ba8091a0bd3456f8c4110095"
    }
  ],
  "event_name": "push",
  "message": null,
  "object_kind": "push",
  "project": {
    "avatar_url": null,
    "ci_config_path": null,
    "default_branch": "master",
    "description": "test",
    "git_http_url": "http://g.haodai.net/wenzhenglin/test.git",
    "git_ssh_url": "git@g.haodai.net:wenzhenglin/test.git",
    "homepage": "http://g.haodai.net/wenzhenglin/test",
    "http_url": "http://g.haodai.net/wenzhenglin/test.git",
    "id": 290,
    "name": "test",
    "namespace": "wenzhenglin",
    "path_with_namespace": "wenzhenglin/test",
    "ssh_url": "git@g.haodai.net:wenzhenglin/test.git",
    "url": "git@g.haodai.net:wenzhenglin/test.git",
    "visibility_level": 20,
    "web_url": "http://g.haodai.net/wenzhenglin/test"
  },
  "project_id": 290,
  "ref": "refs/heads/master",
  "repository": {
    "description": "test",
    "git_http_url": "http://g.haodai.net/wenzhenglin/test.git",
    "git_ssh_url": "git@g.haodai.net:wenzhenglin/test.git",
    "homepage": "http://g.haodai.net/wenzhenglin/test",
    "name": "test",
    "url": "git@g.haodai.net:wenzhenglin/test.git",
    "visibility_level": 20
  },
  "total_commits_count": 1,
  "user_avatar": "http://g.haodai.net/uploads/-/system/user/avatar/75/avatar.png",
  "user_email": "wenzhenglin@haodai.net",
  "user_id": 75,
  "user_name": "wenzhenglin",
  "user_username": "wenzhenglin"
}
`

var tagPush = `

{
  "object_kind": "tag_push",
  "event_name": "tag_push",
  "before": "0000000000000000000000000000000000000000",
  "after": "446bae8f4c3d6cc66900af3b524f3f19f29c0b67",
  "ref": "refs\/tags\/v1.0.1",
  "checkout_sha": "adeb26a4a024541729e7d7626ca183dda55f5c39",
  "message": "small update",
  "user_id": 75,
  "user_name": "wenzhenglin",
  "user_username": "wenzhenglin",
  "user_email": "wenzhenglin@haodai.net",
  "user_avatar": "http:\/\/g.haodai.net\/uploads\/-\/system\/user\/avatar\/75\/avatar.png",
  "project_id": 290,
  "project": {
    "id": 290,
    "name": "test",
    "description": "test",
    "web_url": "http:\/\/g.haodai.net\/wenzhenglin\/test",
    "avatar_url": null,
    "git_ssh_url": "git@g.haodai.net:wenzhenglin\/test.git",
    "git_http_url": "http:\/\/g.haodai.net\/wenzhenglin\/test.git",
    "namespace": "wenzhenglin",
    "visibility_level": 20,
    "path_with_namespace": "wenzhenglin\/test",
    "default_branch": "master",
    "ci_config_path": null,
    "homepage": "http:\/\/g.haodai.net\/wenzhenglin\/test",
    "url": "git@g.haodai.net:wenzhenglin\/test.git",
    "ssh_url": "git@g.haodai.net:wenzhenglin\/test.git",
    "http_url": "http:\/\/g.haodai.net\/wenzhenglin\/test.git"
  },
  "commits": [
    {
      "id": "adeb26a4a024541729e7d7626ca183dda55f5c39",
      "message": "after ccc upstream from AddAndPush after 4",
      "timestamp": "2019-05-10T14:05:24+08:00",
      "url": "http:\/\/g.haodai.net\/wenzhenglin\/test\/commit\/adeb26a4a024541729e7d7626ca183dda55f5c39",
      "author": {
        "name": "robot",
        "email": "john@doe.org"
      },
      "added": [
        "helo-test1"
      ],
      "modified": [
        
      ],
      "removed": [
        
      ]
    }
  ],
  "total_commits_count": 1,
  "repository": {
    "name": "test",
    "url": "git@g.haodai.net:wenzhenglin\/test.git",
    "description": "test",
    "homepage": "http:\/\/g.haodai.net\/wenzhenglin\/test",
    "git_http_url": "http:\/\/g.haodai.net\/wenzhenglin\/test.git",
    "git_ssh_url": "git@g.haodai.net:wenzhenglin\/test.git",
    "visibility_level": 20
  }
}
`
var repositoryUpdate = ` {
  "changes": [
    {
      "after": "dbd1e6bcea75f2ee19238c36185bed46568557d1",
      "before": "8e0243551b079a4c4d35e0bbb81990d56d1218d1",
      "ref": "refs/heads/master"
    }
  ],
  "event_name": "repository_update",
  "project": {
    "avatar_url": null,
    "ci_config_path": null,
    "default_branch": "master",
    "description": "",
    "git_http_url": "http://g.haodai.net/xindaiquan/base-service.git",
    "git_ssh_url": "git@g.haodai.net:xindaiquan/base-service.git",
    "homepage": "http://g.haodai.net/xindaiquan/base-service",
    "http_url": "http://g.haodai.net/xindaiquan/base-service.git",
    "id": 298,
    "name": "????",
    "namespace": "???",
    "path_with_namespace": "xindaiquan/base-service",
    "ssh_url": "git@g.haodai.net:xindaiquan/base-service.git",
    "url": "git@g.haodai.net:xindaiquan/base-service.git",
    "visibility_level": 0,
    "web_url": "http://g.haodai.net/xindaiquan/base-service"
  },
  "project_id": 298,
  "refs": [
    "refs/heads/master"
  ],
  "user_avatar": "https://www.gravatar.com/avatar/eae75e2f72dbeaa947ab2c8408cc07d2?s=80&d=identicon",
  "user_email": "yangzhuo@haodai.net",
  "user_id": 50,
  "user_name": "yangzhuo"
}
`