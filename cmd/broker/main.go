package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NamanBalaji/flux/internal/broker/api"
	"github.com/NamanBalaji/flux/internal/broker/service"
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

	broker := service.NewBroker()
	go broker.StartRequestPrecessing(*cfg)

	apiRouter := api.SetupRouter(*cfg, broker)

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
			log.Printf("Error occurred while trying to start the server %v \n", err)
			close(serverErrChan)
		}
	}()

	go subscriberCleanupScheduler(*cfg, broker, time.Duration(cfg.Subscriber.CleanupTime)*time.Second)
	go messagesCleanupScheduler(*cfg, broker, time.Duration(cfg.Message.CleanupTime)*time.Second)

	select {
	case <-stopChan:
		log.Println("Shutting down server...")
	case <-serverErrChan:
		log.Fatalf("Shutting down server")
	}
}

func subscriberCleanupScheduler(cfg config.Config, broker *service.Broker, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			log.Println("Cleaning up subscribers")
			broker.CleanSubscribers(cfg)
			log.Printf("Subscriber cleanup completed")
		}
	}
}

func messagesCleanupScheduler(cfg config.Config, broker *service.Broker, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			log.Println("Cleaning up messages")
			broker.CleanupMessages(cfg)
			log.Println("Message cleanup completed")
		}
	}
}
