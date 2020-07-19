// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/mockvideo/src/domain"
)

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to user data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(0, 1, "porgy tirebiter", "porgytirebiter@email.com", domain.Primary).
		AddRow(0, 2, "mickey dolenz", "mdolenz@themonkeys.com", domain.Restricted)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnRows(rows)

	expected := domain.Users{
		Users: []*domain.User{
			{
				AccountID: 0,
				ID:        1,
				Name:      "porgy tirebiter",
				EMail:     "porgytirebiter@email.com",
				Role:      domain.Primary,
			},
			{
				AccountID: 0,
				ID:        2,
				Name:      "mickey dolenz",
				EMail:     "mdolenz@themonkeys.com",
				Role:      domain.Restricted,
			},
		},
	}

	return db, mock, &expected
}

// DBDeleteSetupHelper encapsulates the common code needed to setup a mock User delete
func DBDeleteSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("DELETE FROM user WHERE id = ?").WithArgs(u.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	return db, mock
}

// DBDeleteErrorSetupHelper encapsulates the common code needed to mock a user delete error
func DBDeleteErrorSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("DELETE FROM user WHERE id = ?").WithArgs(u.ID).WillReturnError(sql.ErrConnDone)

	return db, mock
}

// DBInsertSetupHelper encapsulates the common code needed to setup a mock User insert
func DBInsertSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("INSERT INTO user").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	return db, mock
}

// DBInsertErrorSetupHelper encapsulates the common code needed to mock a user insert error
func DBInsertErrorSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("INSERT INTO user").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
		WillReturnError(fmt.Errorf("some error"))

	return db, mock
}

// DBNoCallSetupHelper encapsulates the common code needed to mock an error upstream from an actual DB call
func DBNoCallSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock
}

// DBUpdateNonExistingRowSetupHelper mimics an update to a non-existing user, can't update non-existing domain.
func DBUpdateNonExistingRowSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user WHERE id = ?").WithArgs(u.ID).WillReturnError(sql.ErrNoRows)

	return db, mock
}

// DBUpdateErrorSelectSetupHelper mimics an update where the non-existence query fails.
func DBUpdateErrorSelectSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user WHERE id = ?").WithArgs(u.ID).
		WillReturnError(sql.ErrConnDone)

	return db, mock
}

// DBUpdateSetupHelper encapsulates the common code needed to setup a mock User update
func DBUpdateSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(1, 2, "mickey dolenz", "mickeyd@gmail.com", domain.Unrestricted)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user WHERE id = ?").WithArgs(u.ID).WillReturnRows(rows)
	mock.ExpectExec("UPDATE user SET (.+) WHERE (.+)").WithArgs(u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // no insert ID, 1 row affected
	mock.ExpectCommit()
	return db, mock
}

// DBUpdateErrorSetupHelper encapsulates the common code needed to setup a mock User update error
func DBUpdateErrorSetupHelper(t *testing.T, u domain.User) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(1, 100, "Mickey Mouse", "MickeyMoused@disney.com", domain.Unrestricted)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user WHERE id = ?").WithArgs(u.ID).
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE user SET (.+) WHERE (.+)").WithArgs(u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()
	return db, mock
}

// DBCallQueryErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, nil
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.Users) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WillReturnRows(rows)

	return db, mock, nil
}

// DBCallTeardownHelper encapsulates common code needed to finalize processing of mock DB access to user data
func DBCallTeardownHelper(t *testing.T, mock sqlmock.Sqlmock) {
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// GetUserSetupHelper encapsulates common code needed to setup mock DB access a single users's data
func GetUserSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
		AddRow(5, 1, "porgy tirebiter", "porgytirebiter@email.com", domain.Primary)

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnRows(rows)

	expected := domain.User{
		AccountID: 5,
		ID:        1,
		Name:      "porgy tirebiter",
		EMail:     "porgytirebiter@email.com",
		Role:      domain.Primary,
	}

	return db, mock, &expected
}

// DBUserErrNoRowsSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBUserErrNoRowsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnError(sql.ErrNoRows)

	return db, mock, nil
}

// DBUserOtherErrSetupHelper encapsulates common coded needed to mock Queries returning no rows
func DBUserOtherErrSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.User) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM user").
		WithArgs(1).WillReturnError(sql.ErrConnDone)

	return db, mock, nil
}

// DBCallNoExpectationsSetupHelper encapsulates common coded needed to when no expectations are present
func DBCallNoExpectationsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *domain.User) {
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
