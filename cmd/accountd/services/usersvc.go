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
	GetUsers() (*domain.Users, *mverr.MVError)
	GetUser(id int) (*domain.User, *mverr.MVError)
	CreateUser(user domain.User) (id int, err *mverr.MVError)
	UpdateUser(user domain.User) *mverr.MVError
	DeleteUser(id int) *mverr.MVError
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
func (us *UserSvc) GetUsers() (*domain.Users, *mverr.MVError) {
	users, err := us.Repo.GetUsers()

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	return users, nil
}

// GetUser retrieves a user from mySQL
func (us *UserSvc) GetUser(id int) (*domain.User, *mverr.MVError) {
	u, err := us.Repo.GetUser(id)

	if err != nil {
		us.logUserError(err)
		return nil, err
	}

	return u, nil
}

// CreateUser inserts a new User into mySQL
func (us *UserSvc) CreateUser(u domain.User) (id int, err *mverr.MVError) {
	id, err = us.Repo.CreateUser(u)
	if err != nil {
		us.logUserError(err)
		return 0, err
	}

	return id, err
}

// UpdateUser updates an existing user in mySQL
func (us *UserSvc) UpdateUser(user domain.User) *mverr.MVError {
	err := us.Repo.UpdateUser(user)
	if err != nil {
		us.logUserError(err)
		return err
	}

	return nil
}

// DeleteUser deletes an existing user from mySQL
func (us *UserSvc) DeleteUser(id int) *mverr.MVError {
	err := us.Repo.DeleteUser(id)
	if err != nil {
		us.logUserError(err)
		return err
	}

	return nil
}

func (us *UserSvc) logUserError(e *mverr.MVError) {
	us.Logger.WithFields(log.Fields{
		logging.ErrorCode:     e.ErrCode,
		logging.ErrorDetail:   e.WrappedErr,
		logging.MessageDetail: e.ErrMsg,
	}).Error(e.ErrDetail)

}
