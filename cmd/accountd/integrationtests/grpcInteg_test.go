// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrationtests

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/youngkin/mockvideo/pkg/accountd"
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
				resp, err := client.GetUsers(context.Background(), &empty.Empty{})
				testPreconditions(t, resp, err, tc.shouldPass)
				if tc.shouldPass {
					actual, err := json.Marshal(resp)
					if err != nil {
						t.Fatalf("error '%s' was not expected while marshaling %v", err, resp)
					}
					if *update {
						updateGoldenFile(t, tc.testName, string(actual))
					}

					expected := readGoldenFile(t, tc.testName)

					if expected != string(actual) {
						t.Errorf("expected %s, got %s", expected, string(actual))
					}
				}
			case GETUSER:
				resp, err := client.GetUser(context.Background(), tc.userID)
				testPreconditions(t, resp, err, tc.shouldPass)
				if tc.shouldPass {
					actual, err := json.Marshal(resp)
					if err != nil {
						t.Fatalf("error '%s' was not expected while marshaling %v", err, resp)
					}
					if *update {
						updateGoldenFile(t, tc.testName, string(actual))
					}

					expected := readGoldenFile(t, tc.testName)

					if expected != string(actual) {
						t.Errorf("expected %s, got %s", expected, string(actual))
					}
				}
			default:
				t.Errorf("unexpected callType %s provided", tc.callType)
			}
		})
	}
}

func TestAddUpdateDeleteUser(t *testing.T) {
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
		testName   string
		shouldPass bool
		callType   CallType
		expectedID *accountd.UserID
		rqstData   *accountd.User
	}{
		{
			testName:   "testAddUserSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedID: &accountd.UserID{Id: 6},
			rqstData: &accountd.User{
				AccountID: 1,
				Name:      "Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      accountd.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testAddDuplicateUserFailure", //Dup email address
			shouldPass: false,
			callType:   CREATEUSER,
			expectedID: &accountd.UserID{},
			rqstData: &accountd.User{
				AccountID: 1,
				Name:      "Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      accountd.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testUpdateUserSuccess",
			shouldPass: true,
			callType:   UPDATEUSER,
			expectedID: &accountd.UserID{Id: 6},
			rqstData: &accountd.User{
				AccountID: 1,
				ID:        6,
				Name:      "Fleetwood Mac Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      accountd.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testUpdateNonExistingUserSuccess",
			shouldPass: false,
			callType:   UPDATEUSER,
			expectedID: &accountd.UserID{Id: 1000},
			rqstData: &accountd.User{
				AccountID: 1,
				ID:        1000,
				Name:      "Young Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      accountd.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testDeleteUserSuccess",
			shouldPass: true,
			callType:   DELETEUSER,
			expectedID: &accountd.UserID{Id: 6},
			rqstData:   nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			var id *accountd.UserID
			var err error
			switch tc.callType {
			case CREATEUSER:
				id, err = client.CreateUser(context.Background(), tc.rqstData)
			case UPDATEUSER:
				_, err = client.UpdateUser(context.Background(), tc.rqstData)
			case DELETEUSER:
				_, err = client.DeleteUser(context.Background(), tc.expectedID)
			}

			testPreconditions(t, id, err, tc.shouldPass)

			switch tc.callType {
			case DELETEUSER:
				return
			case UPDATEUSER:
				id = tc.expectedID
			}

			if tc.shouldPass {
				resp, err := client.GetUser(context.Background(), id)
				if err != nil {
					t.Fatalf("error '%s' was not expected calling accountd server", err)
				}

				actual, err := json.Marshal(resp)
				if err != nil {
					t.Fatalf("error '%s' was not expected while marshaling %v", err, resp)
				}

				if *update {
					updateGoldenFile(t, tc.testName, string(actual))
				}

				expected := readGoldenFile(t, tc.testName)

				if expected != string(actual) {
					t.Errorf("EXPECTED %s, GOT %s", expected, string(actual))
				}
			}

		})
	}
}

func TestBulkAddUpdateUser(t *testing.T) {
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
		testName              string
		shouldPass            bool
		callType              CallType
		expectedIDs           []*accountd.UserID
		rqstData              *accountd.Users
		expectedOverallStatus accountd.StatusEnum
		expectedResultStatus  []accountd.StatusEnum
	}{
		{
			testName:   "testAddUsersOneSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedIDs: []*accountd.UserID{
				{Id: 6},
			},
			rqstData: &accountd.Users{
				Users: []*pb.User{
					&accountd.User{
						AccountID: 1,
						Name:      "Peter Green",
						EMail:     "blackmagicwoman@gmail.com",
						Role:      accountd.RoleEnum_UNRESTRICTED,
						Password:  "lasttraintosanantone",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusCreated,
			expectedResultStatus:  []pb.StatusEnum{pb.StatusEnum_StatusCreated},
		},
		{
			testName:   "testAddUsersTwoSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedIDs: []*accountd.UserID{
				{Id: 7},
				{Id: 8},
			},
			rqstData: &accountd.Users{
				Users: []*pb.User{
					&accountd.User{
						AccountID: 1,
						Name:      "Brian Wilson",
						EMail:     "sloopjohnb@gmail.com",
						Role:      accountd.RoleEnum_UNRESTRICTED,
						Password:  "helpmerhonda",
					},
					&accountd.User{
						AccountID: 1,
						Name:      "Frank Zappa",
						EMail:     "apostrophe@gmail.com",
						Role:      accountd.RoleEnum_UNRESTRICTED,
						Password:  "donteatyellowsnow",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusCreated,
			expectedResultStatus:  []pb.StatusEnum{pb.StatusEnum_StatusCreated, pb.StatusEnum_StatusCreated},
		},
		{
			testName:   "testUpdateUsersSuccess",
			shouldPass: false,
			callType:   UPDATEUSER,
			expectedIDs: []*accountd.UserID{
				{Id: 6},
			},
			rqstData: &accountd.Users{
				Users: []*pb.User{
					&accountd.User{
						AccountID: 1,
						ID:        6,
						Name:      "Fleetwood Mac Peter Green",
						EMail:     "blackmagicwoman@gmail.com",
						Role:      accountd.RoleEnum_UNRESTRICTED,
						Password:  "lasttraintosanantone",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusOK,
			expectedResultStatus:  []pb.StatusEnum{pb.StatusEnum_StatusOK},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			var resp *pb.BulkResponse
			var err error
			switch tc.callType {
			case CREATEUSER:
				resp, err = client.CreateUsers(context.Background(), tc.rqstData)
			case UPDATEUSER:
				resp, err = client.UpdateUsers(context.Background(), tc.rqstData)
			}

			testPreconditions(t, resp, err, tc.shouldPass)

			if tc.shouldPass {
				if resp.OverallStatus != tc.expectedOverallStatus {
					t.Errorf("OverallStatus: EXPECTED %d, GOT %d", tc.expectedOverallStatus, resp.OverallStatus)
				}
				for i, result := range resp.Response {
					if result.Status != tc.expectedResultStatus[i] {
						t.Errorf("Status: EXPECTED %d, GOT %d", tc.expectedResultStatus[i], result.Status)
					}
					u, err := client.GetUser(context.Background(), result.UserID)
					if err != nil {
						t.Fatalf("error '%s' was not expected calling accountd server", err)
					}

					actual, err := json.Marshal(u)
					if err != nil {
						t.Fatalf("error '%s' was not expected while marshaling %v", err, resp)
					}

					if *update {
						updateGoldenFile(t, tc.testName+tc.rqstData.Users[i].EMail, string(actual))
					}

					expected := readGoldenFile(t, tc.testName+tc.rqstData.Users[i].EMail)

					if expected != string(actual) {
						t.Errorf("EXPECTED %s, GOT %s", expected, string(actual))
					}
				}
			}

		})
	}
}

func testPreconditions(t *testing.T, actual interface{}, err error, shouldPass bool) {
	if err == nil && !shouldPass {
		t.Fatal("the expected error didn't occur")
	}
	if err != nil && shouldPass {
		t.Fatalf("an error '%s' was not expected", err)
	}
	if actual == nil && shouldPass {
		t.Fatal("gRPC call unexpectedly returned a nil result")
	}
}
