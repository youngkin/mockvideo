package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"
)

func main() {
	fmt.Printf("Hello, %s\n", "WORLD!")

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
	mux.Handle("/health", healthHandler)

	fmt.Printf("Starting server on port: %d\n", 9090)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 9090), mux))
}
