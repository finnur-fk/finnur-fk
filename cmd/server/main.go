package main

import (
	"flag"
	"log"

	"github.com/finnur-fk/finnur-fk/api"
)

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	server := api.NewServer()
	
	log.Fatal(server.Start(*port))
}
