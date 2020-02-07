package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/cmd/customerd/logging"
	"github.com/youngkin/mockvideo/internal/customers"
	"github.com/youngkin/mockvideo/internal/platform/constants"
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

	logger *log.Entry
)

func init() {
	prometheus.MustRegister(customerRqstDur)
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	logger = logging.GetLogger()
}

// TODO:
//	1.	Don't panic
//	2.	Wrap in a function that will return an error
//	3.	Use logging levels
//	4.	Add context deadline to DB requests

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case "POST":
		h.handlePost(w, r)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		w.WriteHeader(http.StatusTeapot)
	}

}

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(log.Fields{
		constants.Method:      r.Method,
		constants.URLHostName: r.URL.Host,
		constants.Path:        r.URL.Path,
		constants.RemoteAddr:  r.RemoteAddr,
	}).Info("HTTP request received")
	start := time.Now()

	results, err := h.db.Query("SELECT id, name, streetAddress, city, state, country FROM customer")
	if err != nil {
		logger.WithFields(log.Fields{
			constants.AppError:    constants.DBQueryError,
			constants.ErrorDetail: err.Error(),
		}).Error("Error querying DB")
	}

	for results.Next() {
		var customer customers.Customer

		err = results.Scan(&customer.ID,
			&customer.Name,
			&customer.StreetAddress,
			&customer.City,
			&customer.State,
			&customer.Country)
		if err != nil {
			logger.WithFields(log.Fields{
				constants.AppError:    constants.DBRowScanError,
				constants.ErrorDetail: err.Error(),
			}).Error("Error scanning resultset")
			w.WriteHeader(http.StatusInternalServerError)
			customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))
			return
		}

		fmt.Fprintf(w, "ID: %d, Name: %s, Address: %s, City: %s, State: %s, Country: %s\n",
			customer.ID, customer.Name, customer.State, customer.City, customer.State, customer.Country)
	}

	customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not implemented")
	w.WriteHeader(http.StatusTeapot)
}

// New returns a *http.Handler configured with a database connection
func New(db *sql.DB) (http.Handler, error) {
	if db == nil {
		return nil, errors.New("non-nil sql.DB connection required")
	}

	return handler{db: db}, nil
}
