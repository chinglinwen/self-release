package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

var (
	defaultGitUserName  = "robot"
	defaultGitUserEmail = "robot@example.com"
)

// type Repo struct {
// 	R *git.Repository
// }

// example refs: "refs/remotes/origin/feature1"
// default to repo's branch
// func (r *Repo) CheckoutRemote(refs string) (err error) {
// 	if refs == "" {
// 		refs = r.refs
// 	}
// 	err = r.wrk.Checkout(&git.CheckoutOptions{
// 		Branch: plumbing.ReferenceName(refs),
// 		Force:  true,
// 		// Create: true,
// 	})
// 	if err != nil && err != git.ErrBranchExists {
// 		return fmt.Errorf("git checkout refs: %v err %v", refs, err)
// 	}

// 	return
// }

func (r *Repo) LocalBranchExist() (ok bool) {
	refss, err := r.R.References()
	refss.ForEach(func(item *plumbing.Reference) error {
		if item.Name().String() == r.localrefs {
			ok = true
		}
		return nil
	})
	if err != nil {
		log.Println("localbranchexist check err", err)
	}
	return
}

func (r *Repo) CheckoutLocal() (err error) {
	return r.CheckoutLocalWith("")
}

func (r *Repo) CheckoutLocalWith(refs string) (err error) {
	if refs == "" {
		refs = r.refs
	}
	// fmt.Println("refs", refs)
	err = r.wrk.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(refs),
		Force:  r.force,
	})
	if err != nil && err != git.ErrBranchExists {
		err = fmt.Errorf("git checkout refs: %v err %v", refs, err)
		log.Println(err)
		return
	}

	ok := r.LocalBranchExist()
	err = r.wrk.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(r.localrefs),
		Force:  r.force, // need force by default? let up layer to decide
		Create: !ok,
	})
	if err != nil && err != git.ErrBranchExists {
		err = fmt.Errorf("git checkout branch: %v err %v", r.localrefs, err)
		log.Println(err)
		return
	}
	return
}

type option struct {
	perm os.FileMode
}

func SetPerm(perm os.FileMode) func(*option) {
	return func(o *option) {
		o.perm = perm
	}
}

func (r *Repo) Add(filename, contents string, options ...func(*option)) (err error) {
	err = r.Create(filename, contents, options...)
	if err != nil {
		return
	}
	return r.GitAdd(filename)
}

// create file and add to git
func (r *Repo) Create(filename, contents string, options ...func(*option)) (err error) {
	o := &option{perm: 0755} // default filemode
	for _, op := range options {
		op(o)
	}

	f := filepath.Join(r.Local, filename)
	dir := filepath.Dir(f)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("git create file mkdir err %v", err)
	}
	// log.Printf("gitadd, writing file: %v", f)
	err = ioutil.WriteFile(f, []byte(contents), o.perm)
	if err != nil {
		return fmt.Errorf("git create file err %v", err)
	}
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return fmt.Errorf("git create file check err %v", err)
	}
	// _, err = r.wrk.Add(filename)
	return
}

// file is exist
func (r *Repo) GitAdd(filename string) (err error) {
	// log.Printf("gitadd file: %v\n", filename)
	_, err = r.wrk.Add(filename)
	return
}

func (r *Repo) CommitAndPush(commitText string) (err error) {
	err = r.Commit(commitText)
	if err != nil {
		return
	}
	return r.Push()
}

func (r *Repo) Commit(commitText string) (err error) {

	status, err := r.wrk.Status()

	if status == nil || len(status) == 0 {
		// log.Printf("attempt to commit, but nothing changed in %v:%v\n", r.Project, r.Branch)
		err = fmt.Errorf("attempt to commit, but nothing changed in %v:%v\n", r.Project, r.Branch)
		return
	}
	log.Println("get status ok", r.Project)
	commit, err := r.wrk.Commit(commitText, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  defaultGitUserName,
			Email: defaultGitUserEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("create commit err: %v", err)
	}
	log.Println("create commit ok", r.Project)

	_, err = r.R.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("do commit err: %v", err)
	}
	log.Println("do commit ok", r.Project)
	return
}

func (r *Repo) Push() (err error) {
	// status, err := r.Status()
	// if err != nil {
	// 	return fmt.Errorf("git status err: %v", err)
	// }
	// after commit, it's clean
	// if status.IsClean() {
	// 	return fmt.Errorf("it's clean, no need push")
	// }

	err = r.R.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: r.user,
			Password: r.pass,
		},

		// RefSpecs: []config.RefSpec{config.RefSpec(r.realtag)},
		// RefSpecs: []config.RefSpec{config.RefSpec("+refs/tags/v1.0.0:refs/tags/v1.0.0")},
		// RefSpecs: []config.RefSpec{config.RefSpec("+refs/tags/v1.0.0:refs/remotes/origin/master")},
		// RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("+refs/heads:%v", r.refs))}, // +refs/heads/*:refs/remotes/origin/*
		// RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("+%v:%v", r.refs, r.refs))},

		// RefSpecs: []config.RefSpec{config.RefSpec("+refs/heads/feature1:refs/remotes/origin/feature1")},
	})

	if err == git.NoErrAlreadyUpToDate {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("push commit err: %v", err)
	}
	log.Println("done pushed for ", r.Project)
	return
}
