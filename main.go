package main

import (
	"./p3"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	router := p3.NewRouter()
	if len(os.Args) > 1 {
		fmt.Printf("Starting node on port: %s", os.Args[1])
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}
