package customers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/internal/customers"
)

type AllCustsTests struct {
	testName     string
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers)
	teardownFunc func(*testing.T, sqlmock.Sqlmock)
}
type SingleCustTests struct {
	testName     string
	custID       int
	shouldPass   bool
	setupFunc    func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *customers.Customer)
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

			actual, err := customers.GetAllCustomers(db)
			if tc.shouldPass && err != nil {
				t.Fatalf("error '%s' was not expected", err)
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("expected error didn't occur")
			}

			if len(expected.Customers) != len(actual.Customers) {
				t.Errorf("expected %d customers, got %d", len(expected.Customers), len(actual.Customers))
			}

			fail := false
			failMsg := []string{}

			for i, cust := range expected.Customers {
				if *cust != *actual.Customers[i] {
					fail = true
					failMsg = append(failMsg, fmt.Sprintf("expected %+v, got %+v", cust, actual.Customers[i]))
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
			setupFunc:    GetCustSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetCustomerNoRow",
			custID:       1,
			shouldPass:   true, // true because we get a nil 'Customer' if not found
			setupFunc:    DBCustErrNoRowsSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
		{
			testName:     "testGetCustomerQueryError",
			custID:       1,
			shouldPass:   false,
			setupFunc:    DBCustOtherErrSetupHelper,
			teardownFunc: DBCallTeardownHelper,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			actual, err := customers.GetCustomer(db, tc.custID)

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
