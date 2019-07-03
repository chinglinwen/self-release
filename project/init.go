package project

import (
	"flag"
	"wen/self-release/git"
)

var (
	defaultConfigRepo = flag.String("config-repo", "wenzhenglin/config-deploy", "default config-repo")
)

// var configrepo *git.Repo

// how to make config repo updated with remote?
// pull before use?
// func Init() {
// 	var once sync.Once
// 	onceBody := func() {
// 		log.Println("start init config-deploy repo")
// 		var err error
// 		// configrepo, err = git.New(*defaultConfigRepo, git.SetNoPull())
// 		configrepo, err = git.NewWithPull(*defaultConfigRepo, git.SetBranch("templateconfig")) //, git.SetNoPull())
// 		if err != nil {
// 			log.Println("new err:", err)
// 			return
// 		}
// 	}
// 	once.Do(onceBody)
// }

// let's pull for everytime it uses, so to keep update
func GetConfigRepo() (configrepo *git.Repo, err error) {
	return git.NewWithPull(*defaultConfigRepo, git.SetBranch("templateconfig")) //, git.SetNoPull())
}

// func init() {
// 	Init()
// }

// default is php.v1, we assume all is php?
// this can overwrite by release tag
func GetDefaultConfigVer() string {
	return "php.v1"
}

// should init at main
func InitBuildSVC(addr string) {
	defaultBuildsvc = NewBuildSVC(addr)
}
