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

	// Logging is handled by 'handlePostSingleUser()'
	resp := h.handlePostSingleUser(user)
	resourceID := fmt.Sprintf("/users/%d", resp.User.ID)
	w.Header().Add("Location", resourceID)
	w.WriteHeader(resp.HTTPStatus)
	w.Write([]byte(resp.ErrMsg))
	UserRqstDur.WithLabelValues(strconv.Itoa(resp.HTTPStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePostSingleUser(user domain.User) Response {
	h.logger.Debugf("handlePostSingleUser: user %+v", user)
	if user.ID != 0 { // User ID must *NOT* be populated (i.e., with a non-zero value) on an insert
		errMsg := fmt.Sprintf("expected User.ID > 0, got User.ID = %d", user.ID)
		h.logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.InvalidInsertErrorCode,
			logging.HTTPStatus:  http.StatusBadRequest,
			logging.Path:        fmt.Sprintf("/users/%d", user.ID),
			logging.ErrorDetail: errMsg,
		}).Error(mverr.InvalidInsertErrorMsg)
		return Response{
			HTTPStatus: http.StatusBadRequest,
			ErrMsg:     errMsg,
			ErrReason:  mverr.UserRqstErrorCode,
			User:       user,
		}
	}

	userID, err := h.userSvc.CreateUser(user)
	if err != nil {
		status := http.StatusInternalServerError
		if err.ErrCode == mverr.DBInsertDuplicateUserErrorCode || err.ErrCode == mverr.UserValidationErrorCode {
			status = http.StatusBadRequest
		}
		return Response{
			HTTPStatus: status,
			ErrMsg:     err.ErrMsg,
			ErrReason:  err.ErrCode,
			User:       user,
		}
	}

	user.ID = userID
	user.HREF = fmt.Sprintf("/users/%d", userID)

	return Response{
		HTTPStatus: http.StatusCreated,
		User:       user,
	}

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

	// Logging is handled by 'handlePostSingleUser()'
	resp := h.handlePutSingleUser(*user)
	w.WriteHeader(resp.HTTPStatus)
	w.Write([]byte(resp.ErrMsg))
	UserRqstDur.WithLabelValues(strconv.Itoa(resp.HTTPStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePutSingleUser(user domain.User) Response {
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
		// h.logger.WithFields(log.Fields{
		// 	logging.ErrorCode:   err.ErrCode,
		// 	logging.HTTPStatus:  httpStatus,
		// 	logging.Path:        fmt.Sprintf("/users/%d", user.ID),
		// 	logging.ErrorDetail: err.ErrDetail,
		// }).Error(errMsg)
		resp := Response{
			HTTPStatus: httpStatus,
			ErrMsg:     errMsg,
			ErrReason:  err.ErrCode,
			User:       user,
		}
		return resp
	}

	return Response{
		HTTPStatus: http.StatusOK,
		ErrMsg:     "",
		ErrReason:  mverr.NoErrorCode,
		User:       user,
	}
}

func (h handler) handleRqstMultipleUsers(start time.Time, w http.ResponseWriter, path string, users domain.Users, method string) {
	h.logger.Debugf("handleRqstMultipleUsers for %s", method)
	bp := NewBulkProcessor(h.maxBulkOps)
	defer bp.Stop()

	br := NewBulkRequest(users, method, h.userSvc)
	rqstCompleteC := make(chan Response)
	numUsers := len(users.Users)

	for i := 0; i < numUsers; i++ {
		go h.handleConcurrentRqst(br.Requests[i], bp.RequestC, rqstCompleteC)
	}

	h.logger.Debugf("handleRqstMultipleUsers: Started %d %s goroutines", numUsers, method)
	responses := BulkResponse{}
	overallHTTPStatus := http.StatusOK
	if method == http.MethodPost {
		overallHTTPStatus = http.StatusCreated
	}

	for i := 0; i < numUsers; i++ {
		resp := <-rqstCompleteC
		responses.Results = append(responses.Results, resp)
		if resp.HTTPStatus != http.StatusCreated && resp.HTTPStatus != http.StatusOK {
			overallHTTPStatus = http.StatusConflict
		}
		if resp.ErrReason != mverr.NoErrorCode {
			h.logger.WithFields(log.Fields{
				logging.ErrorCode:   resp.ErrReason,
				logging.HTTPStatus:  resp.HTTPStatus,
				logging.ErrorDetail: fmt.Sprintf("error creating/updating user: Name: %s, email: %s", resp.User.Name, resp.User.EMail),
			}).Errorf(resp.ErrMsg)
		}
	}

	responses.OverallStatus = overallHTTPStatus
	marshResp, err := json.Marshal(responses)
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

	w.WriteHeader(overallHTTPStatus)
	w.Header().Set("Content-Type", "application/json")
	w.Write(marshResp)

	h.logger.Debugf("handleRqstMultipleUsers: reponse sent, bulkprocessor stopped for %s", method)
	UserRqstDur.WithLabelValues(strconv.Itoa(overallHTTPStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
	return
}

func (h handler) handleConcurrentRqst(rqst Request, rqstC chan Request, rqstCompC chan Response) {
	h.logger.Debugf("handleConcurrentRqst: request %+v", rqst)
	rqstC <- rqst
	h.logger.Debug("handleConcurrentRqst: request sent")
	resp := <-rqst.ResponseC
	h.logger.Debugf("handleConcurrentRqst: response %+v received", rqst)
	rqstCompC <- resp
	h.logger.Debug("handleConcurrentRqst: response sent")
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
