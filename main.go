package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	Config *config
)

type config struct {
	address string
	port    int
}

func LoadConfig() *config {
	address, found := os.LookupEnv("SERVER_ADDRESS")
	if !found {
		address = ""
	}
	port, found := os.LookupEnv("SERVER_PORT")
	if !found {
		port = "8080"
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("invalid port: %v", err)
	}
	return &config{

		address: address,
		port:    p,
	}
}

func init() {
	log.SetFlags(log.Lshortfile)

	Config = LoadConfig()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res := struct {
		ContentType string `json:"content_type"`
		Msg         string `json:"msg"`
		A           string `json:"a"`
		B           string `json:"b"`
	}{
		ContentType: r.Header.Get("Content-Type"),
		Msg:         "hello",
	}

	if v := r.FormValue("a"); v != "" {
		res.A = v
	}

	if v := r.FormValue("b"); v != "" {
		res.B = v
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("Error:", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Config.address, Config.port),
		Handler: mux,
	}
	log.Printf("Listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() raise an error: %v", err)
	}
}
