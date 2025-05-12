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
	Config    *config
	db        *sql.DB = nil
	sampleVar *string = nil
)

type config struct {
	address    string
	port       int
	dbHost     string
	dbPort     string
	dbUser     string
	dbName     string
	dbPassword string
	dbOptions  string
}

func (c *config) isValidDatabaseInfo() bool {
	if c.dbHost == "" || c.dbPort == "" || c.dbUser == "" || c.dbName == "" || c.dbPassword == "" {
		return false
	}
	return true
}

func (c *config) constructDatabaseUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s", c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName, c.dbOptions)
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

	dbHost, found := os.LookupEnv("DB_HOST")
	if !found {
		log.Println("DB_HOST is not set")
	}

	dbPort, found := os.LookupEnv("DB_PORT")
	if !found {
		log.Println("DB_PORT is not set")
	}

	dbUser, found := os.LookupEnv("DB_USER")
	if !found {
		log.Println("DB_USER is not set")
	}

	dbName, found := os.LookupEnv("DB_NAME")
	if !found {
		log.Println("DB_NAME is not set")
	}

	dbPassword, found := os.LookupEnv("DB_PASSWORD")
	if !found {
		log.Println("DB_PASSWORD is not set")
	}

	dbOptions, found := os.LookupEnv("DB_OPTIONS")
	if !found {
		log.Println("DB_OPTIONS is not set")
	}

	return &config{
		address:    address,
		port:       p,
		dbHost:     dbHost,
		dbPort:     dbPort,
		dbUser:     dbUser,
		dbName:     dbName,
		dbPassword: dbPassword,
		dbOptions:  dbOptions,
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func rootHandlerGenerator() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received a '%s' request from %s", r.RequestURI, r.RemoteAddr)
		fmt.Fprintln(w, "Hello, net/http", r.RemoteAddr)
		if sampleVar != nil {
			fmt.Fprintln(w, "SAMPLE_VAR:", *sampleVar)
		}
	}
}

func dbHandlerGenerator(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received a '%s' request from %s", r.RequestURI, r.RemoteAddr)
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

func connectToDB(db *sql.DB) error {
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
				log.Println("Database is down.")
				return err
			}
			continue
		}

		log.Printf("Database is up. Starting server...")
		break
	}
	return nil
}

func main() {
	Config = LoadConfig()

	// Set sample var
	v, found := os.LookupEnv("SAMPLE_VAR")
	if !found {
		log.Println("SAMPLE_VAR is not set")
	}
	sampleVar = &v

	if Config.isValidDatabaseInfo() {
		var err error
		db, err = sql.Open("postgres", Config.constructDatabaseUrl())
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		if err := connectToDB(db); err != nil {
			log.Fatalln(err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandlerGenerator())
	if db != nil {
		mux.HandleFunc("/db", dbHandlerGenerator(db))
	}
	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Config.address, Config.port),
		Handler: mux,
	}
	log.Printf("Listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("ListenAndServe() raise an error: %v", err)
	}
}
