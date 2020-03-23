// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package users

import (
	"database/sql"
	"fmt"

	"github.com/juju/errors"
)

// UserRole indicates what role a User can take regarding account actions (e.g., add a service)
type UserRole int

const (
	// Primary user role can do anything on the account
	Primary UserRole = iota
	// Unrestricted user role can do anything except billing
	Unrestricted
	// Restricted can't do much of anything, nothing service related, nothing billing related, basically just email
	Restricted
)

var (
	getAllUsersQuery = "SELECT accountID, id, name, email, role FROM user"
	getUserQuery     = "SELECT accountID, id, name, email, role FROM user WHERE id = ?"
	insertUserStmt   = "INSERT INTO user (accountID, name, email, role, password) VALUES (?, ?, ?, ?, ?)"
)

// User represents the data about a user
type User struct {
	AccountID int      `json:"AccountID"`
	HREF      string   `json:"href"`
	ID        int      `json:"ID"`
	Name      string   `json:"Name"`
	EMail     string   `json:"eMail"`
	Role      UserRole `json:"role"`
	Password  string   `json:"password,omitempty"`
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

	users := Users{}
	for results.Next() {
		var user User

		err = results.Scan(&user.AccountID,
			&user.ID,
			&user.Name,
			&user.EMail,
			&user.Role)
		if err != nil {
			return &Users{}, errors.Annotate(err, "error scanning result set")
		}

		users.Users = append(users.Users, &user)
	}

	return &users, nil
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

// InsertUser takes the provided user data, inserts it into the db, and returns the newly created user ID
func InsertUser(db *sql.DB, u User) (int64, error) {
	err := validateUser(u)
	if err != nil {
		return 0, errors.Annotate(err, "User validation failure")
	}

	r, err := db.Exec(insertUserStmt, u.AccountID, u.Name, u.EMail, u.Role, u.Password)
	if err != nil {
		return 0, errors.Annotate(err, fmt.Sprintf("error inserting user into the database: %+v", u))
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, errors.Annotate(err, "error getting user ID")
	}

	return id, nil
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
