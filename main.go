package main

// listening on hooks

// fetch config-deploy

// trigger projects behavior

import (
	"flag"
	"fmt"
	"wen/self-release/git"
	"wen/self-release/template"

	"github.com/chinglinwen/log"
	"github.com/docker/docker/daemon/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	conf             *config.Config
	env              = flag.String("env", "qa", "env includes (qa,pre,pro)")
	port             = flag.String("p", "8089", "port")
	checkInterval    = flag.String("i", "10s", "check interval (s second,m minutes)")
	concurrentChecks = flag.Int("cc", 100, "number of concurrent checks")
	testproject      = flag.String("test", "", "test project name")
	checkonetime     = flag.Bool("once", false, "check only once")
	dockerOnly       = flag.Bool("docker", true, "check docker only")

	// see wechat-notify, example value: usera|userb  ( the email prefix )
	defaultReceiver = flag.String("default-receiver", "wenzhenglin", "default receivers, using the email prefix, example: usera|userb")

	alertAll = flag.Bool("alertall", true, "alert all changes to default receiver if setted")

	upstreamBase = flag.String("upstream", "http://upstream-test.sched.qianbao-inc.com:8010", "upstream base api url")
)

// // try have two config
// // one for fetch and one for manual editing
// func init() {
// 	flag.Parse()
// 	conf = config.New("config.json", *env) //try the item, project based ?  why not just name?

// 	_ = os.Mkdir("data", 0775)
// 	conf.Notifier = checkup.Qianbao{
// 		Username: "wen",
// 		Channel:  "http://localhost:" + *port + "/notify",
// 	}
// 	conf.Storage = checkup.FS{
// 		Dir:         "data",
// 		CheckExpiry: 7 * 24 * time.Hour,
// 	}
// 	conf.ConcurrentChecks = *concurrentChecks
// 	conf.Save()

// 	var err error
// 	_, err = time.ParseDuration(*checkInterval)
// 	if err != nil {
// 		log.Fatalf("parse checkInterval duration error for %v", *checkInterval)
// 	}
// 	check.CheckInterval = *checkInterval

// 	if *testproject != "" {
// 		log.Printf("test for %v project only\n", *testproject)
// 		check.TestProject = *testproject
// 	}
// 	if *checkonetime {
// 		log.Println("check one time only")
// 		check.CheckOneTime = *checkonetime
// 	}
// 	check.Env = *env
// 	check.DockerOnly = *dockerOnly

// 	check.Init(*upstreamBase)
// 	upstream.Init(*upstreamBase)
// 	log.Println("using upstream", *upstreamBase)

// 	checkup.AlertAll = *alertAll
// 	checkup.DefaultReceiver = *defaultReceiver
// }

// define a global variable
// add new check, update it, and store the config as file(update config)

func init() {
	fmt.Println("start test init setting")
	git.Init("wenzhenglin", "cKGa3eVAF7tZMvCukdsP")

	template.Init()
}

func main() {
	log.Println("starting...")
	log.Debug.Println("debug is on")

	flag.Parse()

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.Static("/data"))

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	// echopprof.Wrap(e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	g := e.Group("/api")
	g.GET("/init", initAPIHandler)
	g.GET("/gen", genAPIHandler)

	// e.Static("/", "statuspage")
	// e.Static("/data", "data")

	e.File("/init", "init.html")
	e.File("/gen", "gen.html")

	e.GET("/", homeHandler)
	e.POST("/hook", hookHandler)

	err := e.Start(":" + *port)
	log.Println("fatal", err)
	//e.Logger.Fatal()

	log.Println("exit")
}
