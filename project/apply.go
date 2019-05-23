package project

import (
	"fmt"
	"os/exec"
	"strings"
)

// apply contents by generate? or let generate apply directly?
//
// how to apply
func ApplyByKubectl(filebody, fileName string) (out string, err error) {
	cmd := exec.Command("sh", "-c", "kubectl apply -f -")
	cmd.Stdin = strings.NewReader(filebody)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("apply file: %v err: %v, \noutput: %v", fileName, err, string(output))
		return
	}
	out = string(output)
	return
}

// create ns if not exist
func CheckOrCreateNamespace(ns string) (out string, err error) {
	s := fmt.Sprintf("kubectl get ns %v || kubectl create ns %v", ns)
	cmd := exec.Command("sh", "-c", s)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("check or create ns: %v err: %v, \noutput: %v", ns, err, string(output))
		return
	}
	out = string(output)
	return
}
