// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package domain

import (
	"database/sql"
	"fmt"

	"github.com/juju/errors"
)

// ErrCode is the application type for reporting error codes
type ErrCode int

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

// UserRepository abstracts the notion of some sort of User persistent store
// such as a database of file system.
// TODO: Consider embedding 'ErrCode' inside an application specific error type. This would
// TODO: likely require rethinking how errors are wrapped currently using 'errors.Annotate'
type UserRepository interface {
	GetUsers() (*Users, error)
	GetUser(id int) (*User, error)
	CreateUser(user User) (id int, errCode ErrCode, err error)
	UpdateUser(user User) (ErrCode, error)
	DeleteUser(id int) (ErrCode, error)
}

// User represents the data about a user
type User struct {
	// TODO: Should a User have an accountID? It certainly does in the DB (secondary index).
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
	Users []*User `json:"users"`
}

// IsAuthenticatedUser will return true if the encryptedPassword matches the
// User's real (i.e., unencrypted) password.
// TODO: Move to usecases package/layer
func (u *User) IsAuthenticatedUser(db *sql.DB, id int, encryptedPassword []byte) (bool, error) {
	// TODO: implement
	return false, errors.NewNotImplemented(nil, "Not implemented")
}

// ValidateUser will return an error if the User is not constructed correctly.
func (u *User) ValidateUser() error {
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
