// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/cmd/accountd/services"
	"github.com/youngkin/mockvideo/internal/logging"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
)

// UserServer implements the gRPC functions required to provide access to user related services
type UserServer struct {
	userSvc    services.UserSvcInterface
	logger     *log.Entry
	maxBulkOps int
}

// GetUser returns the User identified by GetUserRqst.Id
func (s *UserServer) GetUser(ctx context.Context, rqst *pb.GetUserRqst) (*pb.UserResponse, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "GetUser",
		logging.UserID:  rqst.Id,
	}).Info("GetUser RPC request received")

	u, err := s.userSvc.GetUser(int(rqst.Id))
	if err != nil {
		return nil, fmt.Errorf("Error received when getting user %d. Wrapped error: %s", rqst.Id, err)
	}

	userPB := UserToProtobuf(u)
	return userPB, nil
}

// GetUsers returns all known users
func (s *UserServer) GetUsers(context.Context, *empty.Empty) (*pb.UsersResponse, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "GetUsers",
	}).Info("GetUsers RPC request received")

	users, err := s.userSvc.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("Error received when getting users. Wrapped error: %s", err)
	}

	usersPB := UsersToProtobuf(users)
	return usersPB, nil

	// TODO: Implement
	// return &pb.UsersResponse{
	// 	Users: []*pb.UserResponse{
	// 		{
	// 			AccountID: 100,
	// 			HREF:      "some ref",
	// 			ID:        100,
	// 			Name:      "Porgy",
	// 			EMail:     "tirebiter@somewhere.com",
	// 			Role:      0,
	// 			Password:  "nickdanger",
	// 		},
	// 		{
	// 			AccountID: 100,
	// 			HREF:      "some ref",
	// 			ID:        100,
	// 			Name:      "Porgy",
	// 			EMail:     "tirebiter@somewhere.com",
	// 			Role:      0,
	// 			Password:  "nickdanger",
	// 		},
	// 	},
	// }, nil
}

// NewUserServer returns a properly configured grpc Server
func NewUserServer(userSvc services.UserSvcInterface, logger *log.Entry, maxBulkOps int) (pb.UserServer, error) {
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}
	if maxBulkOps == 0 {
		return nil, errors.New("maxBulkOps must be greater than zero")
	}
	return &UserServer{userSvc: userSvc, maxBulkOps: maxBulkOps, logger: logger}, nil
}
