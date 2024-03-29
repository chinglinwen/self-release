package project

import (
	"fmt"
	"testing"
)

func TestHelmGen(t *testing.T) {
	out, err := HelmGen("haodai/main", "pre")
	if err != nil {
		t.Error("HelmGen err", err)
		return
	}

	fmt.Println("out: ", out)
}

func TestHelmGenPrintValidateYaml(t *testing.T) {
	out, err := HelmGenPrintValidateYaml("robot/mileage-planet", "pre")
	if err != nil {
		t.Error("HelmGenPrintValidateYaml err", err)
		return
	}
	fmt.Println("out: ", out)
}

func TestGenExampleYaml(t *testing.T) {
	out, final, err := genExampleYaml("robot/mileage-planet", "test")
	if err != nil {
		t.Error("genExampleYaml err", err)
		return
	}
	fmt.Println("out: ", out, "final", final)
}

/*
    /home/wen/gocode/src/wen/self-release/project/generate_yaml_test.go:22: HelmGen err "\x1b[\xae": unknown escape sequence
		"\x1b[\xae": unknown escape sequence
*/
func TestRunHelmGen(t *testing.T) {
	// dir := "/home/wen/gocode/src/wen/self-release/repos/wenzhenglin/config-deploy"
	// dir := "/home/wen/t/repos/yunwei/config-deploy"
	dir := "/home/wen/git/yunwei/config-deploy"
	out, err := runHelmGen(dir, "robot/mileage-planet", "test", GENSHPRINT)
	if err != nil {
		t.Error("HelmGen err", err)
		return
	}

	fmt.Println("out: ", out)
}
