package project

import (
	"fmt"
	"os/exec"

	"github.com/chinglinwen/log"
)

// build only build develop branch?
func (p *Project) Build(project, tag string) (out string, err error) {

	// clone first
	// if env is empty, it will set to master
	// repo, err := git.New(project, git.SetBranch(branch))
	// if err != nil {
	// 	log.Println("build newrepo err:", err)
	// 	return
	// }
	dir := p.repo.GetWorkDir()

	// f := filepath.Join(dir, "build.sh")
	// err = ioutil.WriteFile(f, []byte(buildBody), 0755)
	// if err != nil {
	// 	err = fmt.Errorf("writefile err: %v", err)
	// 	return
	// }

	// consider this? https://github.com/go-cmd/cmd

	return Build(dir, project, tag)
}

func Build(dir, project, tag string) (out string, err error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v", GetImage(project, tag)))
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("build execute build err:", err)
		return
	}
	out = string(output)
	return
}

func GetImage(project, tag string) string {
	return fmt.Sprintf("harbor.haodai.net/%v:%v", project, tag)
}