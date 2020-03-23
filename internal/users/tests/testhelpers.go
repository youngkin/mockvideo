// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/internal/users"
)

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to user data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, users.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(0, 1, "porgy tirebiter", "porgytirebiter@email.com", users.Primary).
		AddRow(0, 2, "mickey dolenz", "mdolenz@themonkeys.com", users.Restricted)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnRows(rows)

	expected := users.Users{
		Users: []*users.User{
			{
				AccountID: 0,
				ID:        1,
				Name:      "porgy tirebiter",
				EMail:     "porgytirebiter@email.com",
				Role:      users.Primary,
			},
			{
				AccountID: 0,
				ID:        2,
				Name:      "mickey dolenz",
				EMail:     "mdolenz@themonkeys.com",
				Role:      users.Restricted,
			},
		},
	}

	return db, mock, expected
}

// DBInsertSetupHelper encapsulates the common code needed to setup a mock User insert
func DBInsertSetupHelper(t *testing.T, u users.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("INSERT INTO user").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	return db, mock
}

// DBInsertErrorSetupHelper encapsulates the common code needed to mock a user insert error
func DBInsertErrorSetupHelper(t *testing.T, u users.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("INSERT INTO user").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
		WillReturnError(fmt.Errorf("some error"))

	return db, mock
}

// DBUpdateSetupHelper encapsulates the common code needed to setup a mock User insert
func DBUpdateSetupHelper(t *testing.T, u users.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	// mock.ExpectExec("UPDATE user").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
	// 	WillReturnResult(sqlmock.NewResult(1, 1))

	return db, mock
}

// DBCallQueryErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, users.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, users.Users{}
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, users.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnRows(rows)

	return db, mock, users.Users{}
}

// DBCallTeardownHelper encapsulates common code needed to finalize processing of mock DB access to user data
func DBCallTeardownHelper(t *testing.T, mock sqlmock.Sqlmock) {
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// GetUserSetupHelper encapsulates common code needed to setup mock DB access a single users's data
func GetUserSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(5, 1, "porgy tirebiter", "porgytirebiter@email.com", users.Primary)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnRows(rows)

	expected := users.User{
		AccountID: 5,
		ID:        1,
		Name:      "porgy tirebiter",
		EMail:     "porgytirebiter@email.com",
		Role:      users.Primary,
	}

	return db, mock, &expected
}

// DBUserErrNoRowsSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBUserErrNoRowsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnError(sql.ErrNoRows)

	return db, mock, nil
}

// DBUserOtherErrSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBUserOtherErrSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnError(sql.ErrConnDone)

	return db, mock, nil
}

// DBCallNoExpectationsSetupHelper encapsulates common coded needed to when no expectations are present
func DBCallNoExpectationsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *users.User) {
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
