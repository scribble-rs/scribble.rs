package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/scribble-rs/scribble.rs/api"
	"github.com/scribble-rs/scribble.rs/frontend"
)

func main() {
	portHTTPFlag := flag.Int("portHTTP", -1, "defines the port to be used for http mode")
	flag.Parse()

	var portHTTP int
	if *portHTTPFlag != -1 {
		portHTTP = *portHTTPFlag
		log.Printf("Listening on port %d sourced from portHTTP flag.\n", portHTTP)
	} else {
		//Support for heroku, as heroku expects applications to use a specific port.
		envPort, portVarAvailable := os.LookupEnv("PORT")
		if portVarAvailable {
			log.Printf("'PORT' environment variable found: '%s'\n", envPort)
			parsed, parseError := strconv.ParseInt(envPort, 10, 32)
			if parseError == nil {
				portHTTP = int(parsed)
				log.Printf("Listening on port %d sourced from 'PORT' environment variable\n", portHTTP)
			} else {
				log.Printf("Error parsing 'PORT' variable: %s\n", parseError)
				log.Println("Falling back to default port.")
			}
		}
	}

	if portHTTP == 0 {
		portHTTP = 8080
		log.Printf("Listening on default port %d\n", portHTTP)
	}

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	log.Println("Started.")

	api.SetupRoutes()
	frontend.SetupRoutes()

	http.ListenAndServe(fmt.Sprintf(":%d", portHTTP), nil)
}
