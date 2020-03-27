package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/markbates/pkger"
)

var (
	errorPage *template.Template
	portHTTP  *int
)

func readTemplateFile(name string) string {
	templateHandle, pkgerError := pkger.Open("/templates/" + name)
	if pkgerError != nil {
		panic(pkgerError)
	}
	defer templateHandle.Close()

	bytes, readError := ioutil.ReadAll(templateHandle)
	if readError != nil {
		panic(readError)
	}

	return string(bytes)
}

func main() {
	portHTTP = flag.Int("portHTTP", 8080, "defines the port to be used for http mode")

	flag.Parse()

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	var parseError error
	errorPage, parseError = template.New("error.html").Parse(readTemplateFile("error.html"))
	if parseError != nil {
		panic(parseError)
	}
	errorPage, parseError = errorPage.New("footer.html").Parse(readTemplateFile("footer.html"))
	if parseError != nil {
		panic(parseError)
	}

	setupRoutes()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portHTTP), nil))
}

func setupRoutes() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(pkger.Dir("/resources"))))
}
