package main

import (
	"log"
	"net/http"

	"urlshortener/internal/api"
	"urlshortener/internal/shortener"
)

func main() {
	us := shortener.NewURLShortener()
	router := api.NewRouter(us)
	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
