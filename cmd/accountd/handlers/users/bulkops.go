// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"fmt"
	"net/http"

	"github.com/youngkin/mockvideo/cmd/accountd/services"
	"github.com/youngkin/mockvideo/internal/constants"
	"github.com/youngkin/mockvideo/internal/domain"
)

// Response contains the results of in individual User request
type Response struct {
	HTTPStatus int               `json:"httpstatus"`
	ErrMsg     string            `json:"errmsg"`
	ErrReason  constants.ErrCode `json:"-"`
	User       domain.User       `json:"user,omitempty"`
}

// BulkResponse contains the results of in bulk  User request
type BulkResponse struct {
	OverallStatus int        `json:"overallstatus"`
	Results       []Response `json:"results"`
}

// Request contains the information needed to process a request as well
// as capture to result of processing that request.
type Request struct {
	userSvc   services.UserSvcInterface
	handler   handler
	ResponseC chan Response
	user      domain.User
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
func NewBulkRequest(users domain.Users, method string, userSvc services.UserSvcInterface) BulkRequest {
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

	r := Response{
		ErrMsg:    "",
		ErrReason: constants.UnknownErrorCode,
	}
	switch rqst.method {
	case http.MethodPost:
		// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor processing POST request: %v+", rqst)
		_, errCode, err := rqst.userSvc.CreateUser(rqst.user)
		if err != nil {
			r = Response{
				ErrMsg:     err.Error(),
				ErrReason:  errCode,
				HTTPStatus: http.StatusBadRequest,
				User:       rqst.user,
			}
		} else {
			r.HTTPStatus = http.StatusCreated
			r.User = rqst.user
		}
	case http.MethodPut:
		// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor processing POST request: %v+", rqst)
		errCode, err := rqst.userSvc.UpdateUser(rqst.user)
		if err != nil {
			r = Response{
				ErrMsg:     err.Error(),
				ErrReason:  errCode,
				HTTPStatus: http.StatusBadRequest,
				User:       rqst.user,
			}
		} else {
			r.HTTPStatus = http.StatusOK
			r.User = rqst.user
		}
	default:
		// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor received unsupported HTTP method request: %v+", rqst)
		r = Response{
			ErrMsg:     fmt.Sprintf("Bulk %s HTTP method not supported", rqst.method),
			ErrReason:  constants.UserRqstErrorCode,
			HTTPStatus: http.StatusBadRequest,
			User:       rqst.user,
		}
	}

	rqst.ResponseC <- r
	// TODO: FIX rqst.userSvc.logger.Debugf("BulkProcessor.process sent response: %v+", r)
}
