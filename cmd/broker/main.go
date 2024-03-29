package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NamanBalaji/flux/internal/broker/api"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/constants"
)

func main() {
	configFile := flag.String(constants.ConfigFlag, fmt.Sprintf("cmd/broker/%s", constants.DefaultConfigFile), "config file name")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	broker := service.NewBroker(cfg.DefaultPartitions)

	apiRouter := api.SetupRouter(broker)

	port := fmt.Sprintf(":%d", cfg.Api.Port)

	server := &http.Server{
		Addr:    port,
		Handler: apiRouter,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	serverErrChan := make(chan struct{})

	go func() {
		log.Printf("starting broken on port %s", port)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error occured while trying to start the server %v \n", err)
			close(serverErrChan)
		}
	}()

	select {
	case <-stopChan:
		log.Println("Shutting down server...")
	case <-serverErrChan:
		log.Fatalf("Shutting down server")
	}
}
