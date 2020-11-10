package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/kelseyhightower/envconfig"
)

var (
	config configuration
	db     *sql.DB
)

func initServer(ctx context.Context) *http.Server {
	var mux Mux
	// Add a default root 200 handler
	mux.OK("/")
	// Add a version endpoint
	mux.File("/ami-version", "version info", "/etc/ami_version")
	// Add a health handler
	mux.GET("/health", requestHandler(handleHealth))
	addStatusHandlers(&mux)
	if config.EnableDebugEndpoints {
		log.Print("Enabling Debug Endpoints")
		addDebugHandlers(&mux)
	}
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		Handler:           &mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
}

func initDB() {
	var err error
	db, err = sql.Open("pgx", config.Connstr)
	if err != nil {
		log.Printf("Connection string is invalid: %s", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		log.Printf("Could not connect to database: %s", err)
		os.Exit(1)
	}
	log.Printf("Connected to PGBouncer database")
}

func version() {
	fmt.Fprintf(os.Stderr, "%s version %s\n", os.Args[0], VERSION)
	fmt.Fprintf(os.Stderr, "built %s\n", BUILD_DATE)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("Could not process configuration: %s", err)
	}
	flag.Usage = func() {
		version()
		fmt.Fprintln(os.Stderr, "\nUsage:")
		flag.PrintDefaults()
		if err := envconfig.Usage("", &config); err != nil {
			log.Fatalf("Could not process configuration: %s", err)
		}
	}
	flag.Parse()
	initDB()
	webserver := initServer(ctx)
	log.Printf("Listening on port %d", config.Port)
	go func() {
		if err := webserver.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Error occured while listening for connections: %s", err)
		}
	}()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := webserver.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
