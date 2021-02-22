package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Stat struct {
	Time time.Time
	Data string
}

func main() {
	duration := flag.Duration("duration", (7*24)*time.Hour, "the duration for which we'll collect stats.")
	interval := flag.Duration("interval", 5*time.Minute, "interval in which we'll send requests.")
	page := flag.String("page", "https://scribblers-official.herokuapp.com", "page on which we'll call /v1/stats")
	output := flag.String("output", "output.json", "output file for retrieved data")
	flag.Parse()

	url := *page + "/v1/stats"
	outputFile, openError := os.OpenFile(*output, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if openError != nil {
		panic(openError)
	}
	encoder := json.NewEncoder(outputFile)
	shutdownTimer := time.NewTimer(*duration)
	collectionTicker := time.Tick(*interval)
	for {
		select {
		case <-collectionTicker:
			{
				response, err := http.Get(url)
				if err == nil {
					data, err := io.ReadAll(response.Body)
					if err != nil {
						log.Printf("Error reading response: %s\n", err)
					}
					stat := &Stat{
						Time: time.Now(),
						Data: string(data),
					}
					encodeError := encoder.Encode(stat)
					if encodeError != nil {
						log.Printf("Error writing data to hard-drive: %s\n", encodeError)
					}
				} else {
					log.Printf("Error retrieving stats: %s\n", err)
				}
			}
		case <-shutdownTimer.C:
			{
				os.Exit(0)
			}
		}
	}
}
