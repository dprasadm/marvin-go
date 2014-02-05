package main

import (
    "runtime"
    "marvin/cloud/request"
    "marvin/web/server"
    //"log"
)

func main() {
    for {
        runtime.GOMAXPROCS(4)
        
        defer func() {
		    //log.Println("Running Marvin Panic Handler")
		    if x := recover(); x != nil {
			    //log.Printf("run time panic: %v", x)
		    }
	    }()

        cloud.InitRequestHandlers()
        server.Run()
    }
}

