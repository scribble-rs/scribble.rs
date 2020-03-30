package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/scribble-rs/scribble.rs/communication"
)

var (
	portHTTP *int
)

func main() {
	portHTTP = flag.Int("portHTTP", 8080, "defines the port to be used for http mode")
	flag.Parse()

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	log.Println("Started")

	//If this ever fails, it will return and print a fatal logger message
	log.Fatal(communication.Serve(*portHTTP))
}
