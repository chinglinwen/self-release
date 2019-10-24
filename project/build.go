package project

import (
	"fmt"
	"strings"
	"wen/self-release/git"
	"wen/self-release/pkg/harbor"

	"github.com/k0kubun/pp"
)

// Build only build develop branch?
func (p *Project) Build(env, commitid string) (err error) {
	project, tag := p.Project, p.Branch
	pp.Printf("try build for project: %v, tag: %v, env: %v, commitid: %v\n", project, tag, env, commitid)

	if git.BranchIsTag(tag) {
		_, err = git.CheckTagExist(project, tag)
		if err != nil {
			err = fmt.Errorf("check tag exist err: %v", err)
			return
		}
	}

	// consider this? https://github.com/go-cmd/cmd
	err = p.CreateHarborProjectIfNotExist()
	if err != nil {
		err = fmt.Errorf("try create harbor project err: %v", err)
		return
	}
	p.buildsvc = Build(project, tag, env, commitid)
	if p.buildsvc.err != nil {
		// for early error
		err = p.buildsvc.err
	}
	return
}

func (p *Project) CreateHarborProjectIfNotExist() (err error) {
	s := strings.Split(p.Project, "/")
	if len(s) == 0 {
		err = fmt.Errorf("project: %v, format invalid, should be group/repo", p.Project)
		return
	}
	return harbor.CreateProjectIfNotExist(s[0])
}

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
