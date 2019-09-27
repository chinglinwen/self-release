package project

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/instrumenta/kubeval/kubeval"
)

// this need to access internet?
// eg. https://kubernetesjsonschema.dev/master-standalone/service-v1.json
// try replace with kubectl apply -dry -f ?
/*

	[$ kubectl apply --dry-run -f fstest.yaml
	namespace/t created (dry run)
	service/fs created (dry run)
	deployment.apps/fs created (dry run)
	[$ kubectl apply --server-dry-run -f fstest.yaml
	namespace/t created (server dry run)
	Error from server (NotFound): error when creating "fstest.yaml": namespaces "t" not found
	Error from server (NotFound): error when creating "fstest.yaml": namespaces "t" not found
	[$

	$ kubectl apply --dry-run -f fstest.yaml
	service/fs created (dry run)
	deployment.apps/fs created (dry run)
	error: unable to recognize "fstest.yaml": no matches for kind "Namespace" in version "v1aaa aa"
	[$ vi fstest.yaml
	[$ kubectl apply --dry-run -f fstest.yaml
	service/fs created (dry run)
	deployment.apps/fs created (dry run)
	error: unable to recognize "fstest.yaml": no matches for kind "Namespace" in version "v1aaa"
	[$
*/
func ValidateByKubeval(filebody, fileName string) ([]kubeval.ValidationResult, error) {
	return kubeval.Validate([]byte(filebody), fileName)
}

/*

this need to connect to k8s cluster

$ export KUBECONFIG=/t
$ kubectl apply --dry-run -f fstest.yaml
unable to recognize "fstest.yaml": Get http://localhost:8080/api?timeout=32s: dial tcp 127.0.0.1:8080: connect: connection refused
unable to recognize "fstest.yaml": Get http://localhost:8080/api?timeout=32s: dial tcp 127.0.0.1:8080: connect: connection refused
unable to recognize "fstest.yaml": Get http://localhost:8080/api?timeout=32s: dial tcp 127.0.0.1:8080: connect: connection refused

*/
// validate by kubectl, filename is for error
func ValidateByKubectlWithString(filebody string) (out string, err error) {
	cmd := exec.Command("sh", "-c", "kubectl apply --dry-run -f -")
	cmd.Stdin = strings.NewReader(filebody)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("validate err: %v, \noutput: %v", err, string(output))
		return
	}
	out = string(output)
	return
}
