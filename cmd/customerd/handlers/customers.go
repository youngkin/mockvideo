package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/youngkin/mockvideo/internal/customers"
)

type handler struct {
	db *sql.DB
}

var (
	customerRqstDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "mockvideo",
		Subsystem: "customer",
		Name:      "request_duration_seconds",
		Help:      "customer request duration distribution",
		// Buckets:   prometheus.ExponentialBuckets(0.005, 1.1, 40),
		Buckets: prometheus.LinearBuckets(0.001, .004, 50),
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(customerRqstDur)
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

// TODO:
//	1.	Don't panic
//	2.	Wrap in a function that will return an error
//	3.	Use logging levels
//	4.	Add context deadline to DB requests

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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

	customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

// New returns a *http.Handler configured with a database connection
func New(db *sql.DB) (http.Handler, error) {
	if db == nil {
		return nil, errors.New("non-nil sql.DB connection required")
	}

	return handler{db: db}, nil
}
