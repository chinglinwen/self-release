package git

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestNew(t *testing.T) {
	var repo *Repo
	var err error

	testfile := "/tmp/repos/wenzhenglin/test/testfile"
	err = ioutil.WriteFile(testfile, []byte("hello"), 0644)
	if err != nil {
		t.Error("writefile err", err)
		return
	}
	files, err := ioutil.ReadDir("/tmp/repos/wenzhenglin/test")
	if err != nil {
		t.Error("readdir err", err)
		return
	}

	for _, v := range files {
		fmt.Println(v.Name(), v.Mode(), v.Size())
	}
	repo, err = New("wenzhenglin/test", SetNoPull())
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

func TestFetch(t *testing.T) {
	var err error
	// _, err = New("wenzhenglin/test", SetBranch("feature2"))
	// if err != nil {
	// 	t.Error("new err:", err)
	// 	return
	// }
	// return
	// _, err = New("wenzhenglin/test")
	// if err != nil {
	// 	t.Error("new err:", err)
	// 	return
	// }

	_, err = New("yunwei/config-deploy")
	if err != nil {
		t.Error("new err:", err)
		return
	}
	//spew.Dump("r", r)
}

func TestCheckout(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("feature2"))
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

var (
	testfilename = "helo-test1"
)

func TestAdd(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("feature1"))
	if err != nil {
		t.Error("new err:", err)
		return
	}
	err = r.Add(testfilename, "hello from test4")
	if err != nil {
		t.Error("add err:", err)
		return
	}
}

func TestCommit(t *testing.T) {
	// r, err := New("wenzhenglin/test", "v1.0.0")
	r, err := New("wenzhenglin/test", SetBranch("feature2"))
	if err != nil {
		t.Error("new err:", err)
		return
	}

	err = r.CheckoutLocal()
	if err != nil {
		t.Error("checkout err:", err)
		return
	}

	err = r.Add(testfilename, "hello from test6")
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
