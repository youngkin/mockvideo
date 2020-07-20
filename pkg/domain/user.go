// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package domain

import (
	"github.com/youngkin/mockvideo/internal/domain"
)

// User represents the data about a user
type User struct {
	// TODO: Should a User have an accountID? It certainly does in the DB (secondary index).
	AccountID int         `json:"accountid"`
	HREF      string      `json:"href"`
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	EMail     string      `json:"email"`
	Role      domain.Role `json:"role"`
}

// Users is a collection (slice) of User
type Users struct {
	Users []*User `json:"users"`
}
