package handlers

import "net/http"

// CustomersFunc defines a function handler that returns a Hello, World response
func CustomersFunc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from CustomersFunc!"))
}
