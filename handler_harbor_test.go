package main

import (
	"fmt"
	"testing"
)

var okbody = `
{
	"events": [
	   {
		  "id": "32dbedd7-ac9c-458d-bc79-e936385e9547",
		  "timestamp": "2019-07-15T14:27:44.932143541Z",
		  "action": "push",
		  "target": {
			 "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			 "size": 524,
			 "digest": "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a",
			 "length": 524,
			 "repository": "ops/hello-world",
			 "url": "https://harbor.haodai.net/v2/ops/hello-world/manifests/sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a",
			 "tag": "v1"
		  },
		  "request": {
			 "id": "64a71f58-1d56-4dba-8f64-7db90c781899",
			 "addr": "192.168.10.234",
			 "host": "harbor.haodai.net",
			 "method": "PUT",
			 "useragent": "docker/18.06.1-ce go/go1.10.3 git-commit/e68fc7a kernel/3.10.0-957.el7.x86_64 os/linux arch/amd64 UpstreamClient(Docker-Client/18.06.1-ce \\(linux\\))"
		  },
		  "actor": {
			 "name": "wenzhenglin"
		  },
		  "source": {
			 "addr": "539a26a4b289:5000",
			 "instanceID": "b0230325-bafa-4ed2-8b85-97a4c5b259bb"
		  }
	   }
	]
 }`

var badbody = `
 {
	"events": [
	   {
		  "id": "3e678747-d3e5-4741-82f3-c16c28787cfe",
		  "timestamp": "2019-07-15T14:27:44.714653331Z",
		  "action": "pull",
		  "target": {
			 "mediaType": "application/octet-stream",
			 "size": 977,
			 "digest": "sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced",
			 "length": 977,
			 "repository": "ops/hello-world",
			 "url": "https://harbor.haodai.net/v2/ops/hello-world/blobs/sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced"
		  },
		  "request": {
			 "id": "6253c07f-9ea3-48d0-9fb3-cc53104208d8",
			 "addr": "192.168.10.234",
			 "host": "harbor.haodai.net",
			 "method": "HEAD",
			 "useragent": "docker/18.06.1-ce go/go1.10.3 git-commit/e68fc7a kernel/3.10.0-957.el7.x86_64 os/linux arch/amd64 UpstreamClient(Docker-Client/18.06.1-ce \\(linux\\))"
		  },
		  "actor": {
			 "name": "wenzhenglin"
		  },
		  "source": {
			 "addr": "539a26a4b289:5000",
			 "instanceID": "b0230325-bafa-4ed2-8b85-97a4c5b259bb"
		  }
	   }
	]
 }`

var badbody1 = `
 {
	"events": [
	   {
		  "id": "3edbf681-f555-4820-be93-c9e171371aa1",
		  "timestamp": "2019-07-15T14:27:44.877477534Z",
		  "action": "push",
		  "target": {
			 "mediaType": "application/octet-stream",
			 "size": 1510,
			 "digest": "sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e",
			 "length": 1510,
			 "repository": "ops/hello-world",
			 "url": "https://harbor.haodai.net/v2/ops/hello-world/blobs/sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e"
		  },
		  "request": {
			 "id": "5e6bca8e-efc8-4b25-ab19-3e4f182850ee",
			 "addr": "192.168.10.234",
			 "host": "harbor.haodai.net",
			 "method": "PUT",
			 "useragent": "docker/18.06.1-ce go/go1.10.3 git-commit/e68fc7a kernel/3.10.0-957.el7.x86_64 os/linux arch/amd64 UpstreamClient(Docker-Client/18.06.1-ce \\(linux\\))"
		  },
		  "actor": {
			 "name": "wenzhenglin"
		  },
		  "source": {
			 "addr": "539a26a4b289:5000",
			 "instanceID": "b0230325-bafa-4ed2-8b85-97a4c5b259bb"
		  }
	   }
	]
 }`

func TestUnmarshalHarborEvent(t *testing.T) {
	e, err := unmarshalHarborEvent(badbody1)
	if err != nil {
		t.Error("unmarshal body err", err)
		return
	}
	fmt.Printf("type: %v\n", e.Events[0].Target.MediaType)
	fmt.Printf("tag: %v\n", e.Events[0].Target.Tag)
	pretty(e)
}
