package template

import (
	"log"
	"sync"
	"wen/self-release/git"
)

var configrepo *git.Repo

func Init() {
	var once sync.Once
	onceBody := func() {
		log.Println("start init config-deploy repo")
		var err error
		configrepo, err = git.New("yunwei/config-deploy", git.SetNoPull())
		if err != nil {
			log.Println("new err:", err)
			return
		}
	}
	once.Do(onceBody)
}

func GetConfigRepo() *git.Repo {
	if configrepo == nil {
		Init()
	}
	return configrepo
}

func init() {
	Init()
}

// default is php.v1, we assume all is php?
// this can overwrite by release tag
func GetDefaultConfigVer() string {
	return "php.v1"
}
