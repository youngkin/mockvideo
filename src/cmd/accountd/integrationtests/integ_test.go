// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrationtests

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// TestGetAllUsers attempts to retrieve users from a running accountd server
// connected to a real database.
func TestGetUsers(t *testing.T) {
	// Takes a while for the accountd container to start
	time.Sleep(500 * time.Millisecond)

	tcs := []struct {
		testName           string
		url                string
		shouldPass         bool
		expectedHTTPStatus int
	}{
		{
			testName:           "testGetAllUsersSuccess",
			url:                "http://localhost:5000/users",
			shouldPass:         true,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetOneUsersSuccess",
			url:                "http://localhost:5000/users/1",
			shouldPass:         true,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetAllUsersUnexpectedResource",
			url:                "http://localhost:5000/unexpectedresource",
			shouldPass:         false,
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			resp, err := http.Get(tc.url)
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

				if *update {
					updateGoldenFile(t, tc.testName, string(actual))
				}

				expected := readGoldenFile(t, tc.testName)

				if expected != string(actual) {
					t.Errorf("expected %s, got %s", expected, string(actual))
				}
			}
		})
	}
}

func TestPOSTPUTDELETEUser(t *testing.T) {
	client := &http.Client{}

	tcs := []struct {
		testName                 string
		shouldPass               bool
		method                   string
		url                      string
		expectedHTTPStatus       int
		expectedGETSTatus        int
		expectedResourceLocation string
		newResourceURL           string
		rqstData                 string
	}{
		{
			testName:                 "testPOSTUserSuccess",
			shouldPass:               true,
			method:                   http.MethodPost,
			url:                      "http://localhost:5000/users",
			expectedHTTPStatus:       http.StatusCreated,
			expectedGETSTatus:        http.StatusOK,
			expectedResourceLocation: "/users/6",
			newResourceURL:           "http://localhost:5000/users/6",
			rqstData: `{
				"accountid":1,
				"name":"Brian Wilson",
				"email":"goodvibrations@gmail.com",
				"role":1,
				"password":"helpmerhonda"}`,
		},
		{
			testName:                 "testPOSTDuplicateUserFailure", //Dup email address
			shouldPass:               false,
			method:                   http.MethodPost,
			url:                      "http://localhost:5000/users",
			expectedHTTPStatus:       http.StatusBadRequest,
			expectedGETSTatus:        http.StatusTeapot, // NA, shouldn't even test this
			expectedResourceLocation: "",
			newResourceURL:           "",
			rqstData: `{
				"accountid":1,
				"name":"Brian Wilson",
				"email":"goodvibrations@gmail.com",
				"role":1,
				"password":"helpmerhonda"}`,
		},
		{
			testName:                 "testPUTUserSuccess",
			shouldPass:               true,
			method:                   http.MethodPut,
			url:                      "http://localhost:5000/users/6",
			expectedHTTPStatus:       http.StatusOK,
			expectedGETSTatus:        http.StatusOK,
			expectedResourceLocation: "NA, this is a PUT, not a POST",
			newResourceURL:           "http://localhost:5000/users/6",
			rqstData: `{
				"accountid":1,
				"id":6,
				"name":"BeachBoy Brian Wilson",
				"email":"goodvibrations@gmail.com",
				"role":1,
				"password":"helpmerhonda"}`,
		},
		{
			testName:                 "testDELETEUserSuccess",
			shouldPass:               true,
			method:                   http.MethodDelete,
			url:                      "http://localhost:5000/users/6",
			expectedHTTPStatus:       http.StatusOK,
			expectedGETSTatus:        http.StatusNotFound,
			expectedResourceLocation: "NA, this is a DELETE, not a POST",
			newResourceURL:           "http://localhost:5000/users/6",
			rqstData:                 "",
		},
		{
			testName:                 "testPUTNonExistingUserFailure",
			shouldPass:               false,
			method:                   http.MethodPut,
			url:                      "http://localhost:5000/users/6",
			expectedHTTPStatus:       http.StatusBadRequest,
			expectedGETSTatus:        http.StatusTeapot, // NA, shouldn't even test this
			expectedResourceLocation: "",
			newResourceURL:           "",
			rqstData: `{
				"accountid":1,
				"id":6,
				"name":"BeachBoy Brian Wilson",
				"email":"goodvibrations@gmail.com",
				"role":1,
				"password":"helpmerhonda"}`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer([]byte(tc.rqstData)))
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

			// Verify Insert/Update/DELETE
			if tc.shouldPass {
				if tc.method == http.MethodPost {
					location := resp.Header["Location"][0]
					if tc.expectedResourceLocation != location {
						t.Errorf("expected resource location %s, got %s", tc.expectedResourceLocation, location)
					}
				}

				resp, err = http.Get(tc.newResourceURL)
				if err != nil {
					t.Fatalf("error '%s' was not expected calling accountd server", err)
				}
				defer resp.Body.Close()

				status = resp.StatusCode
				if status != tc.expectedGETSTatus {
					t.Errorf("expected StatusCode = %d, got %d", tc.expectedGETSTatus, status)
				}

				actual, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("an error '%s' was not expected reading response body", err)
				}

				if *update {
					updateGoldenFile(t, tc.testName, string(actual))
				}

				expected := readGoldenFile(t, tc.testName)

				if expected != string(actual) {
					t.Errorf("expected %s, got %s", expected, string(actual))
				}
			}

		})
	}
}
