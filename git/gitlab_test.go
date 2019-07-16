package git

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
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