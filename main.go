package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"
	"fmt"
	"net/http"
	"wen/self-release/pkg/harbor"
	"wen/self-release/pkg/sse"

	rice "github.com/GeertJohan/go.rice"
	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

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

	box *rice.Box
)

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()
	projectpkg.Setting(*defaultHarborKey, *buildsvcAddr, *defaultConfigRepo)
	harbor.Setting(*harborURL, *harborUser, *harborPass)

	box = rice.MustFindBox(*defaultWebDir)

	e := echo.New()

	// e.HTTPErrorHandler = customHTTPErrorHandler
	// e.Pre(middleware.AddTrailingSlash())

	e.Use(middleware.Logger())
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

	p := g.Group("/projects")

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
