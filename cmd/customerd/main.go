package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO:
//	1.	TODO: HTTP connection config (e.g., timeouts)
//	2.	TODO: Don't panic
//	3.	TODO: Use logging levels
//	4.	TODO: Config parms (configMap?)
//	5.	TODO: Prometheus
//	6.	TODO: Kube logging
//	7.	TODO: ELK stack for logging
//	8.	TODO: Graphana for Prometheus
//	9.	TODO: Helm/Kube deployment (with P9S support)
//	10.	TODO: Create build system that will compile and create docker image
//	11.	TODO: Use https
//	9.	DONE: Travis CI

func main() {
	port := flag.Int("port", 5999, "the port to start the customer service on")
	flag.Parse()

	db, err := sql.Open("mysql", "admin:2girls1cat@tcp(10.0.0.100:3306)/mockvideo")

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	customersHandler, err := handlers.New(db)
	if err != nil {
		log.Printf("Error initializing customers HTTP handler, error: %s\n", err)
		os.Exit(1)
	}
	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()

	mux.Handle("/customers", customersHandler)
	mux.Handle("/custdhealth", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("customerd starting on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
