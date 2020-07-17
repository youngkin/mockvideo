// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package usecases

import (
	"errors"

	"github.com/youngkin/mockvideo/src/domain"
)

// UserUC defines the operations to be supported by any types that provide
// the implementations of user related usecases
type UserUC interface {
}

// UserUseCase provides the capability needed to interact with application
// usecases related to users
type UserUseCase struct {
	ur domain.UserRepository
}

// NewUserUseCase returns a new instance that handles application usecases related to users
func NewUserUseCase(ur domain.UserRepository) (UserUC, error) {
	if ur == nil {
		return nil, errors.New("non-nil *domain.UserRepository required")
	}
	return UserUseCase{ur: ur}, nil
}
