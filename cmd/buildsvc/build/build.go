// buildsvc use function in this package
package project

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	projectpkg "wen/self-release/project"

	"github.com/pborman/ansi"
	// "github.com/acarl005/stripansi"
	"github.com/chinglinwen/log"
)

const (
	BuildScriptName = "build-docker.sh"
)

// this function can use for testing purpose
func Build(dir, project, tag, env string) (out string, err error) {
	image, err := projectpkg.GetImage(project, tag)
	if err != nil {
		err = fmt.Errorf("getimage string err: %v", err)
		return
	}
	log.Printf("building for image: %v, env: %v\n", image, env)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./%v %v %v", BuildScriptName, image, env))
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("build execute build err: %v\noutput: %v\n", err, string(output))
		return
	}
	// out = stripansi.Strip(stripansi.Strip(string(output))) // let's strip twice for npm error color code
	b, err := ansi.Strip(output)
	out = string(b)
	return
}

func BuildStreamOutput(dir, project, tag, env, commitid string, out chan string) (err error) {
	// out = make(chan string, 100)
	// wg.Add(1)

	image, err := projectpkg.GetImage(project, commitid)
	if err != nil {
		err = fmt.Errorf("getimage string err: %v", err)
		return
	}
	log.Printf("building for image: %v, tag: %v, env: %v\n", image, tag, env)

	var cmd *exec.Cmd
	if isBuildScriptExist(dir) {
		log.Printf("buildscript exist, use it now\n")
		cmd = exec.Command("sh", "-c", fmt.Sprintf("./%v %v %v", BuildScriptName, image, env))
	} else {
		log.Printf("buildscript not exist, use default build scripts\n")
		log.Printf("using internal build script")
		cmd = exec.Command("sh", "-c", getDefaultBuildScript(image, env))
	}
	cmd.Dir = dir

	stdout, _ := cmd.StdoutPipe()
	// stderr, _ := cmd.StderrPipe()
	cmd.Start()

	// scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	scanner := bufio.NewScanner(stdout)
	// scanner.Split(bufio.ScanWords)

	go func() {
		out <- "start building image..."
		for scanner.Scan() {
			out <- scanner.Text()
		}
		log.Println("end of build output, wg.done")
		// wg.Done()
		close(out)
	}()
	go func() {
		cmd.Wait()
	}()
	log.Println("end of build cmd")
	return
}

// func GetImage(project, commitid string) (image string, err error) {
// 	if project == "" {
// 		err = fmt.Errorf("project is empty")
// 		return
// 	}
// 	if commitid == "" {
// 		err = fmt.Errorf("commitid is empty")
// 		return
// 	}
// 	image = fmt.Sprintf("harbor.haodai.net/%v:%v", project, commitid)
// 	return
// }

func isBuildScriptExist(dir string) bool {
	f := filepath.Join(dir, BuildScriptName)
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		return true
	}
	return false
}

func getDefaultBuildScript(image, env string) string {
	return fmt.Sprintf(defaultBuildScript, image, env)
}

// var defaultBuildScript = `
// #!/bin/sh

// image="%v"
// env="%v"

// echo "building $image, env: $env"

// if [ "$env" = "test" ]; then
//   cp -f .env.test .env
// else
//   cp -f .env.online .env
// fi

// echo docker build --pull -t $image .
// echo docker push $image
// `

var defaultBuildScript = `
#!/bin/sh
# build-docker.sh

image="%v"
env="%v"

echo "building $image, env: $env"

if [ "$env" = "test" ]; then
  cp -f .env.test .env
else
  cp -f .env.online .env
fi

docker build --pull -t $image .
docker push $image
`
