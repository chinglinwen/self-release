package configstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func IsExist(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

func Copy(src, dst string) (err error) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read file: %v", err)
	}

	dir := filepath.Base(dst)
	os.MkdirAll(dir, os.ModeDir)

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("write file: %v, err: %v", dst, err)
	}
	return
}
