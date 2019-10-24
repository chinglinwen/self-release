package main

import (
	flagpkg "flag"
	"fmt"
	"os"
	"path/filepath"
	gitpkg "wen/self-release/git"
	"wen/self-release/pkg/harbor"
	"wen/self-release/pkg/k8s"
	"wen/self-release/pkg/notify"
	"wen/self-release/pkg/sse"
	projectpkg "wen/self-release/project"

	rice "github.com/GeertJohan/go.rice"
	"github.com/chinglinwen/log"

	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/ffyaml"
)

var (
	port      string
	secretKey string
	selfURL   string
)

const harborkey = "eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ=="

func InitAll() {
	flag := flagpkg.NewFlagSet("self-release", flagpkg.ExitOnError)
	var (
		flagport      = flag.String("p", "8089", "port")
		flagsecretKey = flag.String("key", "", "secret key keep private")
		flagselfURL   = flag.String("self-url", "http://release.haodai.net", "self URL for log view")

		defaultWebDir = flag.String("webdir", "web", "default web template dir")

		defaultConfigRepo = flag.String("configrepo", "yunwei/config-deploy", "default config-repo")
		buildsvcAddr      = flag.String("buildsvc", "buildsvc:10000", "buildsvc address host:port ( or k8s service name )")
		defaultHarborKey  = flag.String("harborkey", harborkey, "default HarborKey to pull image")

		harborURL  = flag.String("harbor-url", "http://harbor.haodai.net", "harbor URL for harbor auth")
		harborUser = flag.String("harbor-user", "", "harbor user for harbor auth")
		harborPass = flag.String("harbor-pass", "", "harbor pass for harbor auth")

		// git
		defaultGitlabURL  = flag.String("gitlab-url", "http://g.haodai.net", "default gitlab url")
		defaultUser       = flag.String("gitlab-user", "", "default gitlab user")
		defaultPass       = flag.String("gitlab-pass", "", "default gitlab pass(personal token is ok)")
		gitlabAccessToken = flag.String("gitlab-token", "", "gitlab admin access token")
		defaultRepoDir    = flag.String("repoDir", "repos", "default path to store cloned projects")

		logsPath = flag.String("logsDir", "projectlogs", "build logs dir")

		wechatURL = flag.String("wechat-receiver-url", "http://localhost:8002", "wechat-receiver-url")

		kubeconfig = flag.String("kubeconfig", defaultKubeConfig(), "path to the kubeconfig file (optional)")

		_ = flag.String("config", "", "config file (optional)")
	)

	ff.Parse(flag, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ffyaml.Parser),
		ff.WithEnvVarPrefix("SR"),
	)

	// set for main
	port = *flagport
	secretKey = *flagsecretKey
	selfURL = *flagselfURL

	fmt.Println("args:", os.Args)
	if secretKey == "" {
		log.Fatal("secretKey is empty")
	}

	// git
	if *defaultGitlabURL == "" {
		log.Fatal("no defaultGitlabURL provided")
	}
	if *gitlabAccessToken == "" {
		log.Fatal("no gitlabAccessToken provided")
	}
	if *defaultRepoDir == "" {
		log.Fatal("no defaultRepoDir provided")
	}

	if *harborUser == "" {
		log.Fatal("no harborUser provided")
	}
	if *harborPass == "" {
		log.Fatal("no harborPass provided")
	}

	if *defaultUser == "" {
		log.Fatal("no defaultUser provided")
	}
	if *defaultPass == "" {
		log.Fatal("no defaultPass provided")
	}
	gitpkg.Init(*defaultGitlabURL, *defaultUser, *defaultPass, *gitlabAccessToken, *defaultRepoDir)
	log.Printf("using default notify user: %v", *defaultUser)

	sse.Init(*logsPath)
	projectpkg.Setting(*defaultHarborKey, *buildsvcAddr, *defaultConfigRepo)
	harbor.Setting(*harborURL, *harborUser, *harborPass)

	notify.Init(*wechatURL)
	k8s.Init(*kubeconfig)

	box = rice.MustFindBox(*defaultWebDir)
}

func defaultKubeConfig() string {
	return filepath.Join(homeDir(), ".kube", "config")
}
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
