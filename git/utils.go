package git

import (
	"io/ioutil"
	"os"
	"path/filepath"

	git "gopkg.in/src-d/go-git.v4"
)

func (r *Repo) AddAndPush(filename, contents, commitText string, options ...func(*option)) (err error) {
	err = r.CheckoutLocal()
	if err != nil {
		return
	}

	err = r.Add(filename, contents, options...)
	if err != nil {
		return
	}
	err = r.Commit(commitText)
	if err != nil {
		return
	}
	return r.Push()
}

func (r *Repo) AddFileAndPush(filename, commitText string) (err error) {
	err = r.CheckoutLocal()
	if err != nil {
		return
	}

	err = r.GitAdd(filename)
	if err != nil {
		return
	}
	err = r.Commit(commitText)
	if err != nil {
		return
	}
	return r.Push()
}

func (r *Repo) GetFile(filename string) ([]byte, error) {
	f := filepath.Join(r.Local, filename)
	return ioutil.ReadFile(f)
}

func (r *Repo) IsExist(filename string) bool {
	if r == nil {
		return false
	}
	f := filepath.Join(r.Local, filename)
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		return true
	}
	return false
}

func (r *Repo) Status() (git.Status, error) {
	return r.wrk.Status()
}
