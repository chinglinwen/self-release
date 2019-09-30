package project

import (
	"fmt"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
)

// var (
// 	defaultConfigRepo = flag.String("config-repo", "wenzhenglin/config-deploy", "default config-repo")
// )

type base struct {
	harborkey    string
	buildsvcAddr string
	configRepo   string
}

var defaultBase *base

// pkg need init default base
func Setting(harborkey, addr, configrepo string) {
	defaultBase = &base{
		harborkey:    harborkey,
		buildsvcAddr: addr,
		configRepo:   configrepo,
	}
	defaultBuildsvc = NewBuildSVC(addr)
	log.Println("inited project base with:", defaultBase)
}
func (b *base) String() string {
	return fmt.Sprintf("\nharborkey: %v\nbuildsvcAddr:%v\nconfigRepo:%v\n", b.harborkey, b.buildsvcAddr, b.configRepo)
}

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
	if defaultBase == nil {
		err = fmt.Errorf("base not initialized")
		return
	}
	// return git.NewWithPull(defaultBase.configRepo, git.SetBranch("templateconfig")) //, git.SetNoPull())
	return git.NewWithPull(defaultBase.configRepo) //, git.SetNoPull())
}

// func init() {
// 	Init()
// }

// default is php.v1, we assume all is php?
// this can overwrite by release tag
func GetDefaultConfigVer() string {
	return "helm/phpv1"
}

func GetDefaultVer() string {
	return "v0.0.1"
}

// // should init at main
// func InitBuildSVC(addr string) {
// 	defaultBuildsvc = NewBuildSVC(addr)
// }
