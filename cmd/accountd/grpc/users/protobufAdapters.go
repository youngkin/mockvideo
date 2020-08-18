// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"github.com/youngkin/mockvideo/internal/domain"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
)

// UserToProtobuf converts a User domain object into a protobuf UserResponse
func UserToProtobuf(u *domain.User) *pb.UserResponse {
	return &pb.UserResponse{
		AccountID: int64(u.AccountID),
		HREF:      u.HREF,
		ID:        int64(u.ID),
		Name:      u.Name,
		EMail:     u.EMail,
		Role:      pb.RoleType(u.Role),
	}
}

// UserToProtobuf converts a Users domain object into a protobuf UsersResponse
func UsersToProtobuf(us *domain.Users) *pb.UsersResponse {
	pbUsers := pb.UsersResponse{}

	for _, u := range us.Users {
		ub := UserToProtobuf(u)
		pbUsers.Users = append(pbUsers.Users, ub)
	}
	return &pbUsers
}
