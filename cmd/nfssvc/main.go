package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port          = flag.String("p", "8080", "port")
	nfsServerName = flag.String("server", "172.31.83.26", "server name info")
	exportFile    = flag.String("path", "/etc/exports", "exports file path")
)

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()

	nfs := newNfs(*nfsServerName, *exportFile)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}))
	// e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	//e.Use(middleware.Static("/data"))

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	// echopprof.Wrap(e)

	g := e.Group("/api")
	g.GET("/", nfs.listNfsHandler)

	e.Logger.Fatal(e.Start(":" + *port))

	log.Println("exit")
}
