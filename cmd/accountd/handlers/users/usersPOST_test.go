// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	logging "github.com/youngkin/mockvideo/internal/platform/logging"
	user "github.com/youngkin/mockvideo/internal/user"
	tests "github.com/youngkin/mockvideo/internal/user/tests"
)

// logger is used to control code-under-test logging behavior
var logger *log.Entry

func init() {
	logger = logging.GetLogger()
	// Uncomment for more verbose logging
	// logger.Logger.SetLevel(log.DebugLevel)
	// Suppress all application logging
	logger.Logger.SetLevel(log.PanicLevel)
	// Uncomment for non-tty logging
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	//  })
}

type POSTTest struct {
	testName           string
	shouldPass         bool
	expectedHTTPStatus int
	updateResourceID   string
	expectedResourceID string
	postData           string
	user               user.User
	setupFunc          func(*testing.T, user.User) (*sql.DB, sqlmock.Sqlmock)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
}

func TestPOSTUser(t *testing.T) {
	tcs := []POSTTest{
		{
			testName:           "testInsertUserSuccess",
			shouldPass:         true,
			expectedHTTPStatus: http.StatusCreated,
			expectedResourceID: "1",
			postData: `
			{
				"AccountID":1,
				"Name":"mickey dolenz",
				"eMail":"mickeyd@gmail.com",
				"role":1,
				"password":"myawesomepassword"
			}
			`,
			user: user.User{
				AccountID: 1,
				Name:      "mickey dolenz",
				EMail:     "mickeyd@gmail.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBInsertSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			testName:           "testUpdateUserSuccess",
			shouldPass:         true,
			expectedHTTPStatus: http.StatusCreated,
			updateResourceID:   "/2",
			expectedResourceID: "2",
			postData: `
			{
				"ID": 2,
				"AccountID":1,
				"Name":"mickey dolenz",
				"eMail":"mickeyd@gmail.com",
				"role":1,
				"password":"myawesomepassword"
			}
			`,
			user: user.User{
				ID:        1,
				AccountID: 2,
				Name:      "mickey dolenz",
				EMail:     "mickeyd@gmail.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBUpdateSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.user)

			postHandler, err := NewUserHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(postHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + "/users" + tc.updateResourceID
			resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(tc.postData)))
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling accountd server", err)
			}
			defer resp.Body.Close()

			resourceURL := resp.Header.Get("Location")
			expectedResourceURL := "/users/" + tc.expectedResourceID
			if string(resourceURL) != expectedResourceURL {
				t.Errorf("expected resource %s, got %s", expectedResourceURL, resourceURL)
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}
