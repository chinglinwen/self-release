package main

import (
	"fmt"
	"net/http"
	"wen/self-release/pkg/sse"

	rice "github.com/GeertJohan/go.rice"
	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	box *rice.Box
)

const (
	defaultDevBranch = "master"
)

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	InitAll()

	e := echo.New()

	// e.HTTPErrorHandler = customHTTPErrorHandler
	// e.Pre(middleware.AddTrailingSlash())

	// e.Use(middleware.Logger()) // too many harbor log
	// e.Use(middleware.Recover()) // comments out for testing

	// automatically add routers for net/http/pprof
	//  /debug/pprof, /debug/pprof/heap, etc.
	// echopprof.Wrap(e)

	g := e.Group("/api")
	g.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}))

	g.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		log.Printf("url: %q, method: %v, body: %q\n", c.Request().URL, c.Request().Method, reqBody)
	}))

	u := g.Group("/users")

	u.Use(loginCheck())
	u.GET("/", getUserHandler)

	g.Any("/gen/:ns/:project/:env", genYAMLHandler)
	g.Any("/apply/:ns/:project/:env", applyYAMLHandler)
	g.Any("/delete/:ns/:project/:env", deleteYAMLHandler)
	g.Any("/imagecheck/:ns/:project", projectImageCheckHandler)

	p := g.Group("/projects")
	p.Use(loginCheck())

	p.Any("/:ns/:project", projectUpdateHandler)

	// get and put values files
	p.GET("/:ns/:project/values", projectValuesGetHandler)
	p.POST("/:ns/:project/values", projectValuesUpdateHandler)

	p.GET("/:ns/:project/config", projectConfigGetHandler)
	p.POST("/:ns/:project/config", projectConfigUpdateHandler)

	p.GET("/", projectListHandler)

	r := g.Group("/resources")
	r.GET("/:ns", projectResourceListHandler)

	// deprecated
	// g.GET("/init", initAPIHandler)
	// g.GET("/gen", genAPIHandler)
	// g.GET("/rollback", rollbackAPIHandler)

	g.GET("/wechat", wechatHandler)
	e.Any("/harbor", harborHandler)
	e.POST("/hook", hookHandler)

	dosse(e)

	assetHandler := http.FileServer(box.HTTPBox())
	e.GET("/ui/*", echo.WrapHandler(http.StripPrefix("/ui/", assetHandler)))

	e.Logger.Fatal(e.Start(":" + port))
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
