package handlers

import "net/http"

// HealthFunc returns current health/status of the service
func HealthFunc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("customerd is healthy"))
}
