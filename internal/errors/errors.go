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

// MySQLDupInsertErrorCode is an alias for the MySQL error code for duplicate row insert attempt
const MySQLDupInsertErrorCode = 1062

// ErrCode is the application type for reporting error codes
type ErrCode int

//
// ---------------------- Errors ---------------------
//

// MVError is the MockVideo type for application specific errors
type MVError struct {
	ErrCode    ErrCode
	ErrMsg     string
	ErrDetail  string
	WrappedErr error
}

func (e *MVError) Error() string {
	return fmt.Sprintf("ErrorCode: %d, ErrMsg: %s, ErrDetail: %s, WrappedErr: %s", e.ErrCode, e.ErrMsg, e.ErrDetail, e.WrappedErr)
}

//
// ---------------------- Miscellaneous error messages ------------------------------
//
const (
	// BulkRequestErrorMsg provides information about a failed bulk request
	BulkRequestErrorMsg = "an error occurred during a bulk request operation"

	// DBDeleteErrorMsg is an indication of a DB error during a DELETE operation
	DBDeleteErrorMsg = "a DB error occurred during a DELETE operation"
	// DBInsertDuplicateUserErrorMsg indicates an attempt to insert a duplicate row
	DBInsertDuplicateUserErrorMsg = "attempt to insert duplicate user"
	// DBNoUserErrorMsg indicates that the requested user could not be found in the DB
	DBNoUserErrorMsg = "User not found"
	// DBRowScanErrorMsg indicates results from DB query could not be processed
	DBRowScanErrorMsg = "DB resultset processing failed"
	// DBUpSertErrorMsg indicates that there was a problem executing a DB insert or update operation
	DBUpSertErrorMsg = "DB insert or update failed"

	// HTTPWriteErrorMsg indicates that there was a problem writing an HTTP response body
	HTTPWriteErrorMsg = "Error writing HTTP response body"

	// InvalidInsertErrorMsg indicates that an unexpected User.ID was detected in an insert request
	InvalidInsertErrorMsg = "Unexpected User.ID in insert request"
	// InvalidProtocolTypeErrorMsg indicates that an invalid protocol was specified (e.g., not 'http' or 'grpc')
	InvalidProtocolTypeErrorMsg = "Invalid protocol type specified at application startup, must be 'http' or 'grpc'"

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
	UnableToCreateHTTPHandlerMsg = "Unable to create HTTP service endpoint"
	// UnableToCreateRepositoryMsg indicates that there was a problem creating Repository instance referencing
	// storage for an application domain type
	UnableToCreateRepositoryMsg = "Unable to create domain object repository"
	// UnableToCreateRPCServerErrorMsg indicates there was a problem creating a gRPC endpoint
	UnableToCreateRPCServerErrorMsg = "Unable to create gRPC service endpoint"
	// UnableToCreateUserSvcMsg indicates that there was a problem creating an application use case
	UnableToCreateUserSvcMsg = "Unable to create UserService"
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
// ---------------------- User related error messages --------------
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

//
// ---------------------- Error codes -------------------------------
//

const (
	//
	// Misc related error codes start at 0 and go to 99
	//

	// NoErrorCode is a placeholder indicating no error has occurred.
	NoErrorCode ErrCode = iota

	// UnknownErrorCode is applied when unexpected errors occur and none of the other error codes apply
	UnknownErrorCode

	// BulkRequestErrorCode indicates there was a problem with a bulk request (CREATE or UPDATE)
	BulkRequestErrorCode

	// DBDeleteErrorCode indication of a DB error during a DELETE operation
	DBDeleteErrorCode
	// DBInsertDuplicateUserErrorCode indicates an attempt to insert a duplicate row
	DBInsertDuplicateUserErrorCode
	// DBInvalidRequestCode indication of an invalid request, e.g., an update was attempted on an existing user
	DBInvalidRequestCode
	// DBNoUserErrorCode indicates an invalid DB request, like attempting to update a non-existent user
	DBNoUserErrorCode
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
	// InvalidProtocolTypeErrorCode indicates that an invalid protocol was specified (e.g., not 'html' or 'grpc')
	InvalidProtocolTypeErrorCode

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
	// UnableToCreateRPCServerErrorCode is the error code associated with UnableToCreateRPCServerErrorMsg
	UnableToCreateRPCServerErrorCode
	// UnableToCreateUserSvcErrorCode is the error code associated with UnableToCreateUseCase
	UnableToCreateUserSvcErrorCode
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
