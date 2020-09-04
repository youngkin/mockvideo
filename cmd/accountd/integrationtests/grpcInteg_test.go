// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrationtests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
	"google.golang.org/grpc"
)

type CallType int

const (
	GETUSER CallType = iota
	GETUSERS
	CREATEUSER
	UPDATEUSER
	DELETEUSER
)

func (ct CallType) String() string {
	callTypes := []string{"GETUSER", "GETUSERS", "CREATEUSER", "UPDATEUSER", "DELETEUSER"}
	if int(ct) > len(callTypes)-1 {
		return "UNKNOWN"
	}
	return callTypes[ct]
}

// TestGetAllUsers attempts to retrieve users from a running accountd server
// connected to a real database.
func TestGetUsersGRPC(t *testing.T) {
	if protocol != "grpc" {
		return
	}

	// Takes a while for the accountd container to start
	time.Sleep(500 * time.Millisecond)

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:5000", opts)
	if err != nil {
		log.Fatalf("\terror %s attempting to connect to Accountd gRPC server", err)
	}
	defer cc.Close()

	client := pb.NewUserServerClient(cc)

	tcs := []struct {
		testName           string
		callType           CallType
		shouldPass         bool
		userID             *pb.UserID
		expectedHTTPStatus int
	}{
		{
			testName:   "testGetAllUsersSuccessGRPC",
			callType:   GETUSERS,
			shouldPass: true,
		},
		{
			testName:   "testGetOneUserSuccessGRPC",
			callType:   GETUSER,
			shouldPass: true,
			userID:     &pb.UserID{Id: 1},
		},
		{
			testName:   "testGetOneUserFailGRPC",
			callType:   GETUSER,
			shouldPass: false,
			userID:     &pb.UserID{Id: 1000},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			switch tc.callType {
			case GETUSERS:
				actual, err := client.GetUsers(context.Background(), &empty.Empty{})
				testPreconditions(t, actual, err, tc.shouldPass)
				if tc.shouldPass {
					if *update {
						updateGoldenFile(t, tc.testName, fmt.Sprintf("%v", actual.Users))
					}

					expected := readGoldenFile(t, tc.testName)

					if expected != fmt.Sprintf("%v", actual.Users) {
						t.Errorf("expected %s, got %s", expected, fmt.Sprintf("%v", actual.Users))
					}
				}
			case GETUSER:
				actual, err := client.GetUser(context.Background(), tc.userID)
				testPreconditions(t, actual, err, tc.shouldPass)
				if tc.shouldPass {
					if *update {
						updateGoldenFile(t, tc.testName, fmt.Sprintf("%s", actual))
					}

					expected := readGoldenFile(t, tc.testName)

					if expected != fmt.Sprintf("%s", actual) {
						t.Errorf("expected %s, got %s", expected, fmt.Sprintf("%s", actual))
					}
				}
			default:
				t.Errorf("unexpected callType %s provided", tc.callType)
			}
		})
	}
}

func testPreconditions(t *testing.T, actual interface{}, err error, shouldPass bool) {
	if err == nil && !shouldPass {
		t.Fatal("the expected error didn't occur")
	}
	if err != nil && shouldPass {
		t.Fatalf("an error '%s' was not expected calling GetUsers", err)
	}
	if actual == nil && shouldPass {
		t.Fatal("call to GetUsers unexpectedly returned a nil result")
	}
}
