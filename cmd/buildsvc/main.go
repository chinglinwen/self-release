package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"
	"wen/self-release/pkg/harbor"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port = flag.String("p", "8005", "port")

	defaultConfigRepo = flag.String("configrepo", "wenzhenglin/config-deploy", "default config-repo")
	buildsvcAddr      = flag.String("buildsvc", "buildsvc:10000", "buildsvc address host:port ( or k8s service name )")
	defaultHarborKey  = flag.String("harborkey", "eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ==", "default HarborKey to pull image")

	harborURL  = flag.String("harbor-url", "http://harbor.haodai.net", "harbor URL for harbor auth")
	harborUser = flag.String("harbor-user", "", "harbor user for harbor auth")
	harborPass = flag.String("harbor-pass", "", "harbor pass for harbor auth")
)

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()

	projectpkg.Setting(*defaultHarborKey, *buildsvcAddr, *defaultConfigRepo)
	harbor.Setting(*harborURL, *harborUser, *harborPass)

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
