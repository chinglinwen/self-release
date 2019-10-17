package project

import (
	"fmt"
	"strings"
	"wen/self-release/git"
	"wen/self-release/pkg/harbor"

	"github.com/k0kubun/pp"
)

// Build only build develop branch?
func (p *Project) Build(project, tag, env, commitid string) (out chan string, err error) {
	pp.Printf("try build for project: %v, tag: %v, env: %v, commitid: %v\n", project, tag, env, commitid)

	if git.BranchIsTag(tag) {
		_, err = git.CheckTagExist(project, tag)
		if err != nil {
			err = fmt.Errorf("check tag exist err: %v", err)
			return
		}
	}

	// clone first
	// if env is empty, it will set to master
	// repo, err := git.New(project, git.SetBranch(branch))
	// if err != nil {
	// 	log.Println("build newrepo err:", err)
	// 	return
	// }
	// dir := p.repo.GetWorkDir()

	// f := filepath.Join(dir, "build.sh")
	// err = ioutil.WriteFile(f, []byte(buildBody), 0755)
	// if err != nil {
	// 	err = fmt.Errorf("writefile err: %v", err)
	// 	return
	// }

	// consider this? https://github.com/go-cmd/cmd
	err = p.CreateHarborProjectIfNotExist()
	if err != nil {
		err = fmt.Errorf("try create harbor project err: %v", err)
		return
	}
	return Build(project, tag, env, commitid)
}

func (p *Project) CreateHarborProjectIfNotExist() (err error) {
	s := strings.Split(p.Project, "/")
	if len(s) == 0 {
		err = fmt.Errorf("project: %v, format invalid, should be group/repo", p.Project)
		return
	}
	return harbor.CreateProjectIfNotExist(s[0])
}

// func Build(dir, project, tag, env string) (out string, err error) {
// 	image := GetImage(project, tag)
// 	log.Printf("building for image: %v, env: %v\n", image, env)
// 	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v %v", image, env))
// 	cmd.Dir = dir
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		log.Printf("build execute build err: %v\noutput: %v\n", err, string(output))
// 		return
// 	}
// 	out = stripansi.Strip(string(output))
// 	return
// }

// func (p *Project) BuildStreamOutput(project, tag, env string) (out chan string, err error) {
// 	dir := p.repo.GetWorkDir()

// 	// f := filepath.Join(dir, "build.sh")
// 	// err = ioutil.WriteFile(f, []byte(buildBody), 0755)
// 	// if err != nil {
// 	// 	err = fmt.Errorf("writefile err: %v", err)
// 	// 	return
// 	// }

// 	// consider this? https://github.com/go-cmd/cmd
// 	out = make(chan string)
// 	err = BuildStreamOutput(dir, project, tag, env, out)
// 	return
// }

// func BuildStreamOutput(dir, project, tag, env string, out chan string) (err error) {
// 	image := GetImage(project, tag)
// 	log.Printf("building for image: %v, env: %v\n", image, env)
// 	cmd := exec.Command("sh", "-c", fmt.Sprintf("./build-docker.sh %v %v", image, env))
// 	cmd.Dir = dir
// 	// output, err := cmd.CombinedOutput()
// 	// if err != nil {
// 	// 	log.Printf("build execute build err: %v\noutput: %v\n", err, string(output))
// 	// 	return
// 	// }
// 	// out = string(output)

// 	stdout, _ := cmd.StdoutPipe()
// 	// stderr, _ := cmd.StderrPipe()
// 	cmd.Start()

// 	// out = make(chan<- string)
// 	// defer close(out)
// 	// defer stdout.Close()
// 	// defer stderr.Close()

// 	// scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
// 	scanner := bufio.NewScanner(stdout)
// 	// scanner.Split(bufio.ScanWords)

// 	go func() {
// 		out <- "start building image..."
// 		for scanner.Scan() {
// 			out <- scanner.Text()
// 		}
// 	}()
// 	go func() {
// 		cmd.Wait()
// 		close(out)
// 	}()
// 	log.Println("end of build cmd")
// 	return
// }

// if it's test, should generate unique id? so new apply will take effects?

// GetImage generate fixed image name and tag.
// share with build package ( code must be the same )
func GetImage(project, commitid string) (image string, err error) {
	if project == "" {
		err = fmt.Errorf("project is empty")
		return
	}
	if commitid == "" {
		err = fmt.Errorf("commitid is empty")
		return
	}
	image = fmt.Sprintf("harbor.haodai.net/%v:%v", project, commitid)
	return
}
