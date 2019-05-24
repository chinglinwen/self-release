package git

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestBranchIsTag(t *testing.T) {
	if BranchIsTag("develop") {
		t.Error("develop should not be a tag")
		return
	}
	if !BranchIsTag("v1.0.0") {
		t.Error("v1.0.0 should be a tag")
		return
	}
	if !BranchIsTag("v1.0.0-beta") {
		t.Error("v1.0.0-beta should be a tag")
		return
	}
	if !BranchIsTag("v1.0.0-alpha") {
		t.Error("v1.0.0-alpha should be a tag")
		return
	}
	// if !BranchIsTag("1.0.0") {
	// 	t.Error("1.0.0 should be a tag")
	// 	return
	// }
	// if BranchIsTag("x1.0.0") {
	// 	t.Error("x1.0.0 should be a tag")
	// 	return
	// }
	return
}

func TestFetch(t *testing.T) {
	repo, err := New("wenzhenglin/test", SetBranch("v1.0.5"))
	if err != nil {
		t.Error("new err", err)
		return
	}
	_ = repo
}
func TestNew(t *testing.T) {
	var repo *Repo
	var err error
	// repo, err = New("wenzhenglin/test", SetNoPull())
	repo, err = New("wenzhenglin/test", SetBranch("develop"))
	if err != nil {
		t.Error("new err", err)
		return
	}

	if !repo.IsExist("dev") {
		t.Error("new err dev not exist for develop branch")
		return
	}

	repo, err = New("wenzhenglin/test", SetBranch("feature1"))
	if err != nil {
		t.Error("new err", err)
		return
	}

	if !repo.IsExist("branch1") {
		t.Error("new err dev not exist for develop feature1")
		return
	}
}

func TestNewWithPull(t *testing.T) {
	var repo *Repo
	var err error
	// repo, err = New("wenzhenglin/test", SetNoPull())
	repo, err = NewWithPull("wenzhenglin/test", SetBranch("develop"))
	if err != nil {
		t.Error("new err", err)
		return
	}

	if !repo.IsExist("dev") {
		t.Error("new err dev not exist for develop branch")
		return
	}

	repo, err = NewWithPull("wenzhenglin/test", SetBranch("feature1"))
	if err != nil {
		t.Error("new err", err)
		return
	}

	if !repo.IsExist("branch1") {
		t.Error("new err dev not exist for develop feature1")
		return
	}
}
func TestNewWithLocalChange(t *testing.T) {
	var repo *Repo
	var err error

	testfile := "/home/wen/t/repos/wenzhenglin/test/testfile"
	err = ioutil.WriteFile(testfile, []byte("hello"), 0644)
	if err != nil {
		t.Error("writefile err", err)
		return
	}
	files, err := ioutil.ReadDir("/home/wen/t/repos/wenzhenglin/test")
	if err != nil {
		t.Error("readdir err", err)
		return
	}

	for _, v := range files {
		fmt.Println(v.Name(), v.Mode(), v.Size())
	}
	// repo, err = New("wenzhenglin/test", SetNoPull())
	repo, err = New("wenzhenglin/test", SetBranch("develop"))
	if err != nil {
		t.Error("new err", err)
		return
	}

	files, err = ioutil.ReadDir(repo.GetWorkDir())
	if err != nil {
		t.Error("readdir err", err)
		return
	}
	var exist bool
	for _, v := range files {
		if v.Name() == "testfile" {
			exist = true
		}
	}
	if !exist {
		t.Errorf("file %v not found after new", testfile)
		return
	}
}

// always discard local changes after new?
func TestCheckout1(t *testing.T) {
	var repo *Repo
	var err error

	testfile := "/home/wen/t/repos/wenzhenglin/test/testfile"
	// err = ioutil.WriteFile(testfile, []byte("hello"), 0644)
	// if err != nil {
	// 	t.Error("writefile err", err)
	// 	return
	// }
	// files, err := ioutil.ReadDir("/home/wen/t/repos/wenzhenglin/test")
	// if err != nil {
	// 	t.Error("readdir err", err)
	// 	return
	// }

	// for _, v := range files {
	// 	fmt.Println(v.Name(), v.Mode(), v.Size())
	// }

	// repo, err = New("wenzhenglin/test", SetNoPull())
	// repo, err = New("wenzhenglin/test", SetNoCheckout())
	repo, err = New("wenzhenglin/test")
	if err != nil {
		t.Error("new err", err)
		return
	}

	// err = repo.GitAdd("testfile")
	// if err != nil {
	// 	log.Println("gitadd err", err)
	// }

	files, err := ioutil.ReadDir(repo.GetWorkDir())
	if err != nil {
		t.Error("readdir err", err)
		return
	}
	var exist bool
	for _, v := range files {
		if v.Name() == "testfile" {
			exist = true
		}
	}
	if !exist {
		t.Errorf("file %v not found after new", testfile)
		return
	}
}

func TestClone(t *testing.T) {
	var repo *Repo
	var err error
	repo, _ = newrepo("wenzhenglin/test", SetNoPull())
	repo.CLone()
	if _, err := ioutil.ReadDir("wenzhenglin/test"); err != nil {
		t.Error("clone wenzhenglin/test err", err)
	}

	repo, _ = newrepo("yunwei/worktile")
	err = repo.CLone()
	if err != nil {
		t.Error("clone worktile err", err)
	}
	if _, err := ioutil.ReadDir("yunwei/worktile"); err != nil {
		t.Error("clone worktile err", err)
	}

	repo, _ = newrepo("yunwei/config-deploy")
	err = repo.CLone()
	if err != nil {
		t.Error("clone config-deploy err", err)
	}
	if _, err := ioutil.ReadDir("yunwei/config-deploy"); err != nil {
		t.Error("clone config-deploy err", err)
	}
}

func TestCheckout(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("develop"), SetForce())
	if err != nil {
		t.Error("new err:", err)
		return
	}
	err = r.CheckoutLocal()
	if err != nil {
		t.Error("checkout err:", err)
		return
	}

	// err = r.Checkout("refs/heads/feature1")
	// if err != nil {
	// 	t.Error("checkout err:", err)
	// 	return
	// }
}
func TestPull(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("develop"), SetForce())
	if err != nil {
		t.Error("new err:", err)
		return
	}
	err = r.Pull()
	if err != nil {
		t.Error("pull err:", err)
		return
	}

	// err = r.Checkout("refs/heads/feature1")
	// if err != nil {
	// 	t.Error("checkout err:", err)
	// 	return
	// }
}

var (
	testfilename = "_ops/helo-test1"
)

func TestCreate(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/project-example", SetNoPull())
	if err != nil {
		t.Error("new err:", err)
		return
	}
	err = r.Create(testfilename, "hello from test4")
	if err != nil {
		t.Error("add err:", err)
		return
	}
}

func TestCommit(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := NewWithPull("wenzhenglin/test", SetBranch("develop"))
	if err != nil {
		t.Error("new err:", err)
		return
	}

	// err = r.CheckoutLocal()
	// if err != nil {
	// 	t.Error("checkout err:", err)
	// 	return
	// }
	err = r.Add(testfilename+"1", "hello from test9")
	if err != nil {
		t.Error("add err:", err)
		return
	}

	err = r.Add(testfilename+"2", "hello from test9")
	if err != nil {
		t.Error("add err:", err)
		return
	}
	err = r.Commit("add " + testfilename + " from Push")
	if err != nil {
		t.Error("commit err:", err)
		return
	}
	err = r.Push()
	if err != nil {
		t.Error("push err:", err)
		return
	}
}

func TestPush(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("feature1"))
	if err != nil {
		t.Error("new err:", err)
		return
	}

	err = r.CheckoutLocal()
	if err != nil {
		t.Error("checkout err:", err)
		return
	}

	err = r.Push()
	if err != nil {
		t.Error("push err:", err)
		return
	}
}

func TestAddAndPush(t *testing.T) {
	// r, err := New("wenzhenglin/test", SetBranch("feature2"))
	r, err := New("wenzhenglin/test")
	if err != nil {
		t.Error("new err:", err)
		return
	}
	err = r.AddAndPush(testfilename, "hello from feature2 merge with existing 5", "after ccc upstream from AddAndPush 5")
	if err != nil {
		t.Error("push err:", err)
		return
	}
}
