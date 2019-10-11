// Example of turn debug on the fly
//      $ kill -s SIGUSR1 pid
// or   $ kill -10 pid
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chinglinwen/log"
)

func init() {
	if os.Getenv("GODEBUG") != "" {
		log.SetLevel("debug")
		log.Println("got debug env, set log level to ", "debug")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1) //SIGUSR1 doesn't work on windows
	go func() {
		for _ = range c {
			var level string
			switch log.GetLevel() {
			case "debug":
				level = "info"
			default:
				level = "debug"
			}
			log.SetLevel(level)
			log.Println("got signal, set log level to ", level)
		}
	}()
}
