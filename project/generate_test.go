package project

import (
	"fmt"
	"testing"
)

// func TestGenerateAll(t *testing.T) {
// 	p, err := NewProject(exampleproject, SetBranch("develop"))
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	// if !p.GetRepo().IsExist("_ops/config.yaml") {
// 	// 	t.Error("not inited")
// 	// 	return
// 	// }

// 	target, err := p.Generate()
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// 	fmt.Println("target", target)
// }
func TestConvertToSubst(t *testing.T) {
	m := map[string]string{
		"CI_PROJECT_NAME_WITH_ENV": "envaa",
		"CI_NAMESPACE":             "ns",
	}
	a := `
apiVersion: apps/v1
kind: Deployment
test: ${CI_PROJECT_NAME_WITH_ENV}
metadata:
  name: {{ $CI_PROJECT_NAME_WITH_ENV }}
  namespace: {{ $CI_NAMESPACE }}
spec:
  annotations:
    namespace : {{ $CI_NAMESPACE }}
    project: {{ $CI_PROJECT_NAME }}
    publish_user: {{ $CI_USER_NAME }}
	publish_at: {{ $CI_TIME }}`
	s := convertToSubst(a)
	fmt.Printf("before:\n%v\nafter:\n%v\n", a, s)

	f, err := generateByMap(s, m)
	if err != nil {
		t.Error("generateByMap err", err)
		return
	}
	fmt.Printf("final: %v\n", f)
}

// func TestGenerateConfig(t *testing.T) {
// 	p, err := NewProject(exampleproject) //, SetInitForce())
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("config.yaml"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateNginx(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("nginx.conf"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGeneratePHP(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("php.ini"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateConfigEnv(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("config.env"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestReadEnv(t *testing.T) {
// 	fmt.Println("what=", os.Getenv("WHAT"))
// 	err := readEnvs([]string{"/home/wen/t/repos/wenzhenglin/project-example/_ops/config.env"})
// 	if err != nil {
// 		t.Errorf("readenvs err: %v", err)
// 		return
// 	}
// 	fmt.Println("after read envwhat=", os.Getenv("EXTRA"))
// }

// func TestGenerateDocker(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("dockerfile"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateBuildDocker(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("build-docker.sh"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateK8sOnline(t *testing.T) {
// 	autoenv := make(map[string]string)
// 	autoenv["PROJECTPATH"] = "PROJECTPATHaa"
// 	autoenv["BRANCH"] = "BRANCHaa"
// 	// autoenv["USERNAME"] = event.UserName
// 	// autoenv["USEREMAIL"] = event.UserEmail
// 	autoenv["MSG"] = "msg11"

// 	p, err := NewProject(exampleproject, SetBranch("develop"))
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("k8s-online"), SetGenAutoEnv(autoenv))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateK8sPre(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("k8s-pre"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGenerateK8sTest(t *testing.T) {
// 	p, err := NewProject(exampleproject)
// 	if err != nil {
// 		t.Error("newproject err", err)
// 		return
// 	}
// 	_, err = p.Generate(SetGenerateName("k8s-test"))
// 	if err != nil {
// 		t.Error("generate err", err)
// 		return
// 	}
// }

// func TestGetEnvFromBranch(t *testing.T) {
// 	if env := GetEnvFromBranch("develop"); env != TEST {
// 		t.Error("develop should be env test, got ", env)
// 		return
// 	}
// 	if GetEnvFromBranch("v1.0.0") != ONLINE {
// 		t.Error("v1.0.0 should be env online")
// 		return
// 	}
// 	if GetEnvFromBranch("v1.0.0.") == ONLINE {
// 		t.Error("v1.0.0. should not be online")
// 		return
// 	}
// 	if GetEnvFromBranch("v1.0.0a") != PRE {
// 		t.Error("v1.0.0a should env pre")
// 		return
// 	}
// 	if GetEnvFromBranch("v1.0.0-beta") != PRE {
// 		t.Error("v1.0.0-beta should be env pre")
// 		return
// 	}
// 	if GetEnvFromBranch("v1.0.0-alpha") != PRE {
// 		t.Error("v1.0.0-alpha should be env pre")
// 		return
// 	}
// }
