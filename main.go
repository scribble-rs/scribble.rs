package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/scribble-rs/scribble.rs/communication"
)

func main() {
	portHTTPFlag := flag.Int("portHTTPFlag", -1, "defines the port to be used for http mode")
	flag.Parse()

	var portHTTP int
	if *portHTTPFlag == -1 {
		//Support for heroku, as heroku expects applications to use a specific port.
		envPort, available := os.LookupEnv("PORT")
		if available {
			parsed, parseError := strconv.ParseInt(envPort, 10, 16)
			if parseError == nil {
				portHTTP = int(parsed)
			} else {
				portHTTP = 8080
			}
		} else {
			portHTTP = 8080
		}
	} else {
		portHTTP = *portHTTPFlag
	}

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	log.Printf("Started. Listening on port %d\n", portHTTP)

	//If this ever fails, it will return and print a fatal logger message
	log.Fatal(communication.Serve(portHTTP))
}
