package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/juju/errors"
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
		Help:      "customer request duration distribution in seconds",
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
//	1.	Add context deadline to DB requests

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
	completeRequest := func(httpStatus int) {
		w.WriteHeader(httpStatus)
		customerRqstDur.WithLabelValues(strconv.Itoa(httpStatus)).
			Observe(float64(time.Since(start)) / float64(time.Second))
	}

	// Expecting a URL.Path like '/customers' or '/customers/{id}'
	pathNodes := strings.Split(r.URL.Path, "/")

	if len(r.URL.Path) < 2 {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode: constants.MalformedURLErrorCode,
			constants.Path:      r.URL.Path,
		}).Error(constants.MalformedURL)
		completeRequest(http.StatusInternalServerError)
		return

	}
	// Strip off empty string that replaces the first '/' in '/customer'
	pathNodes = pathNodes[1:]

	var (
		payload interface{}
		err     error
	)

	if len(pathNodes) == 1 {
		payload, err = h.handleGetCustomers(pathNodes[0])
	} else {
		payload, err = h.handleGetOneCustomer(pathNodes[0], pathNodes[1:])
	}

	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.CustGETErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		completeRequest(http.StatusInternalServerError)
		return
	}

	custFound := true
	switch p := payload.(type) {
	case nil:
		custFound = false
	case *customers.Customer:
		if p == nil {
			custFound = false
		}
	case *customers.Customers:
		if len(p.Customers) == 0 {
			custFound = false
		}
	default:
		h.logger.WithFields(log.Fields{
			constants.ErrorCode: constants.CustTypeConversionErrorCode,
		}).Error(constants.CustTypeConversionError)
		completeRequest(http.StatusInternalServerError)
		return
	}
	if !custFound {
		h.logger.WithFields(log.Fields{
			constants.HTTPStatus: http.StatusNotFound,
		}).Error("Customer not found")
		completeRequest(http.StatusNotFound)
		return
	}

	marshPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONMarshalingErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		completeRequest(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)

	customerRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handleGetCustomers(path string) (interface{}, error) {
	custs, err := customers.GetAllCustomers(h.db)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.DBRowScanErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.DBRowScanError)
		return nil, errors.Annotate(err, "Error retrieving customers from DB")
	}

	h.logger.Debugf("GetAllCustomers() results: %+v", custs)

	for _, cust := range custs.Customers {
		cust.HREF = "/" + path + "/" + strconv.Itoa(cust.ID)
	}

	return custs, nil
}

// handleGetOneCustomer will return the customer referenced by the provided resource path,
// an error if there was a problem retrieving the customer, or a nil customer and a nil
// error if the customer was not found.
func (h handler) handleGetOneCustomer(path string, pathNodes []string) (interface{}, error) {
	if len(pathNodes) > 1 {
		err := errors.Errorf(("expected 1 pathNode, got %d"), len(pathNodes))
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURL,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.MalformedURL)
		return nil, err
	}

	id, err := strconv.Atoi(pathNodes[0])
	if err != nil {
		err := errors.Annotate(err, fmt.Sprintf("expected numeric pathNode, got %+v", id))
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.MalformedURL)
		return nil, err
	}

	// TODO Need to handle ErrNoRows situation which will result in a nil 'cust' being returned.
	cust, err := customers.GetCustomer(h.db, id)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.DBRowScanErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.DBRowScanError)
		return nil, err
	}
	if cust == nil {
		// client will deal with a nil (e.g., not found) customer
		return nil, nil
	}

	h.logger.Debugf("GetCustomer() results: %+v", cust)

	cust.HREF = "/" + path + "/" + strconv.Itoa(cust.ID)

	return cust, nil
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
