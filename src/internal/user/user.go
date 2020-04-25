// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/youngkin/mockvideo/src/api"
	"github.com/youngkin/mockvideo/src/internal/platform/constants"
)

// DBRqstDur is used to capture the length and status of database requests
// The labels for this metric should be used as follows:
//	1.	'operation' should be one of 'create|update|readAll|readOne|delete'
//	2.	'result' should be one of 'ok|error'
//	3.	'target' refers to the target table name. It should be one of 'user' for now.
//		This must be updated when new tables are added.
var DBRqstDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "mockvideo",
	Subsystem: "database",
	Name:      "db_request_duration_seconds",
	Help:      "database request duration distribution in seconds",
	// Buckets:   prometheus.ExponentialBuckets(0.005, 1.1, 40),
	Buckets: prometheus.LinearBuckets(0.001, .004, 50),
}, []string{"target", "operation", "result"})

// Metrics labels
const (
	create  = "create"
	update  = "update"
	readAll = "readAll"
	readOne = "readOne"
	delete  = "delete"
	ok      = "ok"
	dbErr   = "error"
	userTbl = "userTbl"
)

var (
	getAllUsersQuery = "SELECT accountID, id, name, email, role FROM user"
	getUserQuery     = "SELECT accountID, id, name, email, role FROM user WHERE id = ?"
	insertUserStmt   = "INSERT INTO user (accountID, name, email, role, password) VALUES (?, ?, ?, ?, ?)"
	updateUserStmt   = "UPDATE user SET id = ?, accountID = ?, name = ?, email = ?, role = ?, password = ? WHERE id = ?"
	deleteUserStmt   = "DELETE FROM user WHERE id = ?"
)

// GetAllUsers will return all customers known to the application
func GetAllUsers(db *sql.DB) (*api.Users, error) {
	start := time.Now()

	results, err := db.Query(getAllUsersQuery)
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, readAll, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return &api.Users{}, errors.Annotate(err, "error querying DB")
	}

	us := api.Users{}
	for results.Next() {
		var u api.User

		err = results.Scan(&u.AccountID,
			&u.ID,
			&u.Name,
			&u.EMail,
			&u.Role)
		if err != nil {
			DBRqstDur.WithLabelValues(userTbl, readAll, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
			return &api.Users{}, errors.Annotate(err, "error scanning result set")
		}

		us.Users = append(us.Users, &u)
	}

	DBRqstDur.WithLabelValues(userTbl, readAll, ok).Observe(float64(time.Since(start)) / float64(time.Second))
	return &us, nil
}

// GetUser will return the user identified by 'id' or a nil user if there
// wasn't a matching user.
func GetUser(db *sql.DB, id int) (*api.User, error) {
	start := time.Now()

	row := db.QueryRow(getUserQuery, id)
	user := &api.User{}
	err := row.Scan(&user.AccountID,
		&user.ID,
		&user.Name,
		&user.EMail,
		&user.Role)
	if err != nil && err != sql.ErrNoRows {
		DBRqstDur.WithLabelValues(userTbl, readOne, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, errors.Annotate(err, "error scanning user row")
	}
	if err == sql.ErrNoRows {
		DBRqstDur.WithLabelValues(userTbl, readOne, ok).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil, nil
	}

	DBRqstDur.WithLabelValues(userTbl, readOne, ok).Observe(float64(time.Since(start)) / float64(time.Second))
	return user, nil
}

// InsertUser takes the provided user data, inserts it into the db, and returns the newly created user ID.
func InsertUser(db *sql.DB, u api.User) (int, constants.ErrCode, error) {
	start := time.Now()

	err := validateUser(u)
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, create, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return 0, constants.UserValidationErrorCode, errors.Annotate(err, "User validation failure")
	}

	r, err := db.Exec(insertUserStmt, u.AccountID, u.Name, u.EMail, u.Role, u.Password)
	if err != nil {
		errDetail, ok := err.(*mysql.MySQLError)
		if ok {
			if errDetail.Number == constants.MySQLDupInsertErrorCode {
				DBRqstDur.WithLabelValues(userTbl, create, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
				return 0, constants.DBInsertDuplicateUserErrorCode, errors.Annotate(err, fmt.Sprintf("error inserting duplicate user into the database: %+v, possible duplicate email address", u))
			}
		} else {
			DBRqstDur.WithLabelValues(userTbl, create, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
			return 0, constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error inserting user %+v into DB", u))
		}
	}
	id, err := r.LastInsertId()
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, create, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return 0, constants.DBUpSertErrorCode, errors.Annotate(err, "error getting user ID")
	}

	// TODO: Consider not casting 'id' to an int. Depending on where this code runs, an 'int'
	// TODO: is either 32 or 64 bytes, so this cast *could* be OK
	DBRqstDur.WithLabelValues(userTbl, create, ok).Observe(float64(time.Since(start)) / float64(time.Second))
	return int(id), constants.NoErrorCode, nil
}

// UpdateUser takes the provided user data, inserts it into the db, and returns the newly created user ID
func UpdateUser(db *sql.DB, u api.User) (constants.ErrCode, error) {
	start := time.Now()

	err := validateUser(u)
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, update, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBUpSertErrorCode, errors.Annotate(err, "User validation failure")
	}

	// This entire db.Begin/tx.Rollback/Commit seem awkward to me. But it's here because
	// MySQL silently performs an insert if there is no row to update.
	tx, err := db.Begin()
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, update, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error beginning transaction for user: %+v", u))
	}
	r := db.QueryRow(getUserQuery, u.ID)
	userRow := api.User{}
	err = r.Scan(&userRow.AccountID,
		&userRow.ID,
		&userRow.Name,
		&userRow.EMail,
		&userRow.Role)

	if err != nil && err == sql.ErrNoRows {
		tx.Rollback()
		DBRqstDur.WithLabelValues(userTbl, update, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBInvalidRequestCode, errors.New(fmt.Sprintf("error, attempting to update non-existent user, user.ID %d", u.ID))
	}
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		DBRqstDur.WithLabelValues(userTbl, update, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating user in the database: %+v", u))
	}

	_, err = db.Exec(updateUserStmt, u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID)
	if err != nil {
		tx.Rollback()
		DBRqstDur.WithLabelValues(userTbl, update, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating user in the database: %+v", u))
	}
	tx.Commit()

	DBRqstDur.WithLabelValues(userTbl, update, ok).Observe(float64(time.Since(start)) / float64(time.Second))
	return constants.NoErrorCode, nil
}

// DeleteUser deletes the user identified by u.id from the database
func DeleteUser(db *sql.DB, id int) (constants.ErrCode, error) {
	start := time.Now()

	_, err := db.Exec(deleteUserStmt, id)
	if err != nil {
		DBRqstDur.WithLabelValues(userTbl, delete, dbErr).Observe(float64(time.Since(start)) / float64(time.Second))
		return constants.DBDeleteErrorCode, errors.Annotate(err, "Usesr delete error")
	}

	DBRqstDur.WithLabelValues(userTbl, delete, ok).Observe(float64(time.Since(start)) / float64(time.Second))
	return constants.NoErrorCode, nil
}

// IsAuthorizedUser will return true if the encryptedPassword matches the
// User's real (i.e., unencrypted) password.
func IsAuthorizedUser(db *sql.DB, id int, encryptedPassword []byte) (bool, error) {
	// TODO: implement
	return false, errors.NewNotImplemented(nil, "Not implemented")
}

func validateUser(u api.User) error {
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
	if u.Role != api.Primary && u.Role != api.Restricted && u.Role != api.Unrestricted {
		errMsg = errMsg + fmt.Sprintf("; Invalid Role. Role must be one of %d, %d, or %d, got %d",
			api.Primary, api.Restricted, api.Unrestricted, u.Role)
	}

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
