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
		fmt.Println("Starting node on port: " + os.Args[1])
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		fmt.Print("Missing arguments port yes/no")
	}
}
