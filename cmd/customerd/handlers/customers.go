package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/youngkin/mockvideo/internal/customers"
)

type handler struct {
	db *sql.DB
}

// TODO:
//	1.	Don't panic
//	2.	Wrap in a function that will return an error
//	3.	Use logging levels
//	4.	...

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	results, err := h.db.Query("SELECT id, name FROM customer")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	for results.Next() {
		var customer customers.Customer
		// for each row, scan the result into our tag composite object
		err = results.Scan(&customer.ID, &customer.Name)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		fmt.Fprintf(w, "ID: %d, Name: %s\n", customer.ID, customer.Name)
		log.Printf("ID: %d, Name: %s\n", customer.ID, customer.Name)
	}

}

// New returns a *http.Handler configured with a database connection
func New(db *sql.DB) (http.Handler, error) {
	if db == nil {
		return nil, errors.New("non-nil sql.DB connection required")
	}

	return handler{db: db}, nil
}
