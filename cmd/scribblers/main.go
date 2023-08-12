package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/scribble-rs/scribble.rs/internal/api"
	"github.com/scribble-rs/scribble.rs/internal/config"
	"github.com/scribble-rs/scribble.rs/internal/frontend"
	"github.com/scribble-rs/scribble.rs/internal/state"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("error loading configuration:", err)
	}

	if cfg.CPUProfilePath != "" {
		log.Println("Starting CPU profiling ....")
		cpuProfileFile, err := os.Create(cfg.CPUProfilePath)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpuProfileFile)
	}

	// Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	api.SetupRoutes()
	frontend.SetupRoutes()
	state.LaunchCleanupRoutine()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer os.Exit(0)

		log.Printf("Received %s, gracefully shutting down.\n", <-signalChan)

		state.ShutdownLobbiesGracefully()
		if cfg.CPUProfilePath != "" {
			pprof.StopCPUProfile()
			log.Println("Finished CPU profiling.")
		}
	}()

	address := fmt.Sprintf("%s:%d", cfg.NetworkAddress, cfg.Port)
	log.Println("Started, listening on:", address)
	log.Fatalln(http.ListenAndServe(address, nil))
}
