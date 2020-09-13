// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrationtests

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/youngkin/mockvideo/cmd/accountd/grpc/users"
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

func TestAddUpdateDeleteUserGRPC(t *testing.T) {
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
		expectedID *pb.UserID
		rqstData   *pb.User
	}{
		{
			testName:   "testAddUserSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedID: &pb.UserID{Id: 6},
			rqstData: &pb.User{
				AccountID: 1,
				Name:      "Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      pb.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testAddDuplicateUserFailure", //Dup email address
			shouldPass: false,
			callType:   CREATEUSER,
			expectedID: &pb.UserID{},
			rqstData: &pb.User{
				AccountID: 1,
				Name:      "Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      pb.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testUpdateUserSuccess",
			shouldPass: true,
			callType:   UPDATEUSER,
			expectedID: &pb.UserID{Id: 6},
			rqstData: &pb.User{
				AccountID: 1,
				ID:        6,
				Name:      "Fleetwood Mac Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      pb.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testUpdateNonExistingUserSuccess",
			shouldPass: false,
			callType:   UPDATEUSER,
			expectedID: &pb.UserID{Id: 1000},
			rqstData: &pb.User{
				AccountID: 1,
				ID:        1000,
				Name:      "Young Peter Green",
				EMail:     "blackmagicwoman@gmail.com",
				Role:      pb.RoleEnum_UNRESTRICTED,
				Password:  "lasttraintosanantone",
			},
		},
		{
			testName:   "testDeleteUserSuccess",
			shouldPass: true,
			callType:   DELETEUSER,
			expectedID: &pb.UserID{Id: 6},
			rqstData:   nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			var id *pb.UserID
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

func TestBulkAddUpdateUserGRPC(t *testing.T) {
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
		expectedIDs           []*pb.UserID
		rqstData              *pb.Users
		expectedOverallStatus pb.StatusEnum
		expectedResults       []string
	}{
		{
			testName:   "testAddUsersOneSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedIDs: []*pb.UserID{
				{Id: 8},
			},
			rqstData: &pb.Users{
				Users: []*pb.User{
					{
						AccountID: 1,
						Name:      "Peter Green",
						EMail:     "blackmagicwoman@gmail.com",
						Role:      pb.RoleEnum_UNRESTRICTED,
						Password:  "lasttraintosanantone",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusCreated,
			expectedResults:       []string{"Peter Green"},
		},
		{
			testName:   "testAddUsersTwoSuccess",
			shouldPass: true,
			callType:   CREATEUSER,
			expectedIDs: []*pb.UserID{
				{Id: 9},
				{Id: 10},
			},
			rqstData: &pb.Users{
				Users: []*pb.User{
					{
						AccountID: 1,
						Name:      "Brian Wilson",
						EMail:     "sloopjohnb@gmail.com",
						Role:      pb.RoleEnum_UNRESTRICTED,
						Password:  "helpmerhonda",
					},
					{
						AccountID: 1,
						Name:      "Frank Zappa",
						EMail:     "apostrophe@gmail.com",
						Role:      pb.RoleEnum_UNRESTRICTED,
						Password:  "donteatyellowsnow",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusCreated,
			expectedResults:       []string{"Brian Wilson", "Frank Zappa"},
		},
		{
			testName:   "testUpdateUsersOneSuccess",
			shouldPass: true,
			callType:   UPDATEUSER,
			expectedIDs: []*pb.UserID{
				{Id: 8},
			},
			rqstData: &pb.Users{
				Users: []*pb.User{
					{
						AccountID: 1,
						ID:        8,
						Name:      "Fleetwood Mac Peter Green",
						EMail:     "blackmagicwoman@gmail.com",
						Role:      pb.RoleEnum_UNRESTRICTED,
						Password:  "lasttraintosanantone",
					},
				},
			},
			expectedOverallStatus: pb.StatusEnum_StatusOK,
			expectedResults:       []string{"Fleetwood Mac Peter Green"},
		},
		// This test is non-deterministic as the actual creation order of the
		// 2 users can't be predicted. This means the IDs used below for updating
		// may not be correct. If not, there will be a "duplicate entry" error on
		// updating as the 'Email' fields must be unique, e.g., updating 'Beach Boy Brian Wilson's"
		// email address with Zappa's will violate the duplicate email address restriction, as
		// Zappa's already exists.
		// {
		// 	testName:   "testUpdateUsersTwoSuccess",
		// 	shouldPass: true,
		// 	callType:   UPDATEUSER,
		// 	expectedIDs: []*pb.UserID{
		// 		{Id: 9},
		// 		{Id: 10},
		// 	},
		// 	rqstData: &pb.Users{
		// 		Users: []*pb.User{
		// 			&pb.User{
		// 				AccountID: 1,
		// 				ID:        9,
		// 				Name:      "Beach Boy Brian Wilson",
		// 				EMail:     "sloopjohnb@gmail.com",
		// 				Role:      pb.RoleEnum_UNRESTRICTED,
		// 				Password:  "helpmerhonda",
		// 			},
		// 			&pb.User{
		// 				AccountID: 1,
		// 				ID:        10,
		// 				Name:      "Mothers of Invention Frank Zappa",
		// 				EMail:     "apostrophe@gmail.com",
		// 				Role:      pb.RoleEnum_UNRESTRICTED,
		// 				Password:  "donteatyellowsnow",
		// 			},
		// 		},
		// 	},
		// 	expectedOverallStatus: pb.StatusEnum_StatusOK,
		// 	expectedResults:       []string{"Beach Boy Brian Wilson", "Mothers of Invention Frank Zappa"},
		// },
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

				if len(resp.Response) != len(tc.expectedIDs) {
					t.Errorf("EXPECTED %d responses, GOT %d", len(tc.expectedIDs), len(resp.Response))
				}

				var actual string

				for _, result := range resp.Response {
					u, err := client.GetUser(context.Background(), result.UserID)
					if err != nil {
						t.Fatalf("error '%s' was not expected calling accountd server", err)
					}

					uString, err := json.Marshal(u)
					if err != nil {
						t.Fatalf("error '%s' was not expected while marshaling %v", err, resp)
					}

					actual = strings.Join([]string{actual, string(uString)}, "")
				}

				for _, name := range tc.expectedResults {
					if ok := strings.Contains(actual, name); !ok {
						t.Errorf("expected %s substring to be contained within %s", name, actual)
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
