package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/customers"
	"github.com/youngkin/mockvideo/internal/platform/logging"
)

type Tests struct {
	testName           string
	shouldPass         bool
	setupFunc          func(*testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers)
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

func TestGetAllCustomers(t *testing.T) {
	tests := []Tests{
		{
			testName:           "testGetAllCustomersSuccess",
			shouldPass:         true,
			setupFunc:          customers.DBCallSetupHelper,
			teardownFunc:       customers.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetAllCustomersQueryFailure",
			shouldPass:         false,
			setupFunc:          customers.DBCallQueryErrorSetupHelper,
			teardownFunc:       customers.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
		{
			testName:           "testGetAllCustomersRowScanFailure",
			shouldPass:         false,
			setupFunc:          customers.DBCallRowScanErrorSetupHelper,
			teardownFunc:       customers.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			custHandler, err := New(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a customer handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(custHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling customerd server", err)
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
					t.Errorf("expected %+v, got %+v", string(mExpected), actual)
				}
			}

			// we make sure that all post-conditions were met
			customers.DBCallTeardownHelper(t, mock)
		})
	}
}
