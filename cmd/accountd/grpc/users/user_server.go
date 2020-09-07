// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/cmd/accountd/services"
	mverr "github.com/youngkin/mockvideo/internal/errors"
	"github.com/youngkin/mockvideo/internal/logging"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
)

const rqstStatus = "rqstStatus"

// UserRqstDur is used to capture the length of HTTP requests
var UserRqstDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "mockvideo",
	Subsystem: "user",
	Name:      "user_request_duration_seconds",
	Help:      "user request duration distribution in seconds",
	// Buckets:   prometheus.ExponentialBuckets(0.005, 1.1, 40),
	Buckets: prometheus.LinearBuckets(0.001, .004, 50),
}, []string{rqstStatus})

// UserServer implements the gRPC functions required to provide access to user related services
type UserServer struct {
	userSvc services.UserSvcInterface
	logger  *log.Entry
}

// GetUser returns the User identified by GetUserRqst.Id
func (s *UserServer) GetUser(ctx context.Context, rqst *pb.UserID) (*pb.User, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "GetUser",
		logging.UserID:  rqst.Id,
	}).Info("GetUser RPC request received")

	u, err := s.userSvc.GetUser(int(rqst.Id))
	if err != nil {
		UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusServerError]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, fmt.Errorf("Error received when getting user %d. Wrapped error: %s", rqst.Id, err)
	}

	if u == nil {
		UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusNotFound]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, nil
	}

	userPB := DomainUserToProtobuf(u)

	UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusOK]).Observe(float64(time.Since(start)) / float64(time.Second))

	return userPB, nil
}

// GetUsers returns all known users
func (s *UserServer) GetUsers(ctx context.Context, x *empty.Empty) (*pb.Users, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "GetUsers",
	}).Info("GetUsers RPC request received")

	users, err := s.userSvc.GetUsers()
	if err != nil {
		UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusServerError]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, fmt.Errorf("Error received when getting users. Wrapped error: %s", err)
	}

	usersPB := DomainUsersToProtobuf(users)

	UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusOK]).Observe(float64(time.Since(start)) / float64(time.Second))

	return usersPB, nil
}

// CreateUser creates a new User
func (s *UserServer) CreateUser(ctx context.Context, u *pb.User) (*pb.UserID, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc:   "CreateUser",
		logging.UserEMail: u.GetEMail(),
	}).Info("CreateUser RPC request received")

	du, err := ProtobufToUser(u)
	if err != nil {
		return nil, fmt.Errorf("invalid protobuf.User value provided: Error: %s", err)
	}
	id, mvErr := s.userSvc.CreateUser(*du)
	if mvErr != nil {
		status := services.StatusServerError
		switch mvErr.ErrCode {
		case mverr.DBInsertDuplicateUserErrorCode:
			status = services.StatusBadRequest
		case mverr.UserValidationErrorCode:
			status = services.StatusBadRequest
		case mverr.DBUpSertErrorCode:
			status = services.StatusServerError
		default:
			status = services.StatusServerError
		}
		UserRqstDur.WithLabelValues(services.StatusTypeName[status]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, fmt.Errorf("Error received creating a new user. Wrapped error: %s", mvErr)
	}

	userIDPB := pb.UserID{Id: int64(id)}

	UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusCreated]).Observe(float64(time.Since(start)) / float64(time.Second))

	return &userIDPB, nil
}

// CreateUsers creates users from the provided 'users' parameter
func (s *UserServer) CreateUsers(ctx context.Context, users *pb.Users) (*pb.BulkResponse, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "CreateUsers",
	}).Info("CreateUsers RPC request received")
	for _, u := range users.Users {
		s.logger.WithFields(log.Fields{
			logging.RPCFunc:   "CreateUsers",
			logging.UserEMail: u.GetEMail(),
		}).Info("CreateUsers RPC request received")
	}

	du, err := ProtobufToUsers(users)
	if err != nil {
		return nil, fmt.Errorf("invalid protobuf.User value provided: Error: %s", err)
	}

	responses, mvErr := s.userSvc.CreateUsers(*du)

	bulkResponse := pb.BulkResponse{OverallStatus: statusToPBStatus(responses.OverallStatus)}
	for _, result := range responses.Results {
		response := pb.Response{
			Status:    statusToPBStatus(result.Status),
			ErrMsg:    result.ErrMsg,
			ErrReason: int64(result.ErrReason),
			UserID: &pb.UserID{
				Id: int64(result.User.ID),
			},
		}
		bulkResponse.Response = append(bulkResponse.Response, &response)
	}

	var retErr error
	if mvErr != nil {
		retErr = fmt.Errorf("Error received creating a new user. Wrapped error: %s", mvErr)
	}

	UserRqstDur.WithLabelValues(services.StatusTypeName[responses.OverallStatus]).Observe(float64(time.Since(start)) / float64(time.Second))

	// TODO: Un/comment as needed
	fmt.Printf("\n\n ====================> pb.BulkResponse: %+v\n\n", &bulkResponse)

	return &bulkResponse, retErr
}

// UpdateUsers updates the set of users provided in the 'users' parameter
func (s *UserServer) UpdateUsers(ctx context.Context, users *pb.Users) (*pb.BulkResponse, error) {
	// start := time.Now()

	return &pb.BulkResponse{}, errors.New("UpdateUsers() Not Implemented")
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, u *pb.User) (*empty.Empty, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc:   "UpdateUser",
		logging.UserID:    u.GetID(),
		logging.UserEMail: u.GetEMail(),
	}).Info("UpdateUser RPC request received")

	du, err := ProtobufToUser(u)
	if err != nil {
		UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusBadRequest]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, err
	}

	var retErr error
	upErr := s.userSvc.UpdateUser(*du)
	if upErr != nil {
		status := services.StatusServerError
		if upErr.ErrCode == mverr.DBNoUserErrorCode {
			status = services.StatusBadRequest
		}
		UserRqstDur.WithLabelValues(services.StatusTypeName[status]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, fmt.Errorf("error received updating user %d with email %s. Wrapped error: %s", u.GetID(), u.GetEMail(), err)
	}

	UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusOK]).Observe(float64(time.Since(start)) / float64(time.Second))
	return &empty.Empty{}, retErr
}

// DeleteUser updates an existing user
func (s *UserServer) DeleteUser(ctx context.Context, id *pb.UserID) (*empty.Empty, error) {
	start := time.Now()

	s.logger.WithFields(log.Fields{
		logging.RPCFunc: "DeleteUser",
		logging.UserID:  id.GetId(),
	}).Info("DeleteUser RPC request received")

	err := s.userSvc.DeleteUser(int(id.GetId()))
	if err != nil {
		UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusServerError]).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, fmt.Errorf("error received deleting user %d", id.GetId())
	}

	UserRqstDur.WithLabelValues(services.StatusTypeName[services.StatusOK]).Observe(float64(time.Since(start)) / float64(time.Second))
	return &empty.Empty{}, nil
}

// Health is used to determine the status or health of the service
func (s *UserServer) Health(ctx context.Context, _ *empty.Empty) (*pb.HealthMsg, error) {
	return &pb.HealthMsg{
		Status: "gRPC User Service is healthy",
	}, nil
}

// NewUserServer returns a properly configured grpc Server
func NewUserServer(userSvc services.UserSvcInterface, logger *log.Entry) (pb.UserServerServer, error) {
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}
	return &UserServer{userSvc: userSvc, logger: logger}, nil
}
