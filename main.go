package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var (
	Config *config
	db     *sql.DB
)

type AccessLog struct {
	id       int
	datetime time.Time
	path     string
}

type config struct {
	address     string
	port        int
	databaseUrl string
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

	databaseUrl, found := os.LookupEnv("DATABASE_URL")
	if !found {
		log.Fatalf("DATABASE_URL is not set")
	}
	return &config{

		address:     address,
		port:        p,
		databaseUrl: databaseUrl,
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func rootHandlerGenerator(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request from %s", r.RemoteAddr)
		_, err := db.Exec(
			`INSERT INTO access_log (ip, access_ts, url_path) VALUES ($1, $2, $3)`,
			r.RemoteAddr,
			time.Now(),
			r.URL.Path,
		)
		if err != nil {
			log.Printf("failed to insert access_log: %v", err)
			return
		}

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
			return
		}

		rows, err := db.Query(
			`SELECT id, ip, url_path, access_ts FROM "access_log"`,
		)
		if err != nil {
			log.Printf("failed to select access_log: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			id, ip, url_path, access_ts := 0, "", "", ""
			err = rows.Scan(&id, &ip, &url_path, &access_ts)
			if err != nil {
				log.Printf("failed to scan: %v", err)
				return
			}
			log.Println(id, ip, url_path, access_ts)
		}
	}
}

func main() {
	Config = LoadConfig()

	var err error
	db, err = sql.Open("postgres", Config.databaseUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	ctx := context.Background()
	pingDB := func(trial int) error {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			log.Printf("Database is down (%d): %s", trial, err)
			time.Sleep(1 * time.Second)
			return err
		}
		return nil
	}
	trial := 0
	for {
		trial++
		if err := pingDB(trial); err != nil {
			if trial > 30 {
				log.Fatalln("Database is down. Exiting...")
			}
			continue
		}

		log.Printf("Database is up. Starting server...")
		break
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandlerGenerator(db))
	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Config.address, Config.port),
		Handler: mux,
	}
	log.Printf("Listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("ListenAndServe() raise an error: %v", err)
	}
}
