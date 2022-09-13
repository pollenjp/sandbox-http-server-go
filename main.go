package main

import (
	"fmt"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, net/http")
	})
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() raise an error: %v", err)
	}
}
