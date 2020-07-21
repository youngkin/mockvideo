// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package services

import (
	"github.com/juju/errors"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/constants"
	"github.com/youngkin/mockvideo/internal/domain"
	dom2 "github.com/youngkin/mockvideo/internal/domain"
)

// UserSvcInterface defines the operations to be supported by any types that provide
// the implementations of user related usecases
type UserSvcInterface interface {
	GetUsers() (*dom2.Users, error)
	GetUser(id int) (*dom2.User, error)
	CreateUser(user dom2.User) (id int, errCode constants.ErrCode, err error)
	UpdateUser(user dom2.User) (constants.ErrCode, error)
	DeleteUser(id int) (constants.ErrCode, error)
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

// GetUsers ...
func (us *UserSvc) GetUsers() (*dom2.Users, error) {
	return us.Repo.GetUsers()
}

// GetUser ...
func (us *UserSvc) GetUser(id int) (*dom2.User, error) {
	return us.Repo.GetUser(id)
}

// CreateUser ...
func (us *UserSvc) CreateUser(user dom2.User) (id int, errCode constants.ErrCode, err error) {
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

// UpdateUser ...
func (us *UserSvc) UpdateUser(user dom2.User) (constants.ErrCode, error) {
	return us.Repo.UpdateUser(user)
}

// DeleteUser ...
func (us *UserSvc) DeleteUser(id int) (constants.ErrCode, error) {
	return us.Repo.DeleteUser(id)
}
