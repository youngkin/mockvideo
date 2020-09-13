// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"fmt"

	"github.com/youngkin/mockvideo/cmd/accountd/internal/services"
	"github.com/youngkin/mockvideo/internal/domain"
)

// DomainUserToProtobuf converts a User domain object into a protobuf User
func DomainUserToProtobuf(u *domain.User) *User {
	return &User{
		AccountID: int64(u.AccountID),
		HREF:      u.HREF,
		ID:        int64(u.ID),
		Name:      u.Name,
		EMail:     u.EMail,
		Role:      RoleEnum(u.Role),
	}
}

// DomainUsersToProtobuf converts a Users domain object into a protobuf Users
func DomainUsersToProtobuf(us *domain.Users) *Users {
	pbUsers := Users{}

	for _, u := range us.Users {
		ub := DomainUserToProtobuf(u)
		pbUsers.Users = append(pbUsers.Users, ub)
	}
	return &pbUsers
}

// ProtobufToUser converts a User to a domwin.User
func ProtobufToUser(ub *User) (*domain.User, error) {
	u := &domain.User{
		AccountID: int(ub.AccountID),
		HREF:      ub.HREF,
		ID:        int(ub.GetID()),
		Name:      ub.Name,
		EMail:     ub.GetEMail(),
		Role:      domain.Role(ub.GetRole()),
		Password:  ub.GetPassword(),
	}

	err := u.ValidateUser()
	return u, err
}

// ProtobufToUsers converts a Users to a domain.User
func ProtobufToUsers(users *Users) (*domain.Users, error) {
	dUsers := domain.Users{}
	dUsers.Users = []*domain.User{}
	for _, u := range users.Users {
		du, err := ProtobufToUser(u)
		if err != nil {
			return nil, fmt.Errorf("error converting protobuf.User to domain.user: protobufUser: %v, error: %s", u, err)
		}
		dUsers.Users = append(dUsers.Users, du)
	}

	return &dUsers, nil
}

func statusToPBStatus(status services.Status) StatusEnum {
	var pbStatus StatusEnum

	switch status {
	case services.StatusBadRequest:
		pbStatus = StatusEnum_StatusBadRequest
	case services.StatusCreated:
		pbStatus = StatusEnum_StatusCreated
	case services.StatusNotFound:
		pbStatus = StatusEnum_StatusNotFound
	case services.StatusOK:
		pbStatus = StatusEnum_StatusOK
	case services.StatusServerError:
		pbStatus = StatusEnum_StatusServerError
	default:
		pbStatus = StatusEnum_StatusServerError
	}

	return pbStatus
}
