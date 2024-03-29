package main

import (
	"git.blob42.xyz/blob42/hugobot/v3/db"
	"log"
	"os"
	"syscall"
	"time"

	gum "git.blob42.xyz/blob42/gum.git"
)

var (
	DB   = db.DB
	quit chan bool
)

func shutdown(c <-chan os.Signal) {
	ticker := time.NewTicker(JobSchedulerInterval)

	for {
		select {
		case <-ticker.C:
			log.Println("shutdown goroutine")

		default:
			for sig := range c {
				switch sig {

				case os.Interrupt:
					log.Println("shutting down ... ")
					DB.Handle.Close()
					quit <- true

				}

			}
		}

	}
}

func server() {

	manager := gum.NewManager()

	manager.ShutdownOn(syscall.SIGINT)
	manager.ShutdownOn(syscall.SIGTERM)

	// Jobs scheduler
	scheduler := NewScheduler()
	manager.AddUnit(scheduler)

	// API
	api := NewApi()
	manager.AddUnit(api)

	go manager.Run()

	<-manager.Quit

}
