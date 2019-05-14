package configstore

import (
	"wen/self-release/git"
)

var repo *git.Repo

// func Init() {
// 	log.Println("start init config-deploy repo")
// 	var once sync.Once
// 	onceBody := func() {
// 		var err error
// 		repo, err = git.New("yunwei/config-deploy")
// 		if err != nil {
// 			log.Println("new err:", err)
// 			return
// 		}
// 	}
// 	once.Do(onceBody)
// }
