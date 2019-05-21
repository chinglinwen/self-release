package template

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerateAll(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate()
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateConfig(t *testing.T) {
	p, err := NewProject(exampleproject) //, SetInitForce())
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("config.yaml"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateNginx(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("nginx.conf"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGeneratePHP(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("php.ini"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateConfigEnv(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("config.env"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestReadEnv(t *testing.T) {
	fmt.Println("what=", os.Getenv("WHAT"))
	err := readEnvs([]string{"/home/wen/t/repos/wenzhenglin/project-example/_ops/config.env"})
	if err != nil {
		t.Errorf("readenvs err: %v", err)
		return
	}
	fmt.Println("after read envwhat=", os.Getenv("EXTRA"))
}

func TestGenerateDocker(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("dockerfile"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateBuildDocker(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("build-docker.sh"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateK8sOnline(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("k8s-online"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateK8sPre(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("k8s-pre"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}

func TestGenerateK8sTest(t *testing.T) {
	p, err := NewProject(exampleproject)
	if err != nil {
		t.Error("newproject err", err)
		return
	}
	err = p.Generate(SetGenerateName("k8s-test"))
	if err != nil {
		t.Error("generate err", err)
		return
	}
}
