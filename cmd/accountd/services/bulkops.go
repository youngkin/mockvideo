// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package services

import (
	"fmt"

	"github.com/youngkin/mockvideo/internal/domain"
	"github.com/youngkin/mockvideo/internal/errors"
)

// RqstType is used to indicate what kind of request is being made
type RqstType int

const (
	// CREATE indicates that a User is to be created
	CREATE RqstType = iota
	// UPDATE indicates that a User is to be updated
	UPDATE
	// READ indicates a User is to be retrieved
	READ
	// DELETE indicates a User is to be deleted
	DELETE
)

// RqstTypeName maps a specific RqstType value to a descriptive string
var RqstTypeName = map[RqstType]string{
	CREATE: "CREATE",
	UPDATE: "UPDATE",
	READ:   "READ",
	DELETE: "DELETE",
}

// Status indicates the result of a bulk operation
type Status int

const (
	// StatusBadRequest indicates that the client submitted an invalid request
	StatusBadRequest Status = iota
	// StatusOK indicates the request completed successfully
	StatusOK
	// StatusCreated indicates that the requested resource was created
	StatusCreated
	// StatusConflict indicates that one or more of a set of bulk requests failed
	StatusConflict
	// StatusServerError indicates that the server encountered an error while servicing the request
	StatusServerError
	// StatusNotFound indicates the requested resource does not exist
	StatusNotFound
)

// StatusTypeName maps a specific Status value to a descriptive string
var StatusTypeName = map[Status]string{
	StatusBadRequest:  "StatusBadRequest",
	StatusOK:          "StatusOK",
	StatusCreated:     "StatusCreated",
	StatusConflict:    "StatusConflict",
	StatusServerError: "StatusServerError",
	StatusNotFound:    "StatusNotFound",
}

// Response contains the results of in individual User request
type Response struct {
	Status    Status         `json:"status"`
	ErrMsg    string         `json:"errmsg"`
	ErrReason errors.ErrCode `json:"-"`
	User      domain.User    `json:"user,omitempty"`
}

// BulkResponse contains the results of in bulk  User request
type BulkResponse struct {
	OverallStatus Status     `json:"overallstatus"`
	Results       []Response `json:"results"`
}

// Request contains the information needed to process a request as well
// as capture to result of processing that request.
type Request struct {
	userSvc   UserSvcInterface
	ResponseC chan Response
	user      domain.User
	rqstType  RqstType
}

// BulkRequest contains a set of requests to be processed and the single channel to listen to for results
type BulkRequest struct {
	ResponseC chan Response
	Requests  []Request
}

// NewBulkRequest returns a Request. This is the only way to create a valid Request. The
// returned request contains a channel to listen on for concurrent request completion,
// and the individual user instances that are the target of the operation.
func NewBulkRequest(users domain.Users, rqstType RqstType, userSvc UserSvcInterface) BulkRequest {
	// responseC must be a buffered channel of at least 1. This is required to handle a
	// potential race condition that occurs when the client 'Stop()'s a BulkPost while
	// one or more requests are being actively processed but not yet handled by the client.
	responseC := make(chan Response, len(users.Users))
	requests := []Request{}

	for _, u := range users.Users {
		rqst := Request{
			userSvc:   userSvc,
			ResponseC: responseC,
			user:      *u,
			rqstType:  rqstType,
		}
		requests = append(requests, rqst)
	}
	return BulkRequest{ResponseC: responseC, Requests: requests}
}

// BulkProcesor supports (bulk) operations involving multiple users
type BulkProcesor struct {
	RequestC chan Request
	close    chan struct{}
	// limitRqstC acts as a semaphore to limit the number of concurrent requests.
	// It can accept up to 10 messages before blocking. To use, pass a message into
	// the channel to indicate a request for resources; to free up resources accept
	// a message from the channel. These operations should surround any calls requiring
	// resources (i.e., processing Request-s on 'RequestC'.
	limitRqstsC chan struct{}
}

// NewBulkProcessor returns a BulkProcessor which will support bulk user operations with a
// maximum number of concurrent requests limited by concurrencyLimit
func NewBulkProcessor(concurrencyLimit int) *BulkProcesor {
	bp := BulkProcesor{
		RequestC:    make(chan Request, concurrencyLimit),
		close:       make(chan struct{}),
		limitRqstsC: make(chan struct{}, concurrencyLimit),
	}
	go bp.loop()
	return &bp
}

// Stop will halt the BulkProcesor background goroutine
func (bp BulkProcesor) Stop() {
	bp.close <- struct{}{}
	return
}

func (bp BulkProcesor) loop() {
	for {
		select {
		case <-bp.close:
			return
		case rqst := <-bp.RequestC:
			// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor received request: %+v", rqst)
			// Request for resources prior to launching 'process()' as a goroutine.
			// This limits the number of running goroutines.
			bp.limitRqstsC <- struct{}{}
			go bp.process(rqst)
		}
	}
}

func (bp BulkProcesor) process(rqst Request) {
	// After called functions return, accept from the 'limitRqstC'
	// to free up resources for another request.
	releaseResource := func() {
		<-bp.limitRqstsC
	}
	defer releaseResource()

	r := Response{}

	switch rqst.rqstType {
	case CREATE:
		// TODO: Un/comment as needed
		// fmt.Printf("\n\n ====================> BulkProcessor processing CREATE request: %+v\n", rqst)
		id, err := rqst.userSvc.CreateUser(rqst.user)
		if err != nil {
			r = Response{
				ErrMsg:    err.ErrMsg,
				ErrReason: err.ErrCode,
				Status:    StatusBadRequest,
				User:      rqst.user,
			}
		} else {
			r.Status = StatusCreated
			r.User = rqst.user
			r.User.ID = id
		}
	case UPDATE:
		// TODO: Un/comment as needed
		// fmt.Printf("\n\n ====================> BulkProcessor processing UPDATE request: %+v\n", rqst)
		err := rqst.userSvc.UpdateUser(rqst.user)
		if err != nil {
			r = Response{
				ErrMsg:    err.ErrMsg,
				ErrReason: err.ErrCode,
				Status:    StatusBadRequest,
				User:      rqst.user,
			}
		} else {
			r.Status = StatusOK
			r.User = rqst.user
		}
	default:
		// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor received unsupported HTTP method request: %+v", rqst)
		r = Response{
			ErrMsg:    fmt.Sprintf("Bulk RequestType %s not supported\n", RqstTypeName[rqst.rqstType]),
			ErrReason: errors.UserRqstErrorCode,
			Status:    StatusBadRequest,
			User:      rqst.user,
		}
	}

	rqst.ResponseC <- r
	// TODO: Un/comment as needed
	// fmt.Printf("\n\n ====================> BulkProcessor.process sent response: %+v\n\n", r)
}
