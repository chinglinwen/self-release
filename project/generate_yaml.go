// buildsvc use function in this package
package project

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	// "github.com/acarl005/stripansi"
	"github.com/chinglinwen/log"
)

const (
	GENSH        = "./gen.sh"
	GENSHAPICALL = "call"
	GENSHPRINT   = "print"
)

// to generate k8s template or just final yaml ( need to provided with ci env )
// dependent on values, often not include ci info?
// the generated filename is template, not final (to do that need change )
// this function just to provide an api for remote call ( or ui call )
func HelmGen(project, env string) (out string, err error) {
	return dogen(project, env, GENSHAPICALL)
}

// HelmGenPrint print only
func HelmGenPrint(project, env string) (out string, err error) {
	return dogen(project, env, GENSHPRINT)
}

// validate with example
func HelmGenPrintFinal(project, env string, envMap map[string]string) (final string, err error) {
	_, final, err = dogenFinal(project, env, envMap)
	if err != nil {
		return
	}
	_, err = ValidateByKubectlWithString(final)
	if err != nil {
		err = fmt.Errorf("validate yaml err: %v", err)
		return
	}
	return
}

var exampleEnvMap = getexampleEnvMap()

func HelmGenPrintValidateYaml(project, env string) (out string, err error) {
	exampleEnvMap["CI_TIME"] = time.Now().Format(TimeLayout)
	out, final, err := dogenFinal(project, env, exampleEnvMap)
	if err != nil {
		return
	}
	o, err := ValidateByKubectlWithString(final)
	if err != nil {
		err = fmt.Errorf("validate yaml err: %v\noutput: %v\nfinal: %v", err, o, final)
		return
	}
	return
}

const TimeLayout = "2006-1-2_15:04:05"

func getexampleEnvMap() (autoenv map[string]string) {
	autoenv = make(map[string]string)
	autoenv["CI_PROJECT_PATH"] = "demo/hello"
	autoenv["CI_BRANCH"] = "v1.0.0"
	autoenv["CI_ENV"] = "online"
	autoenv["CI_NAMESPACE"] = "demo"
	autoenv["CI_PROJECT_NAME"] = "hello"
	autoenv["CI_PROJECT_NAME_WITH_ENV"] = "hello" + "-" + "online"
	autoenv["CI_REPLICAS"] = "1"
	autoenv["CI_IMAGE"] = "example.com/demo/hello:v1.0.0"
	autoenv["CI_USER_NAME"] = "demouser"
	autoenv["CI_USER_EMAIL"] = "demouser@example.com"
	autoenv["CI_MSG"] = "demo info to validate yaml"
	// autoenv["CI_TIME"] = time.Now().Format(TimeLayout)
	return
}

func dogenFinal(project, env string, envMap map[string]string) (out, final string, err error) {
	out, err = dogen(project, env, GENSHPRINT)
	if err != nil {
		err = fmt.Errorf("get yaml err: %v", err)
		return
	}
	final, err = generateByMap(out, envMap)
	if err != nil {
		err = fmt.Errorf("generate final with map err: %v", err)
		return
	}
	return
}

func dogen(project, env, apicontext string) (out string, err error) {
	// skip project fetch, we fetch config repo directly
	configrepo, err := GetConfigRepo()
	if err != nil {
		err = fmt.Errorf("get config repo err: %v", err)
		return
	}
	dir := configrepo.GetWorkDir()

	out, err = runHelmGen(dir, project, env, apicontext)
	if err != nil {
		err = fmt.Errorf("runHelmGen err: %v\n", err)
		return
	}
	if apicontext == GENSHAPICALL {
		// compare with old to see if need to commit?
		// there's timestamp, there will be different for commit
		commit := fmt.Sprintf("generate yaml for project: %v, env: %v", project, env)
		err = configrepo.PushLocalChange(commit)
	}
	return
}

func runHelmGen(dir, project, env, apicontext string) (out string, err error) {
	log.Printf("helmgen for project: %v, env: %v, workdir: %v\n", project, env, dir)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%v %v %v", GENSH, project, env))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "APICONTEXT="+apicontext)
	cmd.Dir = dir

	// depend on helm gen.sh to return error
	// output, err := cmd.Output() // let stderr goes to err variable
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("call gen.sh err: %v\noutput: %v\n", err, string(output))
		return
	}
	out = string(output)
	return
}

// for local test, no git pull
// func GetConfigRepo1() (configrepo *git.Repo, err error) {
// 	if defaultBase == nil {
// 		err = fmt.Errorf("base not initialized")
// 		return
// 	}
// 	return git.New(defaultBase.configRepo)
// }
