// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tests

/*
This file and associated tests demostrate the use of DB mocking via DATA-DOG's
'go-sqlmock' package.
*/

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/src/api"
	"github.com/youngkin/mockvideo/src/internal/user"
)

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		testName     string
		shouldPass   bool
		setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, api.Users)
		teardownFunc func(*testing.T, sqlmock.Sqlmock)
	}{
		{
			testName:     "testGetAllUsersSuccess",
			shouldPass:   true,
			setupFunc:    DBCallSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetAllUsersQueryFailure",
			shouldPass:   false,
			setupFunc:    DBCallQueryErrorSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetAllUsersRowScanFailure",
			shouldPass:   false,
			setupFunc:    DBCallRowScanErrorSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			actual, err := user.GetAllUsers(db)
			if tc.shouldPass && err != nil {
				t.Fatalf("error '%s' was not expected", err)
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("expected error didn't occur")
			}

			if len(expected.Users) != len(actual.Users) {
				t.Errorf("expected %d users, got %d", len(expected.Users), len(actual.Users))
			}

			fail := false
			failMsg := []string{}

			for i, user := range expected.Users {
				if *user != *actual.Users[i] {
					fail = true
					failMsg = append(failMsg, fmt.Sprintf("expected %+v, got %+v", user, actual.Users[i]))
				}
			}
			if fail {
				t.Errorf("%+v", failMsg)
			}

			tc.teardownFunc(t, mock)

		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		testName     string
		userID       int
		shouldPass   bool
		setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *api.User)
		teardownFunc func(*testing.T, sqlmock.Sqlmock)
	}{
		{
			testName:     "testGetUserSuccess",
			userID:       1,
			shouldPass:   true,
			setupFunc:    GetUserSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetUserNoRow",
			userID:       1,
			shouldPass:   true, // true because we get a nil 'User' if not found
			setupFunc:    DBUserErrNoRowsSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetUserQueryError",
			userID:       1,
			shouldPass:   false,
			setupFunc:    DBUserOtherErrSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			actual, err := user.GetUser(db, tc.userID)

			validateExpectedErrors(t, err, tc.shouldPass)

			if expected == nil && actual == nil {
				validateExpectedErrors(t, err, tc.shouldPass)
				return
			}
			if *expected != *actual {
				t.Errorf("expected %+v , got %+v", expected, actual)
			}
			tc.teardownFunc(t, mock)
		})
	}
}

func TestInsertUser(t *testing.T) {
	tests := []struct {
		testName       string
		user           api.User
		expectedUserID int
		shouldPass     bool
		setupFunc      func(*testing.T, api.User) (*sql.DB, sqlmock.Sqlmock)
		teardownFunc   func(*testing.T, sqlmock.Sqlmock)
	}{
		{
			testName: "testInsertUserSuccess",
			user: api.User{
				AccountID: 1,
				Name:      "mama cass",
				EMail:     "mama@gmail.com",
				Role:      0,
				Password:  "myawsomepassword",
			},
			expectedUserID: 1,
			shouldPass:     true,
			setupFunc:      DBInsertSetupHelper,
			teardownFunc:   DBCallTeardownHelper,
		},
		{
			testName: "testInsertUsererror",
			user: api.User{
				AccountID: 1,
				Name:      "mama cass",
				EMail:     "mama@gmail.com",
				Role:      0,
				Password:  "myawsomepassword",
			},
			expectedUserID: 0,
			shouldPass:     false,
			setupFunc:      DBInsertErrorSetupHelper,
			teardownFunc:   DBCallTeardownHelper,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.user)
			defer db.Close()

			uID, _, err := user.InsertUser(db, tc.user)

			validateExpectedErrors(t, err, tc.shouldPass)

			if tc.expectedUserID != uID {
				t.Errorf("expected %+v , got %+v", tc.expectedUserID, uID)
			}
			tc.teardownFunc(t, mock)
		})
	}
}
