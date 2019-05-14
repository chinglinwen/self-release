// +build !windows

// Example of turn debug on the fly
//      $ kill -s SIGUSR1 prog
package main

func init() {
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, syscall.SIGUSR1) //SIGUSR1 doesn't work on windows
	// go func() {
	// 	for _ = range c {
	// 		var level string
	// 		switch log.GetLevel() {
	// 		case "debug":
	// 			level = "info"
	// 			gopsStop()
	// 			// stopdebug()
	// 			log.Println("stopping gops")
	// 		default:
	// 			level = "debug"
	// 			gopsStart()
	// 			// setdebug()
	// 			log.Println("started gops")
	// 		}
	// 		log.SetLevel(level)
	// 		log.Println("got signal, set log level to ", level)
	// 	}
	// }()
}
