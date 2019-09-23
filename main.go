package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"wen/self-release/pkg/harbor"
	"wen/self-release/pkg/sse"

	rice "github.com/GeertJohan/go.rice"
	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	gitpkg "wen/self-release/git"
	projectpkg "wen/self-release/project"
)

var (
	port = flag.String("p", "8089", "port")

	defaultWebDir = flag.String("webdir", "web", "default web template dir")

	defaultConfigRepo = flag.String("configrepo", "wenzhenglin/config-deploy", "default config-repo")
	buildsvcAddr      = flag.String("buildsvc", "buildsvc:10000", "buildsvc address host:port ( or k8s service name )")
	defaultHarborKey  = flag.String("harborkey", "eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ==", "default HarborKey to pull image")

	harborURL  = flag.String("harbor-url", "http://harbor.haodai.net", "harbor URL for harbor auth")
	harborUser = flag.String("harbor-user", "", "harbor user for harbor auth")
	harborPass = flag.String("harbor-pass", "", "harbor pass for harbor auth")

	secretKey = flag.String("key", "", "secret key keep private")

	box *rice.Box

	// git
	defaultGitlabURL  = flag.String("gitlab-url", "http://g.haodai.net", "default gitlab url")
	defaultUser       = flag.String("gitlab-user", "", "default gitlab user")
	defaultPass       = flag.String("gitlab-pass", "", "default gitlab pass(personal token is ok)")
	gitlabAccessToken = flag.String("gitlab-token", "", "gitlab admin access token")
	defaultRepoDir    = flag.String("repoDir", "repos", "default path to store cloned projects")
)

func checkFlag() {
	fmt.Println("args:", os.Args)
	if *secretKey == "" {
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

	if *defaultUser == "" {
		log.Fatal("no defaultUser provided")
	}
	if *defaultPass == "" {
		log.Fatal("no defaultPass provided")
	}
	gitpkg.Init(*defaultGitlabURL, *defaultUser, *defaultPass, *gitlabAccessToken, *defaultRepoDir)
	log.Printf("using default notify user: %v", *defaultUser)
}

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()
	checkFlag()
	projectpkg.Setting(*defaultHarborKey, *buildsvcAddr, *defaultConfigRepo)
	harbor.Setting(*harborURL, *harborUser, *harborPass)

	box = rice.MustFindBox(*defaultWebDir)

	e := echo.New()

	// e.HTTPErrorHandler = customHTTPErrorHandler
	// e.Pre(middleware.AddTrailingSlash())

	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover()) // comments out for testing
	// e.Use(middleware.Logger())
	//e.Use(middleware.Static("/data"))

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	// echopprof.Wrap(e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	g := e.Group("/api")

	// g.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
	// 	log.Printf("url: %q, method: %v, body: %q\n", c.Request().URL, c.Request().Method, reqBody)
	// }))

	u := g.Group("/users")
	u.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}))
	u.Use(loginCheck())
	u.GET("/", getUserHandler)

	p := g.Group("/projects")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}))
	p.Use(loginCheck())

	// p.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
	// 	// if username == "joe" && password == "secret" {
	// 	// 	return true, nil
	// 	// }
	// 	// return false, nil

	// 	r := c.Request()
	// 	user := r.Header.Get("X-Auth-User")
	// 	log.Printf("got user: %v\n", user)
	// 	usertoken := r.Header.Get("X-Secret")
	// 	if user == "" || usertoken == "" {
	// 		return false, nil
	// 	}
	// 	return true, nil
	// }))

	p.Any("/:ns/:project", projectUpdateHandler)

	// get and put values files
	p.GET("/:ns/:project/values", projectValuesGetHandler)
	p.POST("/:ns/:project/values", projectValuesUpdateHandler)

	p.GET("/", projectListHandler)

	r := g.Group("/resources")
	r.GET("/:ns", projectResourceListHandler)

	// no where to handle auth?
	// g.GET("/init", initAPIHandler)
	// g.GET("/gen", genAPIHandler)
	// g.GET("/rollback", rollbackAPIHandler)

	g.GET("/wechat", wechatHandler)

	e.Any("/harbor", harborHandler)
	e.POST("/hook", hookHandler)

	// e.Static("/logs", "projectlogs")

	dosse(e)

	// e.File("/init", "init.html")
	// e.File("/gen", "gen.html")

	assetHandler := http.FileServer(box.HTTPBox())
	e.GET("/ui/*", echo.WrapHandler(http.StripPrefix("/ui/", assetHandler)))
	// e.GET("/", homeHandler)

	e.Logger.Fatal(e.Start(":" + *port))
	// err := e.Start(":" + *port)
	// log.Println("fatal", err)

	log.Println("exit")
}

func dosse(e *echo.Echo) {
	// b := sse.New()
	e.GET("/events", echo.WrapHandler(http.HandlerFunc(sse.SSEHandler)))
	// e.GET("/logs", echo.WrapHandler(http.HandlerFunc(sse.UIHandler)))

	// e.GET("/events/", homeHandler)
	e.GET("/logs", logsHandler)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	errorPage := fmt.Sprintf("%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
	c.Logger().Error("error here", err)
}
