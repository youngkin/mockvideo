// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package services

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/domain"
	mverr "github.com/youngkin/mockvideo/internal/errors"
	"github.com/youngkin/mockvideo/internal/logging"
)

// UserSvcInterface defines the operations to be supported by any types that provide
// the implementations of user related usecases
// TODO: This exactly matches the UserRepository interface. This smells.
type UserSvcInterface interface {
	GetUsers() (*domain.Users, error)
	GetUser(id int) (*domain.User, error)
	CreateUser(user domain.User) (id int, errCode mverr.ErrCode, err error)
	UpdateUser(user domain.User) (mverr.ErrCode, error)
	DeleteUser(id int) (mverr.ErrCode, error)
}

// UserSvc provides the capability needed to interact with application
// usecases related to users
type UserSvc struct {
	Repo   domain.UserRepository
	Logger *log.Entry
}

// NewUserSvc returns a new instance that handles application usecases related to users
func NewUserSvc(ur domain.UserRepository, logger *log.Entry) (*UserSvc, error) {
	if ur == nil {
		return nil, errors.New("non-nil *domain.UserRepository required")
	}
	return &UserSvc{Repo: ur, Logger: logger}, nil
}

// GetUsers retrieves all Users from mySQL
func (us *UserSvc) GetUsers() (*domain.Users, error) {
	users, err := us.Repo.GetUsers()

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	return users, nil
}

// GetUser retrieves a user from mySQL
func (us *UserSvc) GetUser(id int) (*domain.User, error) {
	u, err := us.Repo.GetUser(id)

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	return u, nil
}

// CreateUser inserts a new User into mySQL
func (us *UserSvc) CreateUser(user domain.User) (id int, errCode mverr.ErrCode, err error) {
	u := domain.User{
		AccountID: user.AccountID,
		EMail:     user.EMail,
		Name:      user.Name,
		Role:      user.Role,
		Password:  user.Password,
	}
	id, errCode, err = us.Repo.CreateUser(u)
	return id, errCode, err
}

// UpdateUser updates an existing user in mySQL
func (us *UserSvc) UpdateUser(user domain.User) (mverr.ErrCode, error) {
	return us.Repo.UpdateUser(user)
}

// DeleteUser deletes an existing user from mySQL
func (us *UserSvc) DeleteUser(id int) (mverr.ErrCode, error) {
	return us.Repo.DeleteUser(id)
}

func (us *UserSvc) logUserError(e error) {
	var eu mverr.UserRqstError

	if errors.As(e, &eu) {
		us.Logger.WithFields(log.Fields{
			logging.ErrorCode:   eu.ErrCode,
			logging.ErrorDetail: eu.Error(),
		}).Error(eu.ErrMsg)

		return
	}

	us.Logger.WithFields(log.Fields{
		logging.ErrorCode:   mverr.UnknownErrorCode,
		logging.ErrorDetail: e.Error(),
	}).Error("unexpected error occurred in UserSvc, check ErrorDetail field for more info")
}
