// buildsvc use function in this package
package project

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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

type Builder struct {
	Project  string
	Dir      string
	Env      string
	Tag      string
	Commitid string

	image string
	// out   chan string

	cmd *exec.Cmd
	out *bufio.Scanner
	err error
}

func NewBuilder(dir, project, tag, env, commitid string) (b *Builder) {
	image, err := projectpkg.GetImage(project, commitid)
	if err != nil {
		err = fmt.Errorf("getimage string err: %v", err)
		log.Printf("new builder err: %v", err)
	}
	log.Printf("building for image: %v, tag: %v, env: %v\n", image, tag, env)
	b = &Builder{
		Project:  project,
		Dir:      dir,
		Env:      env,
		Tag:      tag,
		Commitid: commitid,
		image:    image,
		err:      err,
	}
	b.BuildStreamOutput()
	return
}

func (b *Builder) GetError() (err error) {
	return b.err
}

func (b *Builder) Output() (s *bufio.Scanner, err error) {
	if b.err != nil {
		err = b.err
		log.Printf("builder err: %v", b.err)
		return
	}
	i := 1
	for b.cmd == nil {
		log.Printf("waiting cmd... %v times\n", i)
		if i > 10 {
			err = fmt.Errorf("times out of waiting cmd created err")
			log.Print(err)
			return
		}
		time.Sleep(time.Duration(i * 2 * int(time.Second)))
		i++
	}
	log.Printf("cmd is ready\n")
	if b.cmd == nil {
		err = fmt.Errorf("cmd is nil, should not happen")
		return
	}
	s = b.out
	return
}

func (b *Builder) BuildStreamOutput() {
	if b.err != nil {
		log.Printf("builder err: %v", b.err)
		return
	}
	log.Printf("creating cmd...\n")

	if isBuildScriptExist(b.Dir) {
		log.Printf("buildscript exist, use it now\n")
		b.cmd = exec.Command("sh", "-c", fmt.Sprintf("sh ./%v %v %v", BuildScriptName, b.image, b.Env))
	} else {
		log.Printf("buildscript not exist, use default build scripts\n")
		log.Printf("using internal build script")
		b.cmd = exec.Command("sh", "-c", getDefaultBuildScript(b.image, b.Env))
	}
	b.cmd.Dir = b.Dir

	stdout, err := b.cmd.StdoutPipe()
	if err != nil {
		b.err = err
		return
	}
	stderr, err := b.cmd.StderrPipe()
	if err != nil {
		b.err = err
		return
	}
	b.out = bufio.NewScanner(io.MultiReader(stdout, stderr))

	b.cmd.Start()

	go func() {
		err := b.cmd.Wait()
		if err != nil {
			b.err = err
			log.Printf("build run exit with err: %v\n", err)
		} else {
			log.Println("build run exit with ok")
		}
		log.Println("end of build cmd")
	}()
	return
}

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

var defaultBuildScript = `
#!/bin/sh
# build-docker.sh
set -e

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
