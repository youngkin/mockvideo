// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package customers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/internal/customers"
)

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to customer data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "streetAddress", "city", "state", "country"}).
		AddRow(1, "porgy tirebiter", "123 anyStreet", "anyCity", "anyState", "anyCountry").
		AddRow(2, "mickey dolenz", "123 Laurel Canyon", "LA", "CA", "USA")

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	expected := customers.Customers{
		Customers: []*customers.Customer{
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
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, customers.Customers{}
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, customers.Customers) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	return db, mock, customers.Customers{}
}

// DBCallTeardownHelper encapsulates common code needed to finalize processing of mock DB access to customer data
func DBCallTeardownHelper(t *testing.T, mock sqlmock.Sqlmock) {
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// GetCustSetupHelper encapsulates common code needed to setup mock DB access a single customer's data
func GetCustSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *customers.Customer) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "streetAddress", "city", "state", "country"}).
		AddRow(1, "porgy tirebiter", "123 anyStreet", "anyCity", "anyState", "anyCountry")

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer WHERE id=1").
		WillReturnRows(rows)

	expected := &customers.Customer{
		ID:            1,
		Name:          "porgy tirebiter",
		StreetAddress: "123 anyStreet",
		City:          "anyCity",
		State:         "anyState",
		Country:       "anyCountry",
	}

	return db, mock, expected
}

// DBCustErrNoRowsSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBCustErrNoRowsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *customers.Customer) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer WHERE id=1").
		WillReturnError(sql.ErrNoRows)

	return db, mock, nil
}

// DBCustOtherErrSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBCustOtherErrSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *customers.Customer) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer WHERE id=1").
		WillReturnError(sql.ErrConnDone)

	return db, mock, nil
}

// DBCallNoExpectationsSetupHelper encapsulates common coded needed to when no expectations are present
func DBCallNoExpectationsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *customers.Customer) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock, nil
}

func validateExpectedErrors(t *testing.T, err error, shouldPass bool) {
	if shouldPass && err != nil {
		t.Fatalf("error '%s' was not expected", err)
	}
	if !shouldPass && err == nil {
		t.Fatalf("expected error didn't occur")
	}
}
