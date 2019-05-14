package template

import (
	"log"
	"sync"
	"wen/self-release/git"
)

var configrepo *git.Repo

func Init() {
	log.Println("start init config-deploy repo")
	var once sync.Once
	onceBody := func() {
		var err error
		configrepo, err = git.New("yunwei/config-deploy")
		if err != nil {
			log.Println("new err:", err)
			return
		}
	}
	once.Do(onceBody)
}
