// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/youngkin/mockvideo/src/internal/platform/constants"
)

// Role indicates what role a User can take regarding account actions (e.g., add a service)
type Role int

const (
	// Primary user role can do anything on the account
	Primary Role = iota
	// Unrestricted user role can do anything except billing
	Unrestricted
	// Restricted can't do much of anything, nothing service related, nothing billing related, basically just email
	Restricted
)

var (
	getAllUsersQuery = "SELECT accountID, id, name, email, role FROM user"
	getUserQuery     = "SELECT accountID, id, name, email, role FROM user WHERE id = ?"
	insertUserStmt   = "INSERT INTO user (accountID, name, email, role, password) VALUES (?, ?, ?, ?, ?)"
	updateUserStmt   = "UPDATE user SET id = ?, accountID = ?, name = ?, email = ?, role = ?, password = ? WHERE id = ?"
	deleteUserStmt   = "DELETE FROM user WHERE id = ?"
)

// User represents the data about a user
type User struct {
	AccountID int    `json:"accountid"`
	HREF      string `json:"href"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	EMail     string `json:"email"`
	Role      Role   `json:"role"`
	Password  string `json:"password,omitempty"`
}

// Users is a collection (slice) of User
type Users struct {
	Users []*User `json:"Users"`
}

// GetAllUsers will return all customers known to the application
func GetAllUsers(db *sql.DB) (*Users, error) {
	results, err := db.Query(getAllUsersQuery)
	if err != nil {
		return &Users{}, errors.Annotate(err, "error querying DB")
	}

	us := Users{}
	for results.Next() {
		var u User

		err = results.Scan(&u.AccountID,
			&u.ID,
			&u.Name,
			&u.EMail,
			&u.Role)
		if err != nil {
			return &Users{}, errors.Annotate(err, "error scanning result set")
		}

		us.Users = append(us.Users, &u)
	}

	return &us, nil
}

// GetUser will return the user identified by 'id' or a nil user if there
// wasn't a matching user.
func GetUser(db *sql.DB, id int) (*User, error) {
	row := db.QueryRow(getUserQuery, id)
	user := &User{}
	err := row.Scan(&user.AccountID,
		&user.ID,
		&user.Name,
		&user.EMail,
		&user.Role)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Annotate(err, "error scanning user row")
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, nil
}

// InsertUser takes the provided user data, inserts it into the db, and returns the newly created user ID.
func InsertUser(db *sql.DB, u User) (int64, constants.ErrCode, error) {
	err := validateUser(u)
	if err != nil {
		return 0, constants.UserValidationErrorCode, errors.Annotate(err, "User validation failure")
	}

	r, err := db.Exec(insertUserStmt, u.AccountID, u.Name, u.EMail, u.Role, u.Password)
	if err != nil {
		errDetail, ok := err.(*mysql.MySQLError)
		if ok {
			if errDetail.Number == constants.MySQLDupInsertErrorCode {
				return 0, constants.DBInsertDuplicateUserErrorCode, errors.Annotate(err, fmt.Sprintf("error inserting duplicate user into the database: %+v, possible duplicate email address", u))
			}
		} else {
			return 0, constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error inserting user %+v into DB", u))
		}
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, constants.DBUpSertErrorCode, errors.Annotate(err, "error getting user ID")
	}

	return id, constants.NoErrorCode, nil
}

// UpdateUser takes the provided user data, inserts it into the db, and returns the newly created user ID
func UpdateUser(db *sql.DB, u User) (constants.ErrCode, error) {
	err := validateUser(u)
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, "User validation failure")
	}

	// This entire db.Begin/tx.Rollback/Commit seem awkward to me. But it's here because
	// MySQL silently performs an insert if there is no row to update.
	tx, err := db.Begin()
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error beginning transaction for user: %+v", u))
	}
	r := db.QueryRow(getUserQuery, u.ID)
	userRow := User{}
	err = r.Scan(&userRow.AccountID,
		&userRow.ID,
		&userRow.Name,
		&userRow.EMail,
		&userRow.Role)

	if err != nil && err == sql.ErrNoRows {
		tx.Rollback()
		return constants.DBInvalidRequestCode, errors.New(fmt.Sprintf("error, attempting to update non-existent user, user.ID %d", u.ID))
	}
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating user in the database: %+v", u))
	}

	_, err = db.Exec(updateUserStmt, u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID)
	if err != nil {
		tx.Rollback()
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating user in the database: %+v", u))
	}
	tx.Commit()

	return constants.NoErrorCode, nil
}

// DeleteUser deletes the user identified by u.id from the database
func DeleteUser(db *sql.DB, id int) (constants.ErrCode, error) {
	_, err := db.Exec(deleteUserStmt, id)
	if err != nil {
		return constants.DBDeleteErrorCode, errors.Annotate(err, "Usesr delete error")
	}

	return constants.NoErrorCode, nil
}

// IsAuthorizedUser will return true if the encryptedPassword matches the
// User's real (i.e., unencrypted) password.
func IsAuthorizedUser(db *sql.DB, id int, encryptedPassword []byte) (bool, error) {
	// TODO: implement
	return false, errors.NewNotImplemented(nil, "Not implemented")
}

func validateUser(u User) error {
	errMsg := ""

	if u.AccountID == 0 {
		errMsg = errMsg + "AccountID cannot be 0"
	}
	if len(u.EMail) == 0 {
		errMsg = errMsg + "; Email address must be populated"
	}
	if len(u.Name) == 0 {
		errMsg = errMsg + "; Name must be populated"
	}
	if len(u.Password) == 0 {
		errMsg = errMsg + "; Password must be populated"
	}
	if u.Role != Primary && u.Role != Restricted && u.Role != Unrestricted {
		errMsg = errMsg + fmt.Sprintf("; Invalid Role. Role must be one of %d, %d, or %d, got %d",
			Primary, Restricted, Unrestricted, u.Role)
	}

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
