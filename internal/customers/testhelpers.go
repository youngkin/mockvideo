package customers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to customer data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "streetAddress", "city", "state", "country"}).
		AddRow(1, "porgy tirebiter", "123 anyStreet", "anyCity", "anyState", "anyCountry").
		AddRow(2, "mickey dolenz", "123 Laurel Canyon", "LA", "CA", "USA")

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	expected := Customers{
		Customers: []Customer{
			{
				ID:            1,
				Name:          "porgy tirebiter",
				StreetAddress: "123 anyStreet",
				City:          "anyCity",
				State:         "anyState",
				Country:       "anyCountry",
			},
			{
				ID:            2,
				Name:          "mickey dolenz",
				StreetAddress: "123 Laurel Canyon",
				City:          "LA",
				State:         "CA",
				Country:       "USA",
			},
		},
	}

	return db, mock, expected
}

// DBCallQueryErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, Customers{}
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	return db, mock, Customers{}
}

// DBCallTeardownHelper encapsulates common code needed to finalize processing of mock DB access to customer data
func DBCallTeardownHelper(t *testing.T, mock sqlmock.Sqlmock) {
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
