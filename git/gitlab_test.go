package git

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

// func TestGetToken(t *testing.T) {
// 	token, err := GetToken(USER, PASS)
// 	if err != nil {
// 		t.Error("verify err ", err)
// 		return
// 	}
// 	fmt.Println("token", token)
// }

func TestGetGroup(t *testing.T) {
	p, err := GetGroup("yunwei")
	if err != nil {
		t.Error("get groups err ", err)
		return
	}
	spew.Dump(p)
}

func TestGetUser(t *testing.T) {
	u, err := GetUser("robot")
	if err != nil {
		t.Error("get user err ", err)
		return
	}
	spew.Dump(u)
}

func TestCheckTagExist(t *testing.T) {
	u, err := CheckTagExist("robot/project-example", "v1.0.3-pre5.4dev")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	fmt.Println("tag:", u.Name)
	u, err = CheckTagExist("robot/project-example", "v1.a.3-pre5.4dev")
	if err == nil {
		t.Error("check tag err ", err)
		return
	}
	// fmt.Printf("err: %v\n", err)

	// spew.Dump(u)
}
func pretty(prefix string, a interface{}) {
	out, _ := prettyjson.Marshal(a)
	fmt.Printf("%v: %s\n", prefix, out)
}

/*
commit: {
  "author_email": "374207808@qq.com",
  "author_name": "robot",
  "authored_date": "2019-10-12T12:07:33+08:00",
  "committed_date": "2019-10-12T12:07:33+08:00",
  "committer_email": "374207808@qq.com",
  "committer_name": "robot",
  "created_at": "2019-10-12T12:07:33+08:00",
  "id": "85c2f5b5d951d617a89bfda3e2b140cfced5d097",
  "last_pipeline": null,
  "message": "Merge branch 'develop' into 'master'\n\nDevelop\n\nSee merge request wenzhenglin/project-example!2",
  "parent_ids": [
    "7e904f883d67dc7d5a4375e620f08013c51c154f",
    "a748c3b3871cb21c218eb64a7a60d9e6574523d4"
  ],
  "project_id": 308,
  "short_id": "85c2f5b5",
  "stats": {
    "additions": 165,
    "deletions": 86,
    "total": 251
  },
  "status": null,
  "title": "Merge branch 'develop' into 'master'"
}
*/

func TestSetCommitStatus(t *testing.T) {
	u, err := GetCommitFromTag("robot/project-example", "a748c3b3")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	pretty("commit", u)

	x, err := SetCommitStatusSuccess("robot/project-example", "a748c3b3", "manual")
	if err != nil {
		t.Error("set state err ", err)
		return
	}
	pretty("commit", x)

	u, err = GetCommitFromTag("robot/project-example", "a748c3b3")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	pretty("commit", u)
}
func TestGetCommitFromTag(t *testing.T) {

	u, err := GetCommitFromTag("robot/mileage-planet", "828ead19253b7e6214b6e29db83646c1f3167b1f")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	fmt.Println("commit:", u.ShortID, u.Title)
	pretty("commit", u)
	return
	u, err = GetCommitFromTag("robot/project-example", "v1.0.3-pre5.4dev")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	fmt.Println("commit:", u.ShortID, u.Title)

	u, err = GetCommitFromTag("robot/project-example", "v1.0.4-pre")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	fmt.Println("commit:", u.ShortID, u.Title)

	u, err = GetCommitFromTag("robot/project-example", "v1.0.4-pre")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	fmt.Println("commit:", u.ShortID, u.Title)

	// u, err = GetCommitFromTag("robot/project-example", "v1.a.3-pre5.4dev")
	// if err == nil {
	// 	t.Error("check tag err ", err)
	// 	return
	// }
	// fmt.Println("commit:", u.ShortID, u.Title)
}

func TestListAllTags(t *testing.T) {
	u, err := listAllTags("robot/mileage-planet")
	if err != nil {
		t.Error("check tag err ", err)
		return
	}
	for _, v := range u {
		fmt.Printf("created at: %v, name: %v\n", v.Commit.CreatedAt, v.Name)
	}
	// pretty("tags", u)
}
func TestListLastTwoCommits(t *testing.T) {
	u, err := listLastTwoCommits("robot/mileage-planet")
	if err != nil {
		t.Error("listLastTwoCommits err ", err)
		return
	}
	for _, v := range u {
		fmt.Printf("created at: %v, name: %v\n", v.CreatedAt, v.ShortID)
	}
	// pretty("tags", u)
}

func TestManualPushedImage(t *testing.T) {
	fmt.Println(time.Now())
	fmt.Println(time.Now().Unix())

	return
	imagetag, err := ManualPushedImage("robot/mileage-planet", "6775366d")
	if err != nil {
		t.Error("ManualPushedImage err: ", err)
		return
	}
	pretty("got imagetag", imagetag)
}

// func TestGetLastTagCommitIDg(t *testing.T) {
// 	o, p, err := GetLastTagCommitID("wenzhenglin/project-example")
// 	if err != nil {
// 		t.Error("check tag err ", err)
// 		return
// 	}
// 	fmt.Printf("onlineid: %v, preid: %v\n", o, p)
// }

// func TestGetLastTag(t *testing.T) {
// 	o, p, err := GetLastTag("wenzhenglin/project-example")
// 	if err != nil {
// 		t.Error("check tag err ", err)
// 		return
// 	}
// 	pretty("online", o)
// 	pretty("pre", p)
// }

func TestGetProject(t *testing.T) {
	p := "flow_center/tangguo"
	u, err := GetProject(p)
	if err != nil {
		t.Error("get project err ", err)
		return
	}
	if u.PathWithNamespace != p {
		t.Error("got wrong project err ")
		spew.Dump("p:", u)
		return
	}
	_, err = GetProject("m")
	if err == nil {
		t.Error("get project err, should not found ")
		return
	}
}

// func TestGetProjectMembers(t *testing.T) {

// 	// p := "flow_center/tangguo"
// 	p := "yunwei/config-deploy"
// 	err := GetProjectMembers(p)
// 	if err != nil {
// 		t.Error("get project err ", err)
// 		return
// 	}
// 	// if u.PathWithNamespace != p {
// 	// 	t.Error("got wrong project err ")
// 	// 	spew.Dump("p:", u)
// 	// 	return
// 	// }
// }

func TestCheckPerm(t *testing.T) {
	p, u, env := "flow_center/tangguo", "robot", "online"
	err := CheckPerm(p, u, env)
	if err != nil {
		t.Error("CheckPerm err", err)
		return
	}
	p, u, env = "flow_center/tangguo", "robot", "online1"
	err = CheckPerm(p, u, env)
	if err == nil {
		t.Error("CheckPerm err, shoud not allow", err)
		return
	}

}

// only got 20
func TestGetProjectsAdmin(t *testing.T) {
	ps, err := GetProjectsAdmin()
	// ps, err := GetProjectsAdmin(UserToken) // they need to use different list
	if err != nil {
		t.Error("err", err)
	}
	fmt.Println("got", len(ps))

	// sort.Slice(ps, func(i, j int) bool {
	// 	return ps[i].WebURL < ps[j].WebURL
	// })
	for _, v := range ps {
		fmt.Println(v.ID, v.WebURL)
	}
}

// only got one project
// func TestGetProjectsByUser(t *testing.T) {
// 	ps, err := GetProjectsByUser(UserToken)
// 	if err != nil {
// 		t.Error("err", err)
// 	}
// 	fmt.Println("got", len(ps))

// 	for _, v := range ps {
// 		fmt.Println(v.ID, v.WebURL)
// 	}
// }

// got 61
func TestGetProjectsUser(t *testing.T) {
	ps, err := GetProjectsUser(UserToken)
	if err != nil {
		t.Error("err", err)
	}
	fmt.Println("got", len(ps))

	for _, v := range ps {
		fmt.Println(v.ID, v.WebURL)
	}
}

// func TestGetProjectLists(t *testing.T) {
// 	_, ps, err := GetProjectLists(UserToken)
// 	if err != nil {
// 		t.Error("err", err)
// 	}
// 	fmt.Println("got", len(ps))
// }

// func TestListPersonalProjects(t *testing.T) {
// 	ps, err := listPersonalProjects2(adminToken)
// 	if err != nil {
// 		t.Error("err", err)
// 	}
// 	fmt.Println("got", len(ps))
// }
