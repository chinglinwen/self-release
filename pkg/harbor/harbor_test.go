package harbor

import (
	"testing"
)

func TestListProject(t *testing.T) {
	ps, err := ListProjects()
	if err != nil {
		t.Error("list project err", err)
		return
	}
	pretty("ps", ps)
}

// create project err request failed, status: 401 Unauthorized
func TestCreateProject(t *testing.T) {
	err := CreateProject("aaa")
	if err != nil {
		t.Error("create project err", err)
		return
	}
}

func TestCreateProjectIfNotExist(t *testing.T) {
	created, err := CreateProjectIfNotExist("wenzhenglin")
	if err != nil {
		t.Error("CreateProjectIfNotExist project err", err)
		return
	}
	if created != false {
		t.Error("project wenzhenglin exist, shoult not create")
		return
	}
	// log.Println("created", created)

	created, err = CreateProjectIfNotExist("aaa")
	if err != nil {
		t.Error("CreateProjectIfNotExist project err", err)
		return
	}
	if created != true {
		t.Error("project aaa exist, shoult not create")
		return
	}

}

func TestCheckProject(t *testing.T) {
	if ok, _ := CheckProject("wenzhenglin"); !ok {
		t.Error("project wenzhenglin should exist")
		return
	}

	if ok, _ := CheckProject("a"); ok {
		t.Error("project a should not exist")
		return
	}

}
