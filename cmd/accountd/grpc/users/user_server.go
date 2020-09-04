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
func (s *UserServer) GetUser(ctx context.Context, rqst *pb.UserID) (*pb.User, error) {
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
func (s *UserServer) GetUsers(ctx context.Context, x *empty.Empty) (*pb.Users, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "GetUsers",
	}).Info("GetUsers RPC request received")

	users, err := s.userSvc.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("Error received when getting users. Wrapped error: %s", err)
	}

	usersPB := UsersToProtobuf(users)
	return usersPB, nil
}

// CreateUser creates a new User
func (s *UserServer) CreateUser(ctx context.Context, u *pb.User) (*pb.UserID, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc:   "CreateUser",
		logging.UserEMail: u.GetEMail(),
	}).Info("CreateUser RPC request received")

	id, err := s.userSvc.CreateUser(*ProtobufToUser(u))
	if err != nil {
		return nil, fmt.Errorf("Error received creating a new user. Wrapped error: %s", err)
	}

	userIDPB := pb.UserID{Id: int64(id)}
	return &userIDPB, nil
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, u *pb.User) (*empty.Empty, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc:   "UpdateUser",
		logging.UserID:    u.GetID(),
		logging.UserEMail: u.GetEMail(),
	}).Info("UpdateUser RPC request received")

	err := s.userSvc.UpdateUser(*ProtobufToUser(u))
	if err != nil {
		return nil, fmt.Errorf("Error received updating user %d with email %s. Wrapped error: %s", u.GetID(), u.GetEMail(), err)
	}

	return &empty.Empty{}, nil
}

// DeleteUser updates an existing user
func (s *UserServer) DeleteUser(ctx context.Context, id *pb.UserID) (*empty.Empty, error) {
	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "DeleteUser",
		logging.UserID:  id.GetId(),
	}).Info("DeleteUser RPC request received")

	err := s.userSvc.DeleteUser(int(id.GetId()))
	if err != nil {
		return nil, fmt.Errorf("Error received deleting user %d. Wrapped error: %s", id.GetId(), err)
	}

	return &empty.Empty{}, nil
}

// Health is used to determine the status or health of the service
func (s *UserServer) Health(ctx context.Context, _ *empty.Empty) (*pb.HealthMsg, error) {
	return &pb.HealthMsg{
		Status: "gRPC User Service is healthy",
	}, nil
}

// NewUserServer returns a properly configured grpc Server
func NewUserServer(userSvc services.UserSvcInterface, logger *log.Entry, maxBulkOps int) (pb.UserServerServer, error) {
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}
	if maxBulkOps == 0 {
		return nil, errors.New("maxBulkOps must be greater than zero")
	}
	return &UserServer{userSvc: userSvc, maxBulkOps: maxBulkOps, logger: logger}, nil
}
