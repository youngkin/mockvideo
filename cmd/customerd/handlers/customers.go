package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/customers"
	"github.com/youngkin/mockvideo/internal/platform/constants"
)

type handler struct {
	db     *sql.DB
	logger *log.Entry
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
//	1.	Write unit test
//  2.  Add context deadline to DB requests

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
	h.logger.WithFields(log.Fields{
		constants.Method:     r.Method,
		constants.HostName:   r.URL.Host,
		constants.Path:       r.URL.Path,
		constants.RemoteAddr: r.RemoteAddr,
	}).Info("HTTP request received")
	start := time.Now()

	results, err := h.db.Query("SELECT id, name, streetAddress, city, state, country FROM customer")
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.AppError:    constants.DBQueryError,
			constants.ErrorDetail: err.Error(),
		}).Error("Error querying DB")
		w.WriteHeader(http.StatusInternalServerError)

		customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))

		return
	}

	custs := customers.Customers{}
	for results.Next() {
		var customer customers.Customer

		err = results.Scan(&customer.ID,
			&customer.Name,
			&customer.StreetAddress,
			&customer.City,
			&customer.State,
			&customer.Country)
		if err != nil {
			h.logger.WithFields(log.Fields{
				constants.AppError:    constants.DBRowScanError,
				constants.ErrorDetail: err.Error(),
			}).Error("Error scanning resultset")
			w.WriteHeader(http.StatusInternalServerError)
			customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))

			customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))

			return
		}

		customer = customers.Customer{
			ID:            customer.ID,
			Name:          customer.Name,
			StreetAddress: customer.StreetAddress,
			City:          customer.City,
			State:         customer.State,
			Country:       customer.Country,
		}
		custs.Customers = append(custs.Customers, customer)

		s := fmt.Sprintf("ID: %d, Name: %s, Address: %s, City: %s, State: %s, Country: %s\n",
			customer.ID, customer.Name, customer.State, customer.City, customer.State, customer.Country)
		h.logger.Debug(s)

	}

	marshCusts, err := json.Marshal(custs)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.AppError:    constants.JSONMarshalingError,
			constants.ErrorDetail: err.Error(),
		}).Error("Error marshaling JSON")
		w.WriteHeader(http.StatusInternalServerError)

		customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshCusts)

	customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not implemented")
	w.WriteHeader(http.StatusTeapot)
}

// New returns a *http.Handler configured with a database connection
func New(db *sql.DB, logger *log.Entry) (http.Handler, error) {
	if db == nil {
		return nil, errors.New("non-nil sql.DB connection required")
	}
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}

	return handler{db: db, logger: logger}, nil
}
