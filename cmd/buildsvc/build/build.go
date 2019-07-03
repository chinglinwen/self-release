// buildsvc use function in this package
package project

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/acarl005/stripansi"
	"github.com/chinglinwen/log"
)

func Build(dir, project, tag, env string) (out string, err error) {
	image := GetImage(project, tag)
	log.Printf("building for image: %v, env: %v\n", image, env)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v %v", image, env))
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("build execute build err: %v\noutput: %v\n", err, string(output))
		return
	}
	out = stripansi.Strip(string(output))
	return
}

func BuildStreamOutput(dir, project, tag, env string) (out chan string, err error) {
	out = make(chan string)
	defer close(out)

	image := GetImage(project, tag)
	log.Printf("building for image: %v, env: %v\n", image, env)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v %v", image, env))
	cmd.Dir = dir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	// scanner := bufio.NewScanner(stdout)
	// scanner.Split(bufio.ScanWords)

	go func() {
		out <- "start building image..."
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()
	go func() {
		cmd.Wait()
		close(out)
	}()
	log.Println("end of build cmd")
	return
}

func GetImage(project, tag string) string {
	return fmt.Sprintf("harbor.haodai.net/%v:%v", project, tag)
}
