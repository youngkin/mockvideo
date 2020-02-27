package customers

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/internal/customers"
)

type Tests struct {
	testName     string
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
}

func TestGetAllCustomers(t *testing.T) {
	tests := []Tests{
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

			actual, err := customers.GetAllCustomers(db)
			if tc.shouldPass && err != nil {
				t.Fatalf("error '%s' was not expected", err)
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("expected error didn't occur")
			}

			if reflect.DeepEqual(expected, actual) != true {
				t.Errorf("expected %+v, got %+v", expected, actual)
			}

			tc.teardownFunc(t, mock)

		})
	}
}
