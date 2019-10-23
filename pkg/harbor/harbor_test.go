package harbor

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestListProject(t *testing.T) {
	ps, err := ListProjects()
	if err != nil {
		t.Error("list project err", err)
		return
	}
	pretty("ps", ps)
}

func TestListRepoTagLatest(t *testing.T) {
	ts := "2019-10-22_15:00:00"

	v, err := ListRepoTagLatest("robot/mileage-planet", ts)
	if err != nil {
		t.Error("ListRepoTags err", err)
		return
	}

	fmt.Printf("tag: %v, time: %v\n", v.Name, v.Created)

	// pretty("ps", ps)
}
func TestListRepoTags(t *testing.T) {
	ps, err := ListRepoTags("robot/mileage-planet")
	if err != nil {
		t.Error("ListRepoTags err", err)
		return
	}
	for _, v := range ps {
		fmt.Printf("tag: %v, time: %v\n", v.Created, v.Name)
	}
	// pretty("ps", ps)
}

// func TestListRepoThreeTags(t *testing.T) {
// 	ps, err := ListRepoThreeTags("robot/mileage-planet")
// 	if err != nil {
// 		t.Error("ListRepoTags err", err)
// 		return
// 	}
// 	for _, v := range ps {
// 		fmt.Printf("tag: %v, time: %v\n", v.Created, v.Name)
// 	}
// 	// pretty("ps", ps)
// }

// create project err request failed, status: 401 Unauthorized
func TestCreateProject(t *testing.T) {
	err := CreateProject("robot")
	if err != nil {
		t.Error("create project err", err)
		return
	}
}

func TestCreateProjectIfNotExist(t *testing.T) {
	var err error
	// err := CreateProjectIfNotExist("wenzhenglin")
	// if err != nil {
	// 	t.Error("CreateProjectIfNotExist project err", err)
	// 	return
	// }
	// log.Println("created", created)

	err = CreateProjectIfNotExist("robot")
	if err != nil {
		t.Error("CreateProjectIfNotExist project err", err)
		return
	}
}
func TestDeleteProject(t *testing.T) {
	err := DeleteProject("aaa1")
	if err != nil {
		t.Error("DeleteProject err", err)
		return
	}
}
func TestCheckProject(t *testing.T) {
	if _, err := CheckProject("wenzhenglin"); err != nil {
		t.Error("check project wenzhenglin err", err)
		return
	}

	// if _, err := CheckProject("aaa1"); err != nil {
	// 	t.Error("check project aaa1 err", err)
	// 	return
	// }

	if _, err := CheckProject("a"); err == nil {
		t.Error("project a should not exist")
		return
	}

}

/*
([]harbor.TagResp) (len=5 cap=6) {
 (harbor.TagResp) {
  tagDetail: (harbor.tagDetail) {
   Digest: (string) (len=71) "sha256:5eb84c1d5dcefe528d6e78e474cb2039c72f4a731d4606e7da1f1223186996e1",
   Name: (string) (len=7) "develop",
   Size: (int64) 99919313,
   Architecture: (string) (len=5) "amd64",
   OS: (string) (len=5) "linux",
   DockerVersion: (string) (len=10) "18.06.2-ce",
   Author: (string) "",
   Created: (time.Time) 2019-07-08 07:43:20.317842853 +0000 UTC,
   Config: (*harbor.cfg)(0xc000146180)({
    Labels: (map[string]string) <nil>
   })
  },
  Signature: (*harbor.Signature)(<nil>),
  ScanOverview: (*harbor.ImgScanOverview)(<nil>)
 },
 (harbor.TagResp) {
  tagDetail: (harbor.tagDetail) {
   Digest: (string) (len=71) "sha256:f2bd2483c7067737f3240b06e23c13b8c8027f523baec292704070dfc670b6be",
   Name: (string) (len=11) "v1.0.3-pre3",
   Size: (int64) 99918186,
   Architecture: (string) (len=5) "amd64",
   OS: (string) (len=5) "linux",
   DockerVersion: (string) (len=10) "18.06.1-ce",
   Author: (string) "",
   Created: (time.Time) 2019-05-28 10:23:23.165133328 +0000 UTC,
   Config: (*harbor.cfg)(0xc000146188)({
    Labels: (map[string]string) <nil>
   })
  },
  Signature: (*harbor.Signature)(<nil>),
  ScanOverview: (*harbor.ImgScanOverview)(<nil>)
 },
*/
func TestA(t *testing.T) {
	// defaultClient.Repositories.ListRepositoryTags("project-example")
	// a, r, err := defaultClient.Repositories.ListRepository(&harbor.ListRepositoriesOption{Q: "example"})
	// fmt.Println(a, r, err)

	// spew.Dump(defaultClient.Search())

	spew.Dump(defaultClient.Repositories.ListRepositoryTags("wenzhenglin/project-example"))
}

func TestRepoTagIsExist(t *testing.T) {
	cases := []struct {
		repo, tag string
		exist     bool
	}{
		{repo: "robot/project-example", tag: "a3a9cbff", exist: true},
		{repo: "wenzhenglin/project-example", tag: "v1.0.3-pre5.4dev1", exist: false},
	}
	for _, v := range cases {
		exist, err := RepoTagIsExist(v.repo, v.tag)
		if err != nil {
			t.Error("check tag err", err)
			return
		}
		if exist != v.exist {
			tags, err := ListRepoTags(v.repo)
			if err != nil {
				t.Error("ListRepoTags err", err)
				return
			}
			// pretty("tags", tags)
			for _, v := range tags {
				fmt.Printf("tag: %v\n", v.Name)
			}
			t.Errorf("%v:%v it should exist: %v, got: %v", v.repo, v.tag, v.exist, exist)
			return
		}
	}
	return
}
