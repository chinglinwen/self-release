package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

	err = r.AddFile(filename)
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
	f := filepath.Join(r.Local, filename)
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		return true
	}
	return false
}
