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
	"github.com/youngkin/mockvideo/internal/users"
)

type AllCustsTests struct {
	testName     string
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, users.Users)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
}
type SingleCustTests struct {
	testName     string
	custID       int
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
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

			actual, err := users.GetAllUsers(db)
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

			actual, err := users.GetUser(db, tc.custID)

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
