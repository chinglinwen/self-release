package template

import (
	"flag"
	"log"
	"sync"
	"wen/self-release/git"
)

var (
	defaultConfigRepo = flag.String("config-repo", "wenzhenglin/config-deploy", "default config-repo")
)

var configrepo *git.Repo

func Init() {
	var once sync.Once
	onceBody := func() {
		log.Println("start init config-deploy repo")
		var err error
		// configrepo, err = git.New(*defaultConfigRepo, git.SetNoPull())
		configrepo, err = git.NewWithPull(*defaultConfigRepo, git.SetBranch("templateconfig")) //, git.SetNoPull())
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
