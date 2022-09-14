package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
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
		Addr:    ":8080",
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() raise an error: %v", err)
	}
}
