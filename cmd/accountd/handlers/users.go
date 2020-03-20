// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handlers

/*
This file attempts to showcase several best practices including:

1.	Structured logging (i.e., 'logger.WithFields(...)')
2.	Log 'hygiene'. Lower level functions don't log. The do return
	errors when necessary and allow the calling function to decide
	if it wants to log the error or propagate the error up the stack.
3.	Error handling:
	i.		Early returns
	ii.		Use of error codes vs. text strings
	iii.	Addition of info to errors to help better understand the
			context the error occurred in.
4. 	Request validation - e.g., verify proper URL path construction
5.	Proper use of HTTP status codes
6.	Use of Prometheus to capture metrics
*/

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
	"github.com/youngkin/mockvideo/internal/platform/constants"
	"github.com/youngkin/mockvideo/internal/users"
)

type handler struct {
	db     *sql.DB
	logger *log.Entry
}

var (
	userRqstDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "mockvideo",
		Subsystem: "user",
		Name:      "request_duration_seconds",
		Help:      "user request duration distribution in seconds",
		// Buckets:   prometheus.ExponentialBuckets(0.005, 1.1, 40),
		Buckets: prometheus.LinearBuckets(0.001, .004, 50),
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(userRqstDur)
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
		userRqstDur.WithLabelValues(strconv.Itoa(httpStatus)).
			Observe(float64(time.Since(start)) / float64(time.Second))
	}

	// Expecting a URL.Path like '/users' or '/users/{id}'
	pathNodes := strings.Split(r.URL.Path, "/")

	if len(r.URL.Path) < 2 {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:  constants.MalformedURLErrorCode,
			constants.HTTPStatus: http.StatusBadRequest,
			constants.Path:       r.URL.Path,
		}).Error(constants.MalformedURL)
		completeRequest(http.StatusBadRequest)
		return

	}
	// Strip off empty string that replaces the first '/' in '/customer'
	pathNodes = pathNodes[1:]

	var (
		payload   interface{}
		err       error
		errReason int
	)

	if len(pathNodes) == 1 {
		payload, err = h.handleGetUsers(pathNodes[0])
	} else {
		payload, errReason, err = h.handleGetOneUser(pathNodes[0], pathNodes[1:])
	}

	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errReason,
			constants.ErrorDetail: err.Error(),
			constants.HTTPStatus:  http.StatusInternalServerError,
		}).Error(constants.UserGETError)
		statusCode := http.StatusInternalServerError
		if errReason == constants.MalformedURLErrorCode {
			statusCode = http.StatusBadRequest
		}
		completeRequest(statusCode)
		return
	}

	custFound := true
	switch p := payload.(type) {
	case nil:
		custFound = false
	case *users.User:
		if p == nil {
			custFound = false
		}
	case *users.Users:
		if len(p.Users) == 0 {
			custFound = false
		}
	default:
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:  constants.UserTypeConversionErrorCode,
			constants.HTTPStatus: http.StatusInternalServerError,
		}).Error(constants.UserTypeConversionError)
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
			constants.HTTPStatus:  http.StatusInternalServerError,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		completeRequest(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)

	userRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handleGetUsers(path string) (interface{}, error) {
	users, err := users.GetAllUsers(h.db)
	if err != nil {
		return nil, errors.Annotate(err, "Error retrieving users from DB")
	}

	h.logger.Debugf("GetAllUsers() results: %+v", users)

	for _, cust := range users.Users {
		cust.HREF = "/" + path + "/" + strconv.Itoa(cust.ID)
	}

	return users, nil
}

// handleGetOneUser will return the customer referenced by the provided resource path,
// an error reason and error if there was a problem retrieving the customer, or a nil customer and a nil
// error if the customer was not found. The error reason will only be relevant when the error
// is non-nil.
func (h handler) handleGetOneUser(path string, pathNodes []string) (cust interface{}, errReason int, err error) {
	if len(pathNodes) > 1 {
		err := errors.Errorf(("expected 1 pathNode, got %d"), len(pathNodes))
		return nil, constants.MalformedURLErrorCode, err
	}

	id, err := strconv.Atoi(pathNodes[0])
	if err != nil {
		err := errors.Annotate(err, fmt.Sprintf("expected numeric pathNode, got %+v", id))
		return nil, constants.MalformedURLErrorCode, err
	}

	c, err := users.GetUser(h.db, id)
	if err != nil {
		return nil, constants.UserGETErrorCode, err
	}
	if c == nil {
		// client will deal with a nil (e.g., not found) customer
		return nil, 0, nil
	}

	h.logger.Debugf("GetUser() results: %+v", c)

	c.HREF = "/" + path + "/" + strconv.Itoa(c.ID)

	return c, 0, nil
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
