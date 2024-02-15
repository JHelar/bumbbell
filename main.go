package main

import (
	"dumbbell/internal/server"
	"flag"
	"fmt"
	"log"
)

var Version = "0.0.1"

func main() {
	versionFlag := flag.Bool("version", false, "print the version number")
	flag.Parse()

	if *versionFlag {
		fmt.Print(Version)
		return
	}

	srv, err := server.NewServer()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Fatal(srv.ListenAndServe())
}
