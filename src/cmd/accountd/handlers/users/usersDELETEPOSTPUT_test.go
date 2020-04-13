// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	logging "github.com/youngkin/mockvideo/src/internal/platform/logging"
	"github.com/youngkin/mockvideo/src/internal/user"
	"github.com/youngkin/mockvideo/src/internal/user/tests"
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

type TestCase struct {
	testName           string
	shouldPass         bool
	url                string
	expectedHTTPStatus int
	updateResourceID   string
	expectedResourceID string
	postData           string
	user               user.User
	setupFunc          func(*testing.T, user.User) (*sql.DB, sqlmock.Sqlmock)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
}

func TestPOSTUser(t *testing.T) {
	tcs := []TestCase{
		{
			testName:           "testInsertUserSuccess",
			shouldPass:         true,
			url:                "/users",
			expectedHTTPStatus: http.StatusCreated,
			expectedResourceID: "/users/1",
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
			// On insert the URL must not include a resource ID
			testName:           "testInsertUserFailInvalidURL",
			shouldPass:         false,
			url:                "/users/1",
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResourceID: "",
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
			setupFunc:    tests.DBNoCallSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			// On insert the JSON body must not include user ID
			testName:           "testInsertUserFailInvalidJSON",
			shouldPass:         false,
			url:                "/users",
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResourceID: "",
			postData: `
			{
				"ID": 1,
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
			setupFunc:    tests.DBNoCallSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.user)

			srvHandler, err := NewUserHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(srvHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(tc.postData)))
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling accountd server", err)
			}
			defer resp.Body.Close()

			if tc.shouldPass {
				resourceURL := resp.Header.Get("Location")
				if string(resourceURL) != tc.expectedResourceID {
					t.Errorf("expected resource %s, got %s", tc.expectedResourceID, resourceURL)
				}
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}

func TestPUTUser(t *testing.T) {
	client := &http.Client{}

	tcs := []TestCase{
		{
			testName:           "testUpdateUserSuccess",
			shouldPass:         true,
			url:                "/users/2",
			expectedHTTPStatus: http.StatusOK,
			updateResourceID:   "users/2",
			expectedResourceID: "",
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
				ID:        2,
				AccountID: 1,
				Name:      "mickey dolenz",
				EMail:     "mickeyd@gmail.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBUpdateSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			// MySQL UPDATE will insert a row if it doesn't exist. Needed to prohibit
			// this hence this test.
			testName:           "testUpdateNonExistUser",
			shouldPass:         false,
			url:                "/users/100",
			expectedHTTPStatus: http.StatusBadRequest,
			updateResourceID:   "users/100",
			expectedResourceID: "",
			postData: `
			{
				"ID": 100,
				"AccountID":1,
				"Name":"Mickey Mouse",
				"eMail":"MickeyMoused@disney.com",
				"role":1,
				"password":"myawesomepassword"
			}
			`,
			user: user.User{
				ID:        100,
				AccountID: 1,
				Name:      "Mickey Mouse",
				EMail:     "MickeyMoused@disney.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBUpdateNonExistingRowSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			testName:           "testUpdateDBError",
			shouldPass:         false,
			url:                "/users/100",
			expectedHTTPStatus: http.StatusInternalServerError,
			updateResourceID:   "users/100",
			expectedResourceID: "",
			postData: `
			{
				"ID": 100,
				"AccountID":1,
				"Name":"Mickey Mouse",
				"eMail":"MickeyMoused@disney.com",
				"role":1,
				"password":"myawesomepassword"
			}
			`,
			user: user.User{
				ID:        100,
				AccountID: 1,
				Name:      "Mickey Mouse",
				EMail:     "MickeyMoused@disney.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBUpdateErrorSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			testName:           "testUpdateDBSelectError",
			shouldPass:         false,
			url:                "/users/100",
			expectedHTTPStatus: http.StatusInternalServerError,
			updateResourceID:   "users/100",
			expectedResourceID: "",
			postData: `
		{
			"ID": 100,
			"AccountID":1,
			"Name":"Mickey Mouse",
			"eMail":"MickeyMoused@disney.com",
			"role":1,
			"password":"myawesomepassword"
		}
		`,
			user: user.User{
				ID:        100,
				AccountID: 1,
				Name:      "Mickey Mouse",
				EMail:     "MickeyMoused@disney.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBUpdateErrorSelectSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.user)

			srvHandler, err := NewUserHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(srvHandler.ServeHTTP))
			defer testSrv.Close()

			// NOTE: As there is no http.PUT creating an update request/PUT requires
			//	1.	Creating an http.Client (done at the top of this function)
			//	2.	Creating the request
			//	3. 	Calling client.DO
			//
			// Kind of round-about, but it works
			url := testSrv.URL + tc.url
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(tc.postData)))
			if err != nil {
				t.Fatalf("an error '%s' was not expected creating HTTP request", err)
			}

			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling (client.Do()) accountd server", err)
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}

func TestDELETEUser(t *testing.T) {
	client := &http.Client{}

	tcs := []TestCase{
		{
			testName:           "testDeleteUserSuccess",
			shouldPass:         true,
			url:                "/users/2",
			expectedHTTPStatus: http.StatusOK,
			updateResourceID:   "users/2",
			expectedResourceID: "",
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
				ID:        2,
				AccountID: 1,
				Name:      "mickey dolenz",
				EMail:     "mickeyd@gmail.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBDeleteSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
		{
			testName:           "testDeleteUserFailed",
			shouldPass:         false,
			url:                "/users/2",
			expectedHTTPStatus: http.StatusInternalServerError,
			updateResourceID:   "users/2",
			expectedResourceID: "",
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
				ID:        2,
				AccountID: 1,
				Name:      "mickey dolenz",
				EMail:     "mickeyd@gmail.com",
				Role:      1,
				Password:  "myawesomepassword",
			},
			setupFunc:    tests.DBDeleteErrorSetupHelper,
			teardownFunc: tests.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.user)

			srvHandler, err := NewUserHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(srvHandler.ServeHTTP))
			defer testSrv.Close()

			// NOTE: As there is no http.DELETE creating an update request/DELETE requires
			//	1.	Creating an http.Client (done at the top of this function)
			//	2.	Creating the request
			//	3. 	Calling client.DO
			//
			// Kind of round-about, but it works
			url := testSrv.URL + tc.url
			req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
			if err != nil && tc.shouldPass {
				t.Fatalf("an error '%s' was not expected creating HTTP request", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling (client.Do()) accountd server", err)
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}
