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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/cmd/accountd/usecases"
	"github.com/youngkin/mockvideo/internal/constants"
	"github.com/youngkin/mockvideo/internal/domain"
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
	uc         usecases.UserUC
	ur         domain.UserRepository
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
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		completeRequest(http.StatusBadRequest, constants.MalformedURL)
		return
	}

	var (
		payload   interface{}
		errReason domain.ErrCode
	)

	if len(pathNodes) == 1 {
		payload, err = h.handleGetUsers(pathNodes[0])
	} else {
		payload, errReason, err = h.handleGetOneUser(pathNodes[0], pathNodes[1:])
	}

	if err != nil {
		errMsg := constants.UserRqstError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errReason,
			constants.ErrorDetail: err.Error(),
			constants.HTTPStatus:  http.StatusInternalServerError,
		}).Error(errMsg)
		statusCode := http.StatusInternalServerError
		if errReason == constants.MalformedURLErrorCode {
			statusCode = http.StatusBadRequest
			errMsg = constants.MalformedURL
		}
		completeRequest(statusCode, errMsg)
		return
	}

	userFound := true
	switch p := payload.(type) {
	case nil:
		userFound = false
	case *domain.User:
		if p == nil {
			userFound = false
		}
	case *domain.Users:
		if p == nil || len(p.Users) == 0 {
			userFound = false
		}
	default:
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:  constants.UserTypeConversionErrorCode,
			constants.HTTPStatus: http.StatusInternalServerError,
		}).Error(constants.UserTypeConversionError)
		completeRequest(http.StatusInternalServerError, constants.UserTypeConversionError)
		return
	}
	if !userFound {
		h.logger.WithFields(log.Fields{
			constants.HTTPStatus: http.StatusNotFound,
		}).Error("User not found")
		completeRequest(http.StatusNotFound, "")
		return
	}

	marshPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONMarshalingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		completeRequest(http.StatusBadRequest, constants.JSONMarshalingError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)

	UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handleGetUsers(path string) (interface{}, error) {
	usrs, err := h.ur.GetUsers()
	if err != nil {
		return nil, errors.Annotate(err, "Error retrieving users from DB")
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
func (h handler) handleGetOneUser(path string, pathNodes []string) (cust interface{}, errReason domain.ErrCode, err error) {
	if len(pathNodes) > 1 {
		err := errors.Errorf(("expected 1 pathNode, got %d"), len(pathNodes))
		return nil, constants.MalformedURLErrorCode, err
	}

	id, err := strconv.Atoi(pathNodes[0])
	if err != nil {
		err := errors.Annotate(err, fmt.Sprintf("expected numeric pathNode, got %+v", id))
		return nil, constants.MalformedURLErrorCode, err
	}

	c, err := h.ur.GetUser(id)
	if err != nil {
		return nil, constants.UserRqstErrorCode, err
	}
	if c == nil {
		// client will deal with a nil (e.g., not found) user
		return nil, 0, nil
	}

	h.logger.Debugf("GetUser() results: %+v", c)

	c.HREF = "/" + path + "/" + strconv.Itoa(c.ID)

	return c, 0, nil
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	users := &domain.Users{}
	user := &domain.User{}
	isBulkRqst, msg, err := h.decodeRequest(r, user, users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	// Expecting URL.Path '/users'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if len(pathNodes) != 1 {
		errMsg := fmt.Sprintf("expected '/users', got %s", pathNodes)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if isBulkRqst {
		h.handleRqstMultipleUsers(start, w, r.URL.Path, *users, http.MethodPost)
		return
	}

	// Logging is handled by 'handlePostSingleUser()'
	resp := h.handlePostSingleUser(*user)
	resourceID := fmt.Sprintf("/users/%d", resp.User.ID)
	w.Header().Add("Location", resourceID)
	w.WriteHeader(resp.HTTPStatus)
	w.Write([]byte(resp.ErrMsg))
	UserRqstDur.WithLabelValues(strconv.Itoa(resp.HTTPStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handlePostSingleUser(u domain.User) Response {
	h.logger.Debugf("handlePostSingleUser: user %+v", u)
	if u.ID != 0 { // User ID must *NOT* be populated (i.e., with a non-zero value) on an insert
		errMsg := fmt.Sprintf("expected User.ID > 0, got User.ID = %d", u.ID)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.InvalidInsertErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        fmt.Sprintf("/users/%d", u.ID),
			constants.ErrorDetail: errMsg,
		}).Error(constants.InvalidInsertError)
		return Response{
			HTTPStatus: http.StatusBadRequest,
			ErrMsg:     errMsg,
			ErrReason:  constants.UserRqstErrorCode,
			User:       u,
		}
	}

	userID, errReason, err := h.ur.CreateUser(u)
	if err != nil {
		errMsg := constants.DBUpSertError
		status := http.StatusInternalServerError
		errCode := constants.DBUpSertErrorCode
		if errReason == constants.DBInsertDuplicateUserErrorCode {
			// Invalid to insert a duplicate user, this is a client error hence the StatusBadRequest
			status = http.StatusBadRequest
			errMsg = constants.DBInsertDuplicateUserError
			errCode = constants.DBInsertDuplicateUserErrorCode
		}
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  status,
			constants.Path:        fmt.Sprintf("/users/%d", u.ID),
			constants.ErrorDetail: err,
		}).Error(errMsg)
		return Response{
			HTTPStatus: http.StatusBadRequest,
			ErrMsg:     errMsg,
			ErrReason:  errCode,
			User:       u,
		}
	}

	u.ID = userID
	u.HREF = fmt.Sprintf("/users/%d", userID)

	return Response{
		HTTPStatus: http.StatusCreated,
		ErrMsg:     "",
		ErrReason:  constants.NoErrorCode,
		User:       u,
	}

}

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	users := &domain.Users{}
	user := &domain.User{}
	isBulkRqst, msg, err := h.decodeRequest(r, user, users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	// Expecting URL.Path '/users/{id}' or '/users' (on a bulk PUT)
	if len(pathNodes) != 1 && len(pathNodes) != 2 {
		errMsg := fmt.Sprintf("expecting resource path like '/users' or '/users/{id}', got %+v", pathNodes)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.MalformedURL)
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

func (h handler) handlePutSingleUser(usr domain.User) Response {
	errCode, err := h.ur.UpdateUser(usr)
	if err != nil {
		errMsg := constants.DBUpSertError
		httpStatus := http.StatusInternalServerError
		if errCode == constants.DBInvalidRequestCode {
			httpStatus = http.StatusBadRequest
			errMsg = constants.DBInvalidRequest
		}
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        fmt.Sprintf("/users/%d", usr.ID),
			constants.ErrorDetail: err,
		}).Error(errMsg)
		resp := Response{
			HTTPStatus: httpStatus,
			ErrMsg:     errMsg,
			ErrReason:  errCode,
			User:       usr,
		}
		return resp
	}

	return Response{
		HTTPStatus: http.StatusOK,
		ErrMsg:     "",
		ErrReason:  constants.NoErrorCode,
		User:       usr,
	}
}

func (h handler) handleRqstMultipleUsers(start time.Time, w http.ResponseWriter, path string, users domain.Users, method string) {
	h.logger.Debugf("handleRqstMultipleUsers for %s", method)
	bp := NewBulkProcessor(h.maxBulkOps)
	defer bp.Stop()

	br := NewBulkRequest(users, method, h)
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
		if resp.ErrReason != constants.NoErrorCode {
			h.logger.WithFields(log.Fields{
				constants.ErrorCode:   resp.ErrReason,
				constants.HTTPStatus:  resp.HTTPStatus,
				constants.ErrorDetail: resp.ErrMsg,
			}).Error(resp.ErrMsg)
		}
	}

	responses.OverallStatus = overallHTTPStatus
	marshResp, err := json.Marshal(responses)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONMarshalingErrorCode,
			constants.HTTPStatus:  http.StatusInternalServerError,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
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

func (h handler) decodeRequest(r *http.Request, user *domain.User, users *domain.Users) (bool, string, error) {
	// Get user(s) out of request body and validate
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if user sends extra data
	var err error
	isBulkRqst := false
	hVal, ok := r.Header["Bulk-Request"]

	if ok {
		isBulkRqst, err = strconv.ParseBool(hVal[0])
		if err != nil {
			errMsg := fmt.Sprintf("Expected 'true' or 'false' value for 'Bulk-Request' header, got %s", hVal[0])
			h.logger.WithFields(log.Fields{
				constants.ErrorCode:   constants.UserRqstErrorCode,
				constants.HTTPStatus:  http.StatusBadRequest,
				constants.ErrorDetail: errMsg,
				"Bulk-Request:":       isBulkRqst,
			}).Warn(constants.JSONDecodingError)
			return isBulkRqst, errMsg, err
		}
	}

	if isBulkRqst {
		err = d.Decode(&users)
	} else {
		err = d.Decode(&user)
	}
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
			"Bulk-Request:":       isBulkRqst,
		}).Error(constants.JSONDecodingError)
		return isBulkRqst, constants.JSONDecodingError, err
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Warn(constants.JSONDecodingError)
	}

	return isBulkRqst, "", nil
}

func (h handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if len(pathNodes) != 2 {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("expecting resource path like /users/{id}, got %+v", pathNodes),
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	uid, err := strconv.Atoi(pathNodes[1])
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("Invalid resource ID, must be int, got %v", pathNodes[1]),
		}).Error(constants.MalformedURL)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}
	errCode, err := h.ur.DeleteUser(uid)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  http.StatusInternalServerError,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(errCode)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.DBDeleteError))
		UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	w.WriteHeader(http.StatusOK)

	UserRqstDur.WithLabelValues(strconv.Itoa(http.StatusCreated)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) logRqstRcvd(r *http.Request) {
	h.logger.WithFields(log.Fields{
		constants.Method:     r.Method,
		constants.Path:       r.URL.Path,
		constants.RemoteAddr: r.RemoteAddr,
	}).Info("HTTP request received")
}

func (h handler) getURLPathNodes(path string) ([]string, error) {
	pathNodes := strings.Split(path, "/")

	if len(pathNodes) < 2 {
		return nil, errors.New(constants.UserRqstError)
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
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)

		return domain.User{}, nil, err
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: fmt.Sprintf("Additional JSON after User data: %v", u),
		}).Warn(constants.JSONDecodingError)
	}

	// Expecting a URL.Path like '/users/{id}'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		return domain.User{}, nil, err
	}

	return u, pathNodes, nil
}

// NewUserHandler returns a properly configured *http.Handler
func NewUserHandler(uc usecases.UserUC, ur domain.UserRepository, logger *log.Entry, maxBulkOps int) (http.Handler, error) {
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}
	if maxBulkOps == 0 {
		return nil, errors.New("maxBulkOps must be greater than zero")
	}
	return handler{uc: uc, ur: ur, maxBulkOps: maxBulkOps, logger: logger}, nil
}
