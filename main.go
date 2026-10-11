package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	flags, err := NewFlagsClient()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/payments/authorize", authorizeHandler(flags, time.Now))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
