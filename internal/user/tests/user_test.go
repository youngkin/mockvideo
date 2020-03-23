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
	"github.com/youngkin/mockvideo/internal/user"
)

type AllCustsTests struct {
	testName     string
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, user.Users)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
}
type SingleCustTests struct {
	testName     string
	custID       int
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *user.User)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
}
type InsertCustTests struct {
	testName       string
	user           user.User
	expectedUserID int64
	shouldPass     bool
	setupFunc      func(*testing.T, user.User) (*sql.DB, sqlmock.Sqlmock)
	teardownFunc   func(*testing.T, sqlmock.Sqlmock)
}

func TestGetAllCustomers(t *testing.T) {
	tests := []AllCustsTests{
		{
			testName:     "testGetAllCustomersSuccess",
			shouldPass:   true,
			setupFunc:    DBCallSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetAllCustomersQueryFailure",
			shouldPass:   false,
			setupFunc:    DBCallQueryErrorSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetAllCustomersRowScanFailure",
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

func TestGetCustomer(t *testing.T) {
	tests := []SingleCustTests{
		{
			testName:     "testGetCustomerSuccess",
			custID:       1,
			shouldPass:   true,
			setupFunc:    GetUserSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetCustomerNoRow",
			custID:       1,
			shouldPass:   true, // true because we get a nil 'User' if not found
			setupFunc:    DBUserErrNoRowsSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetCustomerQueryError",
			custID:       1,
			shouldPass:   false,
			setupFunc:    DBUserOtherErrSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			actual, err := user.GetUser(db, tc.custID)

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

func TestInsertCustomer(t *testing.T) {
	tests := []InsertCustTests{
		{
			testName: "testInsertCustomerSuccess",
			user: user.User{
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
			testName: "testInsertCustomererror",
			user: user.User{
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

			uID, err := user.InsertUser(db, tc.user)

			validateExpectedErrors(t, err, tc.shouldPass)

			if tc.expectedUserID != uID {
				t.Errorf("expected %+v , got %+v", tc.expectedUserID, uID)
			}
			tc.teardownFunc(t, mock)
		})
	}
}
