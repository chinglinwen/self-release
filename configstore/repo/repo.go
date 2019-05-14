package repo

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// var (
// 	defaultRepoDir = "/tmp/repos"
// )

type RepoBase struct {
	Base string

	perm os.FileMode
}

func New(base string) *RepoBase {
	return &RepoBase{
		Base: base,
		perm: 0644,
	}
}

// relate to config store
func (b *RepoBase) GetFile(file string) ([]byte, error) {
	f := filepath.Join(b.Base, file)
	return ioutil.ReadFile(f)
}

func SetPerm(perm os.FileMode) func(*RepoBase) {
	return func(b *RepoBase) {
		b.perm = perm
	}
}

func (b *RepoBase) WriteFile(file string, data []byte, options ...func(*RepoBase)) error {
	for _, op := range options {
		op(b)
	}

	f := filepath.Join(b.Base, file)
	dir := filepath.Base(f)
	os.MkdirAll(dir, os.ModeDir)

	return ioutil.WriteFile(f, data, b.perm)
}
