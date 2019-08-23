package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"
)

type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

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

	results, err := db.Query("SELECT id, name FROM customer")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	for results.Next() {
		var customer Customer
		// for each row, scan the result into our tag composite object
		err = results.Scan(&customer.ID, &customer.Name)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		log.Printf("ID: %d, Name: %s\n", customer.ID, customer.Name)
	}

	mux := http.NewServeMux()
	// 't' and 'h' below experiment with 'HandlerOptions' defined as part of the 'NewHandler'
	// 'story.NewHandler(s)' will use defaults for the template and path resolution
	customersHandler := http.HandlerFunc(handlers.CustomersFunc)
	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux.Handle("/customers", customersHandler)
	mux.Handle("/health", healthHandler)

	fmt.Printf("Starting server on port: %d\n", 9090)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 9090), mux))
}
