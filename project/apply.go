package project

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/chinglinwen/log"
)

func Apply(project, env string, envMap map[string]string) (out string, err error) {
	final, err := HelmGenPrintFinal(project, env, envMap)
	if err != nil {
		return
	}
	return ApplyByKubectlWithString(final)
}

func Delete(project, env string, envMap map[string]string) (out string, err error) {
	final, err := HelmGenPrintFinal(project, env, envMap)
	if err != nil {
		return
	}
	return DeleteByKubectlWithString(final)
}

/*
this error does not helpful?

 output: error: error validating "STDIN": error validating data: invalid object to validate; if you choose to ignore these errors, turn validation off with --validate=false
*/
// apply contents by generate? or let generate apply directly?
//
// how to apply
// func ApplyByKubectl(filebody string) (out string, err error) {
// 	s := fmt.Sprintf("kubectl apply -f -")
// 	cmd := exec.Command("sh", "-c", s)
// 	cmd.Stdin = strings.NewReader(filebody)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		err = fmt.Errorf("k8s apply err: %v, \noutput: %v", err, string(output))
// 		return
// 	}
// 	out = string(output)
// 	return
// }

// ApplyByKubectl apply k8s yaml.
// func ApplyByKubectl(fileName string) (out string, err error) {
// 	s := fmt.Sprintf("kubectl apply -f %v", fileName)
// 	cmd := exec.Command("sh", "-c", s)
// 	// cmd.Stdin = strings.NewReader(filebody)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		err = fmt.Errorf("apply file: %v err: %v, \ncmd: %v\noutput: %v", fileName, err, s, string(output))
// 		return
// 	}
// 	log.Printf("applied cmd: %v", s)
// 	out = string(output)
// 	return
// }

func ApplyByKubectlWithString(body string) (out string, err error) {
	s := fmt.Sprintf("kubectl apply -f -")
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = strings.NewReader(body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("apply body err: %v, \ncmd: %v\noutput: %v", err, s, string(output))
		return
	}
	log.Printf("kubectl apply: %v\n", body)
	out = string(output)
	return
}

func DeleteByKubectlWithString(body string) (out string, err error) {
	s := fmt.Sprintf("kubectl delete -f -")
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = strings.NewReader(body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("apply body err: %v, \ncmd: %v\noutput: %v", err, s, string(output))
		return
	}
	log.Printf("kubectl delete: %v\n", body)
	out = string(output)
	return
}

// var defaultHarborKey = flag.String("harborkey", "eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ==", "default HarborKey")

// make harbor key flag?
var nstmpl = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: %v
---
# harborkey
apiVersion: v1
data:
  .dockerconfigjson: %v
kind: Secret
metadata:
  name: devuser-harborkey
  namespace: %v
type: kubernetes.io/dockerconfigjson
`

// CheckOrCreateNamespace create ns if not exist.
func CheckOrCreateNamespace(ns string) (out string, err error) {
	s := fmt.Sprintf("kubectl get ns %v || kubectl apply -f -", ns)
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = strings.NewReader(getnsyaml(ns))
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("check or create ns: %v err: %v, \noutput: %v", ns, err, string(output))
		return
	}
	out = string(output)
	return
}

func getnsyaml(ns string) string {
	a := fmt.Sprintf(nstmpl, ns, defaultBase.harborkey, ns)
	// fmt.Println("yaml:", a)
	return a
}
