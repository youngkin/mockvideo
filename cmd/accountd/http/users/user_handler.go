// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

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
6.	Detailed SQL error handling (i.e., 'mysql.MySQLError.Number') to set HTTP status codes
7. 	Use of Prometheus to capture metrics
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	//"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/cmd/accountd/services"
	"github.com/youngkin/mockvideo/internal/domain"
	mverr "github.com/youngkin/mockvideo/internal/errors"
	"github.com/youngkin/mockvideo/internal/logging"
)

const rqstStatus = "rqstStatus"

// UserRqstDur is used to capture the length of HTTP requests
var UserRqstDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "mockvideo",
	Subsystem: "user",
	Name:      "user_request_duration_seconds",
	Help:      "user request duration distribution in seconds",
	// Buckets:   prometheus.ExponentialBuckets(0.005, 1.1, 40),
	Buckets: prometheus.LinearBuckets(0.001, .004, 50),
}, []string{rqstStatus})

type handler struct {
	userSvc    services.UserSvcInterface
	logger     *log.Entry
	maxBulkOps int
}

// TODO:
//	1.	Add context deadline to DB requests

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logRqstRcvd(r)
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		fmt.Fprintf(w, "Sorry, only GET, PUT, POST, and DELETE methods are supported.")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Sorry, only GET, PUT, POST, and DELETE methods are supported."))
	}

}

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	completeRequest := func(httpStatus int, msg string) {
		w.WriteHeader(httpStatus)
		w.Write([]byte(msg))
		UserRqstDur.WithLabelValues(strconv.Itoa(httpStatus)).
			Observe(float64(time.Since(start)) / float64(time.Second))
	}

	// Expecting a URL.Path like '/users' or '/users/{id}'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err,
		}).Error(mverr.MalformedURLMsg)
		completeRequest(http.StatusBadRequest, mverr.MalformedURLMsg)
		return
	}

	var payload interface{}
	var err2 *mverr.MVError

	if len(pathNodes) == 1 {
		payload, err2 = h.handleGetUsers(pathNodes[0])
	} else {
		payload, err2 = h.handleGetOneUser(pathNodes[0], pathNodes[1:])
	}

	if err2 != nil {
		httpStatus := http.StatusInternalServerError
		switch err2.ErrCode {
		case mverr.MalformedURLErrorCode:
			httpStatus = http.StatusBadRequest
			h.logger.WithFields(log.Fields{
				logging.ErrorCode:   err2.ErrCode,
				logging.ErrorDetail: err2.Error(),
				logging.HTTPStatus:  httpStatus,
				logging.Path:        r.URL.Path,
			}).Error(err2.ErrMsg)
		case mverr.DBNoUserErrorCode:
			httpStatus = http.StatusNotFound
		}

		// For non-mverr.MalformedURLErrors logging done in the service layer
		completeRequest(httpStatus, err2.ErrMsg)
		return
	}

	marshPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.JSONMarshalingErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.ErrorDetail: err.Error(),
		}).Error(mverr.JSONMarshalingErrorMsg)
		completeRequest(http.StatusBadRequest, mverr.JSONMarshalingErrorMsg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)

	UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handleGetUsers(path string) (interface{}, *mverr.MVError) {
	usrs, err := h.userSvc.GetUsers()
	if err != nil {
		return nil, err
	}

	h.logger.Debugf("GetAllUsers() results: %+v", usrs)

	for _, user := range usrs.Users {
		user.HREF = "/" + path + "/" + strconv.Itoa(user.ID)
	}

	return usrs, nil
}

// handleGetOneUser will return the user referenced by the provided resource path,
// an error reason and error if there was a problem retrieving the user, or a nil user and a nil
// error if the user was not found. The error reason will only be relevant when the error
// is non-nil.
func (h handler) handleGetOneUser(path string, pathNodes []string) (interface{}, *mverr.MVError) {
	if len(pathNodes) != 1 {
		return nil, &mverr.MVError{
			ErrCode:    mverr.MalformedURLErrorCode,
			ErrMsg:     mverr.MalformedURLMsg,
			ErrDetail:  fmt.Sprintf("expected 1 pathNode, got %d, pathNode: %s", len(pathNodes), pathNodes),
			WrappedErr: nil}
	}

	id, err1 := strconv.Atoi(pathNodes[0])
	if err1 != nil {
		return nil, &mverr.MVError{
			ErrCode:    mverr.MalformedURLErrorCode,
			ErrMsg:     mverr.MalformedURLMsg,
			ErrDetail:  fmt.Sprintf("expected numeric user ID, got %+v", id),
			WrappedErr: err1}
	}

	u, err2 := h.userSvc.GetUser(id)
	if err2 != nil {
		return nil, err2
	}

	if u == nil {
		// client will deal with a nil (e.g., not found) user
		return nil, nil
	}

	h.logger.Debugf("GetUser() results: %+v", u)

	u.HREF = "/" + path + "/" + strconv.Itoa(u.ID)

	return u, nil
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	users := domain.Users{}
	user := domain.User{}
	isBulkRqst, err := h.decodeRequest(r, &user, &users)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   err.ErrCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err.ErrDetail,
		}).Error(err.ErrMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ErrDetail))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	// Expecting URL.Path '/users'
	pathNodes, err2 := h.getURLPathNodes(r.URL.Path)
	if err2 != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorMsg:    err2,
			logging.ErrorDetail: fmt.Sprintf("error parsing URL Path %s", r.URL.Path),
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if len(pathNodes) != 1 {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: fmt.Sprintf("expected '/users', got %s", pathNodes),
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if isBulkRqst {
		h.handleRqstMultipleUsers(start, w, r.URL.Path, users, http.MethodPost)
		return
	}

	status := h.handlePostSingleUser(w, user)

	UserRqstDur.WithLabelValues(strconv.Itoa(status)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePostSingleUser(w http.ResponseWriter, user domain.User) int {
	h.logger.Debugf("handlePostSingleUser: user %+v", user)
	if user.ID != 0 { // User ID must *NOT* be populated (i.e., with a non-zero value) on an insert
		errMsg := fmt.Sprintf("expected User.ID = 0, got User.ID = %d", user.ID)
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.InvalidInsertErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        fmt.Sprintf("/users/%d", user.ID),
			logging.ErrorDetail: errMsg,
		}).Error(mverr.InvalidInsertErrorMsg)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return http.StatusBadRequest

	}

	userID, err := h.userSvc.CreateUser(user)
	if err != nil {
		status := http.StatusInternalServerError
		if err.ErrCode == mverr.DBInsertDuplicateUserErrorCode || err.ErrCode == mverr.UserValidationErrorCode {
			status = http.StatusBadRequest
		}
		w.WriteHeader(status)
		w.Write([]byte(err.ErrMsg))
		return status
	}

	user.ID = userID
	user.HREF = fmt.Sprintf("/users/%d", userID)

	w.Header().Add("Location", user.HREF)
	w.WriteHeader(http.StatusCreated)
	return http.StatusCreated
}

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	users := &domain.Users{}
	user := &domain.User{}
	isBulkRqst, err := h.decodeRequest(r, user, users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ErrDetail))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	pathNodes, err2 := h.getURLPathNodes(r.URL.Path)
	if err2 != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err2,
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	// Expecting URL.Path '/users/{id}' or '/users' (on a bulk PUT)
	if len(pathNodes) != 1 && len(pathNodes) != 2 {
		errMsg := fmt.Sprintf("expecting resource path like '/users' or '/users/{id}', got %+v", pathNodes)
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: errMsg,
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if isBulkRqst {
		h.handleRqstMultipleUsers(start, w, r.URL.Path, *users, http.MethodPut)
		return
	}

	status := h.handlePutSingleUser(w, *user)
	UserRqstDur.WithLabelValues(strconv.Itoa(status)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePutSingleUser(w http.ResponseWriter, user domain.User) int {
	err := h.userSvc.UpdateUser(user)
	if err != nil {
		errMsg := mverr.DBUpSertErrorMsg
		httpStatus := http.StatusInternalServerError
		if err.ErrCode == mverr.UserValidationErrorCode {
			httpStatus = http.StatusBadRequest
			errMsg = mverr.UserValidationErrorMsg
		}
		if err.ErrCode == mverr.DBNoUserErrorCode {
			httpStatus = http.StatusBadRequest
			errMsg = mverr.DBNoUserErrorMsg
		}

		w.WriteHeader(httpStatus)
		w.Write([]byte(errMsg))
		return httpStatus
	}

	w.WriteHeader(http.StatusOK)
	return http.StatusOK

}

func (h handler) handleRqstMultipleUsers(start time.Time, w http.ResponseWriter, path string, users domain.Users, method string) {
	h.logger.Debugf("handleRqstMultipleUsers for %s", method)

	var responses *services.BulkResponse
	switch method {
	case http.MethodPost:
		responses, _ = h.userSvc.CreateUsers(users)
	case http.MethodPut:
		responses, _ = h.userSvc.UpdateUsers(users)
	default:
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.BulkRequestErrorCode,
			logging.ErrorDetail: fmt.Errorf("unsupported HTTP Method %s specified in bulk request, only POST and PUT are supported", method),
		}).Error(mverr.BulkRequestErrorMsg)
		responses = &services.BulkResponse{OverallStatus: http.StatusBadRequest}
	}

	marshResp, err := json.Marshal(*responses)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.JSONMarshalingErrorCode,
			logging.HTTPStatus:  http.StatusInternalServerError,
			logging.ErrorDetail: err.Error(),
		}).Error(mverr.JSONMarshalingErrorMsg)
		w.WriteHeader(http.StatusInternalServerError)
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	overallStatus := mapStatusToHTTPStatus(responses.OverallStatus)
	w.WriteHeader(overallStatus)
	w.Header().Set("Content-Type", "application/json")
	w.Write(marshResp)

	h.logger.Debugf("handleRqstMultipleUsers: response %s for method %s with HTTP Status %d", marshResp, method, overallStatus)
	UserRqstDur.WithLabelValues(strconv.Itoa(overallStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
	return
}

func mapStatusToHTTPStatus(status services.Status) int {
	var httpStatus int
	switch status {
	case services.StatusBadRequest:
		httpStatus = http.StatusBadRequest
	case services.StatusConflict:
		httpStatus = http.StatusConflict
	case services.StatusCreated:
		httpStatus = http.StatusCreated
	case services.StatusNotFound:
		httpStatus = http.StatusNotFound
	case services.StatusOK:
		httpStatus = http.StatusOK
	case services.StatusServerError:
		httpStatus = http.StatusInternalServerError
	default:
		httpStatus = http.StatusInternalServerError
	}

	return httpStatus
}

func (h handler) decodeRequest(r *http.Request, user *domain.User, users *domain.Users) (bool, *mverr.MVError) {
	// Get user(s) out of request body and validate
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if user sends extra data
	var err error
	isBulkRqst := false
	hVal, ok := r.Header["Bulk-Request"]

	if ok {
		isBulkRqst, err = strconv.ParseBool(hVal[0])
		if err != nil {
			return isBulkRqst, &mverr.MVError{
				ErrCode:    mverr.UserRqstErrorCode,
				ErrDetail:  fmt.Sprintf("Expected 'true' or 'false' value for 'Bulk-Request' header, got %s", hVal[0]),
				ErrMsg:     mverr.UserRqstErrorMsg,
				WrappedErr: err,
			}
		}
	}

	if isBulkRqst {
		err = d.Decode(users)
	} else {
		err = d.Decode(user)
	}
	if err != nil {
		return isBulkRqst, &mverr.MVError{
			ErrCode:    mverr.JSONDecodingErrorCode,
			ErrDetail:  err.Error(),
			ErrMsg:     mverr.JSONDecodingErrorMsg,
			WrappedErr: err,
		}
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.JSONDecodingErrorCode,
			logging.ErrorMsg:    mverr.JSONDecodingErrorMsg,
			logging.ErrorDetail: fmt.Sprintf("JSON request body contained unexpected data: %s", r.Body),
		}).Warn(mverr.JSONDecodingErrorMsg)
	}

	return isBulkRqst, nil
}

func (h handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err,
		}).Error(mverr.MalformedURLMsg)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if len(pathNodes) != 2 {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: fmt.Sprintf("expecting resource path like /users/{id}, got %+v", pathNodes),
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	uid, err := strconv.Atoi(pathNodes[1])
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: fmt.Sprintf("Invalid resource ID, must be int, got %v", pathNodes[1]),
		}).Error(mverr.MalformedURLMsg)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}
	err2 := h.userSvc.DeleteUser(uid)
	if err2 != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   err2.ErrCode,
			logging.HTTPStatus:  http.StatusInternalServerError,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err2.ErrDetail,
		}).Error(err2.ErrMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(mverr.DBDeleteErrorMsg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	w.WriteHeader(http.StatusOK)

	UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusCreated)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) logRqstRcvd(r *http.Request) {
	h.logger.WithFields(log.Fields{
		logging.Method:     r.Method,
		logging.Path:       r.URL.Path,
		logging.RemoteAddr: r.RemoteAddr,
	}).Info("HTTP request received")
}

func (h handler) getURLPathNodes(path string) ([]string, error) {
	pathNodes := strings.Split(path, "/")

	if len(pathNodes) < 2 {
		return nil, errors.New(mverr.MalformedURLMsg)
	}

	// Strip off empty string that replaces the first '/' in '/users'
	pathNodes = pathNodes[1:]

	// Strip off the empty string that replaces the second '/' in '/users/'
	if pathNodes[len(pathNodes)-1] == "" {
		pathNodes = pathNodes[0 : len(pathNodes)-1]
	}

	return pathNodes, nil
}

func (h handler) parseRqst(r *http.Request) (domain.User, []string, error) {
	//
	// Get user out of request body and validate
	//
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if user sends extra data
	u := domain.User{}
	err := d.Decode(&u)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.JSONDecodingErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.ErrorDetail: err.Error(),
		}).Error(mverr.JSONDecodingErrorMsg)

		return domain.User{}, nil, err
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.JSONDecodingErrorCode,
			logging.ErrorDetail: fmt.Sprintf("Additional JSON after User data: %v", u),
		}).Warn(mverr.JSONDecodingErrorMsg)
	}

	// Expecting a URL.Path like '/users/{id}'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.MalformedURLErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        r.URL.Path,
			logging.ErrorDetail: err,
		}).Error(mverr.MalformedURLMsg)

		return domain.User{}, nil, err
	}

	return u, pathNodes, nil
}

// NewUserHandler returns a properly configured *http.Handler
func NewUserHandler(userSvc services.UserSvcInterface, logger *log.Entry, maxBulkOps int) (http.Handler, error) {
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}
	if maxBulkOps == 0 {
		return nil, errors.New("maxBulkOps must be greater than zero")
	}
	return handler{userSvc: userSvc, maxBulkOps: maxBulkOps, logger: logger}, nil
}
