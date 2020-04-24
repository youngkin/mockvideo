// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"fmt"
	"net/http"

	"github.com/youngkin/mockvideo/src/internal/platform/constants"
	"github.com/youngkin/mockvideo/src/internal/user"
)

// Response contains the results of in individual User request
type Response struct {
	HTTPStatus int               `json:"httpstatus"`
	ErrMsg     string            `json:"errmsg"`
	ErrReason  constants.ErrCode `json:"-"`
	User       user.User         `json:"user,omitempty"`
}

// BulkResponse contains the results of in bulk  User request
type BulkResponse struct {
	OverallStatus int        `json:"overallstatus"`
	Results       []Response `json:"results"`
}

// Request contains the information needed to process a request as well
// as capture to result of processing that request.
type Request struct {
	handler   handler
	ResponseC chan Response
	user      *user.User
	method    string
}

// BulkRequest contains a set of requests to be processed and the single channel to listen to for results
type BulkRequest struct {
	ResponseC chan Response
	Requests  []Request
}

// NewBulkRequest returns a Request. This is the only way to create a valid Request. The
// returned request contains a channel to listen on for concurrent request completion,
// and the individual user instances that are the target of the operation.
func NewBulkRequest(users user.Users, method string, handler handler) BulkRequest {
	// responseC must be a buffered channel of at least 1. This is required to handle a
	// potential race condition that occurs when the client 'Stop()'s a BulkPost while
	// one or more requests are being actively processed but not yet handled by the client.
	responseC := make(chan Response, len(users.Users))
	requests := []Request{}

	for _, u := range users.Users {
		rqst := Request{
			handler:   handler,
			ResponseC: responseC,
			user:      u,
			method:    method,
		}
		requests = append(requests, rqst)
	}
	return BulkRequest{ResponseC: responseC, Requests: requests}
}

// BulkProcesor supports (bulk) operations involving multiple users
type BulkProcesor struct {
	RequestC chan Request
	close    chan struct{}
}

// NewBulkProcessor returns a BulkProcessor which will support bulk user operations with a
// maximum number of concurrent requests limited by concurrencyLimit
func NewBulkProcessor(concurrencyLimit int) *BulkProcesor {
	bp := BulkProcesor{
		RequestC: make(chan Request, concurrencyLimit),
		close:    make(chan struct{}),
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
	// TODO: This doesn't quite implement a limited concurrency model. Need 'process()' to obtain/release a
	// TODO: space for another request
	for {
		select {
		case <-bp.close:
			return
		case rqst := <-bp.RequestC:
			rqst.handler.logger.Debugf("BulkProcessor received request: %+v", rqst)
			go bp.process(rqst)
		}
	}
}

func (bp BulkProcesor) process(rqst Request) {
	var r Response
	switch rqst.method {
	case http.MethodPost:
		rqst.handler.logger.Debugf("BulkProcessor processing POST request: %v+", rqst)
		r = rqst.handler.handlePostSingleUser(*rqst.user)
	case http.MethodPut:
		rqst.handler.logger.Debugf("BulkProcessor processing POST request: %v+", rqst)
		r = rqst.handler.handlePutSingleUser(*rqst.user)
	default:
		rqst.handler.logger.Debugf("BulkProcessor received unsupported HTTP method request: %v+", rqst)
		r = Response{
			ErrMsg:     fmt.Sprintf("Bulk %s HTTP method not supported", rqst.method),
			ErrReason:  constants.UserRqstErrorCode,
			HTTPStatus: http.StatusBadRequest,
			User:       *rqst.user,
		}
	}
	rqst.ResponseC <- r
	rqst.handler.logger.Debugf("BulkProcessor.process sent response: %v+", r)
}
