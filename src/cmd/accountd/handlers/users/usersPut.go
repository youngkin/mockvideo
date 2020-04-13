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
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/src/internal/platform/constants"
	"github.com/youngkin/mockvideo/src/internal/user"
)

// TODO:
//	1.	Add context deadline to DB requests

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// parseRqst() logs parsing errors, no need to log again
	u, pathNodes, err := h.parseRqst(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		userRqstDur.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	if len(pathNodes) != 2 {
		errMsg := fmt.Sprintf("expecting resource path like /users/{id}, got %+v", pathNodes)
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

	errCode, err := user.UpdateUser(h.db, u)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errCode == constants.DBInvalidRequestCode {
			httpStatus = http.StatusBadRequest
		}
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(errCode)
		w.WriteHeader(httpStatus)
		userRqstDur.WithLabelValues(strconv.Itoa(httpStatus)).Observe(float64(time.Since(start)) / float64(time.Second))
		return
	}

	w.WriteHeader(http.StatusOK)
	userRqstDur.WithLabelValues(strconv.Itoa(http.StatusCreated)).Observe(float64(time.Since(start)) / float64(time.Second))
}
