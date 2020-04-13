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

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/src/internal/platform/constants"
	"github.com/youngkin/mockvideo/src/internal/user"
)

// TODO:
//	1.	Add context deadline to DB requests

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	completeRequest := func(httpStatus int) {
		w.WriteHeader(httpStatus)
		userRqstDur.WithLabelValues(strconv.Itoa(httpStatus)).
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
		completeRequest(http.StatusBadRequest)
		return
	}

	var (
		payload   interface{}
		errReason constants.ErrCode
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
		}).Error(constants.UserRqstError)
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
	case *user.User:
		if p == nil {
			custFound = false
		}
	case *user.Users:
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
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		completeRequest(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)

	userRqstDur.WithLabelValues(strconv.Itoa(http.StatusFound)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func (h handler) handleGetUsers(path string) (interface{}, error) {
	usrs, err := user.GetAllUsers(h.db)
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
func (h handler) handleGetOneUser(path string, pathNodes []string) (cust interface{}, errReason constants.ErrCode, err error) {
	if len(pathNodes) > 1 {
		err := errors.Errorf(("expected 1 pathNode, got %d"), len(pathNodes))
		return nil, constants.MalformedURLErrorCode, err
	}

	id, err := strconv.Atoi(pathNodes[0])
	if err != nil {
		err := errors.Annotate(err, fmt.Sprintf("expected numeric pathNode, got %+v", id))
		return nil, constants.MalformedURLErrorCode, err
	}

	c, err := user.GetUser(h.db, id)
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
