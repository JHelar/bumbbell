package main

import (
	"dumbbell/internal/server"
	"log"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Fatal(srv.ListenAndServe())
}
