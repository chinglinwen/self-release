package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"
	"fmt"
	"wen/self-release/git"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port = flag.String("p", "8005", "port")
)

func init() {
	fmt.Println("start test init setting")
	git.Init("wenzhenglin", "cKGa3eVAF7tZMvCukdsP")
}

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	//e.Use(middleware.Static("/data"))

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	// echopprof.Wrap(e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	g := e.Group("/api")
	g.GET("/build", buildAPIHandler)

	// assetHandler := http.FileServer(rice.MustFindBox("web").HTTPBox())
	// e.GET("/ui/*", echo.WrapHandler(http.StripPrefix("/ui/", assetHandler)))
	// e.GET("/", homeHandler)

	// grpcServer := grpc.NewServer()
	// pb.RegisterBuildsvcServer(grpcServer, newServer())
	// srv := &http.Server{
	// 	Addr:    ":" + *port,
	// 	Handler: grpcHandlerFunc(grpcServer, e),
	// 	// TLSConfig: &tls.Config{
	// 	// 	Certificates: []tls.Certificate{*demoKeyPair},
	// 	// 	NextProtos:   []string{"h2"},
	// 	// },
	// }
	// e.Logger.Fatal(e.StartServer(srv))

	go func() {
		runGRPC()
	}()
	e.Logger.Fatal(e.Start(":" + *port))

	log.Println("exit")
}
