package project

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/chinglinwen/log"
)

func Apply(project, env string, envMap map[string]string) (out string, err error) {
	log.Debug.Printf("try kubectl apply for project: %v, env: %v", project, env)
	final, err := HelmGenPrintFinal(project, env, envMap)
	if err != nil {
		return
	}
	return ApplyByKubectlWithString(final)
}

func Delete(project, env string, envMap map[string]string) (out string, err error) {
	log.Debug.Printf("try kubectl delete for project: %v, env: %v", project, env)
	final, err := HelmGenPrintFinal(project, env, envMap)
	if err != nil {
		return
	}
	return DeleteByKubectlWithString(final)
}

func ApplyByKubectlWithString(body string) (out string, err error) {
	// return // TODO: disbale it

	s := fmt.Sprintf("kubectl apply -f -")
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = strings.NewReader(body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("apply body err: %v, \ncmd: %v\noutput: %v", err, s, string(output))
		log.Printf("kubectl apply yaml: \n%v\n", body)
		return
	}
	// log.Printf("kubectl apply: %v\n", body)
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
	// log.Printf("kubectl deleted\n")
	out = string(output)
	return
}

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
	return fmt.Sprintf(nstmpl, ns, defaultBase.harborkey, ns)
}
