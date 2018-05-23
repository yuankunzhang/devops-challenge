package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

func main() {
	controller, err := NewController()
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan struct{})
	defer close(stop)

	// Run the worker loop to process items.
	go controller.Run(stop)

	log.Info("Controller running...")

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGTERM or SIGINT.
	<-term
}
