// buildsvc use function in this package
package project

import (
	"fmt"
	"os/exec"

	"github.com/pborman/ansi"
	// "github.com/acarl005/stripansi"
	"github.com/chinglinwen/log"
)

// to generate k8s template or just final yaml ( need to provided with ci env )
// dependent on values, often not include ci info?
// the generated filename is template, not final (to do that need change )
// this function just to provide an api for remote call ( or ui call )
func HelmGen(dir, project, env string) (out string, err error) {

	log.Printf("helmgen for project: %v, env: %v\n", project, env)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./gen.sh %v %v", project, env))
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("call gen.sh err: %v\noutput: %v\n", err, string(output))
		return
	}
	// out = stripansi.Strip(stripansi.Strip(string(output))) // let's strip twice for npm error color code
	b, err := ansi.Strip(output)
	out = string(b)
	return
}
