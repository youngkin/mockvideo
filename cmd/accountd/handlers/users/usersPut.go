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
6.	Use of Prometheus to capture metrics
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/platform/constants"
	"github.com/youngkin/mockvideo/internal/user"
)

// TODO:
//	1.	Add context deadline to DB requests

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
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

	//
	// TODO: Write user to DB and get ID
	// Is this an insert or update?
	//
	// Expecting a URL.Path like '/users' or '/users/{id}'
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

	var userID int64
	if len(pathNodes) == 1 {
		userID, err = h.insertUser(user)
	} else {
		// TODO: PROHIBIT UPDATES ON POST, REQUIRE PUT
		userID, err = h.updateUser(user)
	}
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.DBUpSertErrorCode,
			constants.HTTPStatus:  http.StatusInternalServerError,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.DBUpSertError)
		w.WriteHeader(http.StatusInternalServerError)
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusInternalServerError)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	//
	// Wrap up request
	//
	w.Header().Add("Location", fmt.Sprintf("/users/%d", userID))
	w.WriteHeader(http.StatusCreated)

	userRqstDur.WithLabelValues(strconv.Itoa(http.StatusCreated)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) updateUser(u user.User) (int64, error) {
	// TODO: Implement
	return int64(u.ID), nil
}
