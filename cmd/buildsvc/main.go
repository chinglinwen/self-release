package main

import (
	"flag"
	"fmt"
	"os"
	gitpkg "wen/self-release/git"
	"wen/self-release/pkg/harbor"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port = flag.String("p", "8005", "port")

	defaultConfigRepo = flag.String("configrepo", "yunwei/config-deploy", "default config-repo")
	buildsvcAddr      = flag.String("buildsvc", "buildsvc:10000", "buildsvc address host:port ( or k8s service name )")
	defaultHarborKey  = flag.String("harborkey", "eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ==", "default HarborKey to pull image")

	harborURL  = flag.String("harbor-url", "http://harbor.haodai.net", "harbor URL for harbor auth")
	harborUser = flag.String("harbor-user", "", "harbor user for harbor auth")
	harborPass = flag.String("harbor-pass", "", "harbor pass for harbor auth")

	// git
	defaultGitlabURL = flag.String("gitlab-url", "http://g.haodai.net", "default gitlab url")
	defaultUser      = flag.String("gitlab-user", "", "default gitlab user")
	defaultPass      = flag.String("gitlab-pass", "", "default gitlab pass(personal token is ok)")
	// gitlabAccessToken = flag.String("gitlab-token", "", "gitlab admin access token")
	defaultRepoDir = flag.String("repoDir", "repos", "default path to store cloned projects")
)

func checkFlag() {
	flag.Parse()
	fmt.Println("args:", os.Args)

	// git
	if *defaultGitlabURL == "" {
		log.Fatal("no defaultGitlabURL provided")
	}
	// if *gitlabAccessToken == "" {
	// 	log.Fatal("no gitlabAccessToken provided")
	// }
	if *defaultRepoDir == "" {
		log.Fatal("no defaultRepoDir provided")
	}

	if *defaultUser == "" {
		log.Fatal("no defaultUser provided")
	}
	if *defaultPass == "" {
		log.Fatal("no defaultPass provided")
	}
	gitpkg.Init(*defaultGitlabURL, *defaultUser, *defaultPass, "", *defaultRepoDir)
	log.Printf("using default notify user: %v", *defaultUser)
}

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	checkFlag()
	projectpkg.Setting(*defaultHarborKey, *buildsvcAddr, *defaultConfigRepo)
	harbor.Setting(*harborURL, *harborUser, *harborPass)

	e := echo.New()
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	// g := e.Group("/api")
	// g.GET("/build", buildAPIHandler)

	go func() {
		runGRPC()
	}()
	e.Logger.Fatal(e.Start(":" + *port))

	log.Println("exit")
}
