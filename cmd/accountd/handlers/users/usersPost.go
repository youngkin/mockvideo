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
	"time"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/platform/constants"
	"github.com/youngkin/mockvideo/internal/user"
)

// TODO:
//	1.	Add context deadline to DB requests

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	h.logRqstRcvd(r)
	start := time.Now()

	//
	// Get user out of request body and validate
	//
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if user sends extra data
	user := user.User{}
	err := d.Decode(&user)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)
		w.WriteHeader(http.StatusBadRequest)
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Warn(constants.JSONDecodingError)
	}

	// Expecting t URL.Path '/users'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if user.ID != 0 { // User ID must *NOT* be populated (i.e., with a non-zero value) on an insert
		errMsg := fmt.Sprintf("expected User.ID > 0, got User.ID = %d", user.ID)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.InvalidInsertErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.InvalidInsertError)
		w.WriteHeader(http.StatusBadRequest)
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
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
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	userID, errReason, err := h.insertUser(user)
	if err != nil {
		status := http.StatusInternalServerError
		if errReason == constants.DBInsertDuplicateUserErrorCode {
			// Invalid to insert a duplicate user, this is a client error hence the StatusBadRequest
			status = http.StatusBadRequest
		}
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.DBUpSertErrorCode,
			constants.HTTPStatus:  status,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.DBUpSertError)
		w.WriteHeader(status)
		userRqstDur.WithLabelValues(strconv.Itoa(status)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	w.Header().Add("Location", fmt.Sprintf("/users/%d", userID))
	w.WriteHeader(http.StatusCreated)

	userRqstDur.WithLabelValues(strconv.Itoa(http.StatusCreated)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) insertUser(u user.User) (int64, constants.ErrCode, error) {
	id, errReason, err := user.InsertUser(h.db, u)
	if err != nil {
		return -1, errReason, errors.Annotate(err, "error inserting user")
	}
	return id, errReason, nil
}
