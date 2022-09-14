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

type IpV4 struct {
	// v4.v4.v2.v1
	v4 int
	v3 int
	v2 int
	v1 int
}

func NewIpV4FromString(address string) (*IpV4, error) {
	var v4, v3, v2, v1 int
	if _, err := fmt.Sscanf(address, "%d.%d.%d.%d", &v4, &v3, &v2, &v1); err != nil {
		return nil, err
	}
	return NewIpV4(v4, v3, v2, v1)
}

func NewIpV4(v4, v3, v2, v1 int) (*IpV4, error) {
	for _, v := range []int{v4, v3, v2, v1} {
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("invalid ip address: %d", v)
		}
	}
	return &IpV4{
		v4: v4,
		v3: v3,
		v2: v2,
		v1: v1,
	}, nil
}

func (ip *IpV4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip.v4, ip.v3, ip.v2, ip.v1)
}

type config struct {
	ipv4 IpV4
	port int
}

func LoadConfig() *config {
	ipv4, found := os.LookupEnv("SERVER_ADDRESS")
	if !found {
		ipv4 = "127.0.0.1"
	}
	port, found := os.LookupEnv("SERVER_PORT")
	if !found {
		port = "8080"
	}

	address, err := NewIpV4FromString(ipv4)
	if err != nil {
		log.Fatalf("invalid ip address: %v", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("invalid port: %v", err)
	}
	return &config{

		ipv4: *address,
		port: p,
	}
}

func init() {
	log.SetFlags(log.Lshortfile)

	Config = LoadConfig()
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
		Addr:    fmt.Sprintf("%s:%d", Config.ipv4.String(), Config.port),
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() raise an error: %v", err)
	}
}
