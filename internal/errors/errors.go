// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//
// **NOTE** When adding errors, please order them in alphabetical sequence
//

package errors

import (
	"fmt"
)

// ErrCode is the application type for reporting error codes
type ErrCode int

// MVError is the MockVideo type for application specific errors
type MVError struct {
	ErrCode    ErrCode
	ErrMsg     string
	ErrDetail  string
	WrappedErr error
}

func (e MVError) Error() string {
	return fmt.Sprintf("ErrorCode: %d, ErrMsg: %s, ErrDetail: %s, WrappedErr: %s", e.ErrCode, e.ErrMsg, e.ErrDetail, e.WrappedErr)
}

// UserRqstError indicates there was a problem with a query for a user or users
type UserRqstError struct {
	MVError
}

func (e UserRqstError) Error() string {
	return e.MVError.Error()
}

// MalformedURLError indicates an HTTP request had an invalid URL
type MalformedURLError struct {
	MVError
}

func (e MalformedURLError) Error() string {
	return e.MVError.Error()
}

// MySQLDupInsertErrorCode is an alias for the MySQL error code for duplicate row insert attempt
const MySQLDupInsertErrorCode = 1062

//
// Miscellaneous errors
//
const (
	// DBDeleteErrorMsg is an indication of a DB error during a DELETE operation
	DBDeleteErrorMsg = "a DB error occurred during a DELETE operation"
	// DBInsertDuplicateUserErrorMsg indicates an attempt to insert a duplicate row
	DBInsertDuplicateUserErrorMsg = "attempt to insert duplicate user"
	// DBNoUserErrorMsg indicates an invalid DB request, like attempting to update a non-existent user
	DBNoUserErrorMsg = "attempted update on a non-existent user"
	// DBRowScanErrorMsg indicates results from DB query could not be processed
	DBRowScanErrorMsg = "DB resultset processing failed"
	// DBUpSertErrorMsg indications that there was a problem executing a DB insert or update operation
	DBUpSertErrorMsg = "DB insert or update failed"

	// HTTPWriteErrorMsg indicates that there was a problem writing an HTTP response body
	HTTPWriteErrorMsg = "Error writing HTTP response body"

	// InvalidInsertErrorMsg indicates that an unexpected User.ID was detected in an insert request
	InvalidInsertErrorMsg = "Unexpected User.ID in insert request"

	// JSONDecodingErrorMsg indicates that there was a problem decoding JSON input
	JSONDecodingErrorMsg = "JSON Decoding Error, possibly malformed JSON object"
	// JSONMarshalingErrorMsg indicates that there was a problem un/marshaling JSON
	JSONMarshalingErrorMsg = "JSON Marshaling Error"

	// MalformedURLMsg indicates there was a problem with the structure of the URL
	MalformedURLMsg = "Malformed URL, URL must be of the form /users, /users/{id}, /accountdhealth, or /metrics"

	// RqstParsingErrorMsg indicates that an error occurred while the path and/or body of the was
	// being evaluated.
	RqstParsingErrorMsg = "Request parsing error, possible malformed JSON"

	// UnableToCreateHTTPHandlerMsg indicates that there was a problem creating an http handler
	UnableToCreateHTTPHandlerMsg = "Unable to create HTTP handler"
	// UnableToCreateRepositoryMsg indicates that there was a problem creating Repository instance referencing
	// storage for an application domain type
	UnableToCreateRepositoryMsg = "Unable to create domain object repository"
	// UnableToCreateUseCaseMsg indicates that there was a problem creating an application use case
	UnableToCreateUseCaseMsg = "Unable to create application use case"
	// UnableToGetConfigMsg indicates there was a problem obtaining the application configuration
	UnableToGetConfigMsg = "Unable to get information from configuration"
	// UnableToGetDBConnStrMsg indicates there was a problem constructing a DB connection string
	UnableToGetDBConnStrMsg = "Unable to get DB connection string"
	// UnableToLoadConfigMsg indicates there was a problem loading the configuration
	UnableToLoadConfigMsg = "Unable to load configuration"
	// UnableToLoadSecretsMsg indicates there was a problem loading the application's secrets
	UnableToLoadSecretsMsg = "Unable to load secrets"
	// UnableToOpenConfigMsg indicates there was a problem opening the configuration file
	UnableToOpenConfigMsg = "Unable to open configuration file"
	// UnableToOpenDBConnMsg indicates there was a problem opening a database connection
	UnableToOpenDBConnMsg = "Unable to open DB connection"

	// UnknownErrorMsg is needed when none of the other defined errors apply
	UnknownErrorMsg = "unexpected error occurred"
)

//
// User related error codes start at 1000 and go to 1999
//
const (
	// UserRqstErrorMsg indicates that GET(or PUT) /users or GET(or PUT) /users/{id} failed in some way
	UserRqstErrorMsg = "GET /users or GET /users/{id} failed"
	// UserTypeConversionErrorMsg indicates that the payload returned from GET /users/{id} could
	// not be converted to either a Users (/users) or User (/users/{id}) type
	UserTypeConversionErrorMsg = "Unable to convert payload to User(s) type"
	// UserValidationErrorMsg indicates a problem with the User data
	UserValidationErrorMsg = "invalid user data"
)

const (
	//
	// Misc related error codes start at 0 and go to 99
	//

	// UnknownErrorCode is applied when unexpected errors occur and none of the other error codes apply
	UnknownErrorCode ErrCode = iota

	// DBDeleteErrorCode indication of a DB error during a DELETE operation
	DBDeleteErrorCode
	// DBInsertDuplicateUserErrorCode indicates an attempt to insert a duplicate row
	DBInsertDuplicateUserErrorCode
	// DBInvalidRequestCode indication of an invalid request, e.g., an update was attempted on an existing user
	DBInvalidRequestCode
	// DBQueryErrorCode is the error code associated with DBQueryError
	DBQueryErrorCode
	// DBRowScanErrorCode is the error code associated with DBRowScan
	DBRowScanErrorCode
	// DBUpSertErrorCode indications that there was a problem executing a DB insert or update operation
	DBUpSertErrorCode

	// HTTPWriteErrorCode indicates that there was a problem writing an HTTP response body
	HTTPWriteErrorCode

	// InvalidInsertErrorCode is the error code associated with InvalidInsertError
	InvalidInsertErrorCode

	// JSONDecodingErrorCode indicates that there was a problem decoding JSON input
	JSONDecodingErrorCode
	// JSONMarshalingErrorCode is the error code associated with JSONMarshaling
	JSONMarshalingErrorCode

	// MalformedURLErrorCode is the error code associated with MalformedURL
	MalformedURLErrorCode

	// RqstParsingErrorCode is the error code associated with RqstParsingErrorCode
	RqstParsingErrorCode

	// UnableToCreateHTTPHandlerErrorCode is the error code associated with UnableToCreateHTTPHandler
	UnableToCreateHTTPHandlerErrorCode
	// UnableToCreateRepositoryErrorCode indicates that there was a problem creating Repository instance referencing
	// storage for an application domain type
	UnableToCreateRepositoryErrorCode
	// UnableToCreateUseCaseErrorCode is the error code associated with UnableToCreateUseCase
	UnableToCreateUseCaseErrorCode
	// UnableToGetConfigErrorCode is the error code associated with UnableToGetConfig
	UnableToGetConfigErrorCode
	// UnableToGetDBConnStrErrorCode is the error code associated with UnableToGetDBConnStr
	UnableToGetDBConnStrErrorCode
	// UnableToLoadConfigErrorCode is the error code associated with UnableToLoadConfig
	UnableToLoadConfigErrorCode
	// UnableToLoadSecretsErrorCode is the error code associated with UnableToLoadSecrets
	UnableToLoadSecretsErrorCode
	// UnableToOpenConfigErrorCode is the error code associated with UnableToOpenConfig
	UnableToOpenConfigErrorCode
	// UnableToOpenDBConnErrorCode is the error code associated with UnableToOpenDBConn
	UnableToOpenDBConnErrorCode
)

const (
	//
	// User related error codes start at 1000 and go to 1999
	//

	// UserRqstErrorCode is the error code associated with UserRqstErrorCode
	UserRqstErrorCode ErrCode = iota + 1000
	// UserTypeConversionErrorCode is the error code associated with UserTypeConversion
	UserTypeConversionErrorCode
	// UserValidationErrorCode indicates a problem with the User data
	UserValidationErrorCode
)
