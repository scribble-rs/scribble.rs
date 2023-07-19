package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/frontend"
	"github.com/scribble-rs/scribble.rs/state"
)

const defaultPort = 8080

func determinePort(portHTTPFlag int) int {
	portHTTP := -1
	if portHTTPFlag != -1 {
		portHTTP = portHTTPFlag
		log.Printf("Listening on port %d sourced from portHTTP flag.\n", portHTTP)
	} else {
		// Support for heroku, as heroku expects applications to use a specific port.
		envPort, portVarAvailable := os.LookupEnv("PORT")
		if portVarAvailable {
			log.Printf("'PORT' environment variable found: '%s'\n", envPort)
			parsed, parseError := strconv.ParseInt(strings.TrimSpace(envPort), 10, 32)
			if parseError == nil {
				portHTTP = int(parsed)
				log.Printf("Listening on port %d sourced from 'PORT' environment variable\n", portHTTP)
			} else {
				log.Printf("Error parsing 'PORT' variable: %s\n", parseError)
				log.Println("Falling back to default port.")
			}
		}
	}

	if portHTTP != -1 && portHTTP < 0 || portHTTP > 65535 {
		log.Println("Port has to be between 0 and 65535.")
		log.Println("Falling back to default port.")
		portHTTP = -1
	}

	if portHTTP < 0 {
		portHTTP = defaultPort
		log.Printf("Listening on default port %d\n", portHTTP)
	}

	return portHTTP
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	portHTTPFlag := flag.Int("portHTTP", -1, "defines the port to be used for http mode")
	flag.Parse()

	if *cpuprofile != "" {
		log.Println("Starting CPU profiling ....")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	portHTTP := determinePort(*portHTTPFlag)

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
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
			log.Println("Finished CPU profiling.")
		}
	}()

	log.Println("Started.")
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", portHTTP), nil))
}
