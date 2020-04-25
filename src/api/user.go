// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

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
	Users []*User `json:"users"`
}
