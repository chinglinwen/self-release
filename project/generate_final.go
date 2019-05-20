//
package project

import (
	"fmt"
	"log"
	"os/exec"
	"wen/self-release/git"
)

// https://medium.com/@thedevsaddam/easy-way-to-render-html-in-go-34575f858026

// generate from repotemplate or configrepo to final

// generate final
func GenerateFinal(project, branch string) (out string, err error) {
	// clone first
	// if env is empty, it will set to master
	repo, err := git.New(project, git.SetBranch(branch))
	if err != nil {
		log.Println("build newrepo err:", err)
		return
	}
	dir := repo.GetWorkDir()

	// f := filepath.Join(dir, "build.sh")
	// err = ioutil.WriteFile(f, []byte(buildBody), 0755)
	// if err != nil {
	// 	err = fmt.Errorf("writefile err: %v", err)
	// 	return
	// }

	// cosider this? https://github.com/go-cmd/cmd
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v", branch))
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("build execute build err:", err)
		return
	}

	// fmt.Printf("out: %vs\n", out)
	out = string(output)
	// using dockerfile(provided) to build
	return
}
