package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"go-server/api"
	"go-server/routes"
	"go-server/utils"
)

func main() {
	_ = godotenv.Load()
	config, err := utils.DatabaseConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err := db.Ping(); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	linkHost, linkBaseURL, err := utils.ParseLinkDomain(os.Getenv("LINK_DOMAIN"))
	if err != nil {
		log.Fatal(err)
	}
	app := api.New(db, linkBaseURL)
	app.DeleteExpiredLinks(context.Background())
	go app.CleanupExpiredLinks()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	handler := routes.Admin(app, linkHost == "")
	if linkHost != "" {
		handler = routes.SplitByHost(linkHost, routes.Links(app), handler)
		log.Printf("serving links for host %q, dashboard for every other host", linkHost)
	}
	if linkAddr := strings.TrimSpace(os.Getenv("LINK_ADDR")); linkAddr != "" {
		go func() {
			log.Printf("serving links on http://localhost%s", linkAddr)
			log.Fatal(http.ListenAndServe(linkAddr, utils.Recoverer(routes.Links(app))))
		}()
	}
	log.Printf("Golink listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, utils.Recoverer(handler)))
}
