// buildsvc use function in this package
package project

import (
	"fmt"
	"os"
	"os/exec"

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
	output, err := cmd.Output() // let stderr goes to err variable
	if err != nil {
		log.Printf("call gen.sh err: %v\noutput: %v\n", err, string(output))
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
