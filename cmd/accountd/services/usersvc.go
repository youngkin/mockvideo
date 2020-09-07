// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package services

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/domain"
	mverr "github.com/youngkin/mockvideo/internal/errors"
	"github.com/youngkin/mockvideo/internal/logging"
)

// UserSvcInterface defines the operations to be supported by any types that provide
// the implementations of user related usecases
// TODO: This exactly matches the UserRepository interface. This smells.
type UserSvcInterface interface {
	GetUsers() (*domain.Users, *mverr.MVError)
	GetUser(id int) (*domain.User, *mverr.MVError)
	CreateUser(user domain.User) (id int, err *mverr.MVError)
	CreateUsers(users domain.Users) (bulkResponse *BulkResponse, err *mverr.MVError)
	UpdateUser(user domain.User) *mverr.MVError
	UpdateUsers(users domain.Users) (bulkResponse *BulkResponse, err *mverr.MVError)
	DeleteUser(id int) *mverr.MVError
}

// UserSvc provides the capability needed to interact with application
// usecases related to users
type UserSvc struct {
	repo       domain.UserRepository
	logger     *log.Entry
	maxBulkOps int
}

// NewUserSvc returns a new instance that handles application usecases related to users.
// 'ur' and 'logger' must be non-nil. 'maxBulkOps' must be greater than 0.
func NewUserSvc(ur domain.UserRepository, logger *log.Entry, maxBulkOps int) (*UserSvc, error) {
	if ur == nil {
		return nil, errors.New("non-nil *domain.UserRepository required")
	}
	if logger == nil {
		return nil, errors.New("non-nil *log.Entry required")
	}
	if maxBulkOps < 1 {
		return nil, errors.New("maxBulkOps must be greater than 0")
	}
	return &UserSvc{repo: ur, logger: logger, maxBulkOps: maxBulkOps}, nil
}

// GetUsers retrieves all Users from the database
func (us *UserSvc) GetUsers() (*domain.Users, *mverr.MVError) {
	users, err := us.repo.GetUsers()

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	return users, nil
}

// GetUser retrieves a user from the database
func (us *UserSvc) GetUser(id int) (*domain.User, *mverr.MVError) {
	u, err := us.repo.GetUser(id)

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	if u == nil {
		err := mverr.MVError{
			ErrCode:   mverr.DBNoUserErrorCode,
			ErrDetail: fmt.Sprintf("User %d not found", id),
			ErrMsg:    mverr.DBNoUserErrorMsg,
		}
		us.logUserError(&err)
		return nil, &err
	}
	return u, nil
}

// CreateUser inserts a new User into the database
func (us *UserSvc) CreateUser(u domain.User) (id int, err *mverr.MVError) {
	id, err = us.repo.CreateUser(u)
	if err != nil {
		us.logUserError(err)
		return 0, err
	}

	return id, err
}

// CreateUsers inserts a group new Users into the database
func (us *UserSvc) CreateUsers(users domain.Users) (bulkResponse *BulkResponse, err *mverr.MVError) {
	responses := us.handleRqstMultipleUsers(time.Now(), users, CREATE)

	for _, result := range responses.Results {
		if result.ErrReason != mverr.NoErrorCode {
			us.logger.WithFields(log.Fields{
				logging.ErrorCode:   result.ErrReason,
				logging.Status:      result.Status,
				logging.ErrorDetail: fmt.Sprintf("error creating user: Name: %s, email: %s", result.User.Name, result.User.EMail),
			}).Errorf(result.ErrMsg)
		}
	}

	if responses.OverallStatus != StatusCreated {
		err = &mverr.MVError{
			ErrCode: mverr.BulkRequestErrorCode,
			ErrMsg:  mverr.BulkRequestErrorMsg,
			WrappedErr: fmt.Errorf("part or all of a bulk request failed, overall request status %s",
				StatusTypeName[responses.OverallStatus]),
		}
		us.logUserError(err)
		return nil, err
	}

	// TODO: Un/comment as needed
	// fmt.Printf("\n\n ====================> BulkResponse: %+v\n\n", responses)

	return responses, nil
}

// UpdateUser updates an existing user in the database
func (us *UserSvc) UpdateUser(user domain.User) *mverr.MVError {
	err := us.repo.UpdateUser(user)
	if err != nil {
		us.logUserError(err)
		return err
	}

	return nil
}

// UpdateUsers updates a group existing Users in the database
func (us *UserSvc) UpdateUsers(users domain.Users) (bulkResponse *BulkResponse, err *mverr.MVError) {
	// TODO: Implement
	return nil, &mverr.MVError{
		ErrCode:    mverr.BulkRequestErrorCode,
		ErrMsg:     "NOT_IMPLEMENTED",
		ErrDetail:  "UpdateUsers() not implemented",
		WrappedErr: nil,
	}
}

// DeleteUser deletes an existing user from the database
func (us *UserSvc) DeleteUser(id int) *mverr.MVError {
	err := us.repo.DeleteUser(id)
	if err != nil {
		us.logUserError(err)
		return err
	}

	return nil
}

func (us *UserSvc) logUserError(e *mverr.MVError) {
	us.logger.WithFields(log.Fields{
		logging.ErrorCode:    e.ErrCode,
		logging.ErrorDetail:  e.ErrDetail,
		logging.WrappedError: e.WrappedErr,
	}).Error(e.ErrMsg)

}

func (us *UserSvc) handleRqstMultipleUsers(start time.Time, users domain.Users, rqstType RqstType) *BulkResponse {
	us.logger.Debugf("handleRqstMultipleUsers for %d", rqstType)
	bp := NewBulkProcessor(us.maxBulkOps)
	defer bp.Stop()

	br := NewBulkRequest(users, rqstType, us)
	rqstCompleteC := make(chan Response)
	numUsers := len(users.Users)

	for i := 0; i < numUsers; i++ {
		go us.handleConcurrentRqst(br.Requests[i], bp.RequestC, rqstCompleteC)
	}

	us.logger.Debugf("handleRqstMultipleUsers: Started %d goroutines for RqstType %d", numUsers, rqstType)
	responses := BulkResponse{}
	overallStatus := StatusOK
	if rqstType == CREATE {
		overallStatus = StatusCreated
	}

	for i := 0; i < numUsers; i++ {
		resp := <-rqstCompleteC
		responses.Results = append(responses.Results, resp)
		if resp.Status != StatusCreated && resp.Status != StatusOK {
			overallStatus = StatusConflict
		}
	}

	responses.OverallStatus = overallStatus
	return &responses
}

func (us *UserSvc) handleConcurrentRqst(rqst Request, rqstC chan Request, rqstCompC chan Response) {
	us.logger.Debugf("handleConcurrentRqst: request %+v", rqst)
	rqstC <- rqst
	us.logger.Debug("handleConcurrentRqst: request sent")
	resp := <-rqst.ResponseC
	us.logger.Debugf("handleConcurrentRqst: response %+v received", rqst)
	rqstCompC <- resp
	us.logger.Debug("handleConcurrentRqst: response sent")
}
