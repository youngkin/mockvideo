// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handlers

/*
These tests and supporting code demonstrate the following:

1.  Table driven tests using 'Tests' and 'CustTests' structs and appropriate
	test instance definitions using struct literals in each test function
2.	Sub-tests. These are useful to get more detailed information from your test
	executions.
3.	The use of external helper functions for test setup and teardown.
*/

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/platform/logging"
	"github.com/youngkin/mockvideo/internal/users"
	tests "github.com/youngkin/mockvideo/internal/users/tests"
)

type Tests struct {
	testName           string
	shouldPass         bool
	setupFunc          func(*testing.T) (*sql.DB, sqlmock.Sqlmock, users.Users)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
	expectedHTTPStatus int
}
type CustTests struct {
	testName           string
	url                string
	shouldPass         bool
	setupFunc          func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
	expectedHTTPStatus int
}

// logger is used to control code-under-test logging behavior
var logger *log.Entry

func init() {
	logger = logging.GetLogger()
	// Uncomment for more verbose logging
	// logger.Logger.SetLevel(log.DebugLevel)
	// Supress all application logging
	logger.Logger.SetLevel(log.PanicLevel)
	// Uncomment for non-tty logging
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	//  })
}

func TestGetAllUsers(t *testing.T) {
	tcs := []Tests{
		{
			testName:           "testGetAllUsersSuccess",
			shouldPass:         true,
			setupFunc:          tests.DBCallSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetAllUsersQueryFailure",
			shouldPass:         false,
			setupFunc:          tests.DBCallQueryErrorSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
		{
			testName:           "testGetAllUsersRowScanFailure",
			shouldPass:         false,
			setupFunc:          tests.DBCallRowScanErrorSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			// populate User.HREF from User.ID
			for _, user := range expected.Users {
				user.HREF = "/users/" + strconv.Itoa(user.ID)
			}

			userHandler, err := New(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a user handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(userHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + "/users"
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling accountd server", err)
			}
			defer resp.Body.Close()

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			if tc.shouldPass {
				actual, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("an error '%s' was not expected reading response body", err)
				}

				mExpected, err := json.Marshal(expected)
				if err != nil {
					t.Fatalf("an error '%s' was not expected Marshaling %+v", err, expected)
				}

				if bytes.Compare(mExpected, actual) != 0 {
					t.Errorf("expected %+v, got %+v", string(mExpected), string(actual))
				}
			}

			// we make sure that all post-conditions were met
			tests.DBCallTeardownHelper(t, mock)
		})
	}
}

func TestGetUser(t *testing.T) {
	tcs := []CustTests{
		{
			testName:           "testGetUserSuccess",
			url:                "/users/1",
			shouldPass:         true,
			setupFunc:          tests.GetUserSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetUserURLTooLong",
			url:                "/users/1/extraNode",
			shouldPass:         false,
			setupFunc:          tests.DBCallNoExpectationsSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			testName:           "testGetUserURLNonNumericID",
			url:                "/users/notanumber",
			shouldPass:         false,
			setupFunc:          tests.DBCallNoExpectationsSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			testName:           "testGetUserErrNoRow",
			url:                "/users/notanumber",
			shouldPass:         false,
			setupFunc:          tests.DBCallNoExpectationsSetupHelper,
			teardownFunc:       tests.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			// populate User.HREF from User.ID
			if expected != nil {
				expected.HREF = tc.url
			}

			userHandler, err := New(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(userHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling accountd server", err)
			}
			defer resp.Body.Close()

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			if tc.shouldPass {
				actual, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("an error '%s' was not expected reading response body", err)
				}

				mExpected, err := json.Marshal(expected)
				if err != nil {
					t.Fatalf("an error '%s' was not expected Marshaling %+v", err, expected)
				}

				if bytes.Compare(mExpected, actual) != 0 {
					t.Errorf("expected %+v, got %+v", string(mExpected), string(actual))
				}
			}

			// we make sure that all post-conditions were met
			tests.DBCallTeardownHelper(t, mock)
		})
	}
}
