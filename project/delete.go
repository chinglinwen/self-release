package project

import (
	"fmt"
	"os/exec"
)

/*
this error does not helpful?

 output: error: error validating "STDIN": error validating data: invalid object to validate; if you choose to ignore these errors, turn validation off with --validate=false
*/
// apply contents by generate? or let generate apply directly?
//
// how to apply
func DeleteByKubectl(project, branch, env string) (out string, err error) {
	if branch == "" {
		branch = "develop" // get this config?
	}
	if env == "" {
		env = GetEnvFromBranch(branch)
	}
	ns, p, err := GetProjectName(project)
	if err != nil {
		return
	}
	name := fmt.Sprintf("%v-%v", p, env)

	s := fmt.Sprintf("kubectl delete deploy -n %v %v", ns, name)
	cmd := exec.Command("sh", "-c", s)
	// cmd.Stdin = strings.NewReader(filebody)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("delete deploy: %v(ns: %v) err: %v, \noutput: %v", name, ns, err, string(output))
		return
	}
	out = string(output)
	return
}
