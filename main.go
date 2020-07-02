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
	portHTTPFlag := flag.Int("portHTTP", -1, "defines the port to be used for http mode")
	flag.Parse()

	var portHTTP int
	if *portHTTPFlag != -1 {
		portHTTP = *portHTTPFlag
		log.Printf("Listening on port %d sourced from portHTTP flag.\n", portHTTP)
	} else {
		//Support for heroku, as heroku expects applications to use a specific port.
		envPort, _ := os.LookupEnv("PORT")
		parsed, parseError := strconv.ParseInt(envPort, 10, 16)
		if parseError == nil {
			portHTTP = int(parsed)
			log.Printf("Listening on port %d sourced from PORT environment variable\n", portHTTP)
		} else {
			portHTTP = 24000
			log.Printf("Listening on default port %d\n", portHTTP)
		}
	}

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	log.Println("Started.")

	//If this ever fails, it will return and print a fatal logger message
	log.Fatal(communication.Serve(portHTTP))
}
