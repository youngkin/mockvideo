// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package constants

const (
	//
	// Miscellaneous errors
	//

	// UnableToOpenConfig indicates there was a problem opening the configuration file
	UnableToOpenConfig = "Unable to open configuration file"
	// UnableToGetConfig indicates there was a problem obtaining the application configuration
	UnableToGetConfig = "Unable to get information from configuration"
	// UnableToLoadConfig indicates there was a problem loading the configuration
	UnableToLoadConfig = "Unable to load configuration"
	// UnableToLoadSecrets indicates there was a problem loading the application's secrets
	UnableToLoadSecrets = "Unable to load secrets"

	// UnableToGetDBConnStr indicates there was a problem constructing a DB connection string
	UnableToGetDBConnStr = "Unable to get DB connection string"
	// UnableToOpenDBConn indicates there was a problem opening a database connection
	UnableToOpenDBConn = "Unable to open DB connection"
	// DBRowScanError indicates results from DB query could not be processed
	DBRowScanError = "DB resultset processing failed"
	// DBQueryError indications that there was a problem executing a DB query operation
	DBQueryError = "DB query failed"

	// UnableToCreateHTTPHandler indications that there was a problem creating an http handler
	UnableToCreateHTTPHandler = "Unable to create HTTP handler"

	// JSONMarshalingError indicates that there was a problem un/marshaling JSON
	JSONMarshalingError = "JSON Marshaling Error"
	// MalformedURL indicates there was a problem with the structure of the URL
	MalformedURL = "Malformed URL"

	//
	// User related error codes start at 1000 and go to 1999
	//

	// UserTypeConversionError indicates that the payload returned from GET /users/{id} could
	// not be converted to either a Users (/users) or User (/users/{id}) type
	UserTypeConversionError = "Unable to convert payload to User(s) type"
	// UserGETError indicates that GET /users or GET /users/{id} failed in some way
	UserGETError = "GET /users or GET /users/{id} failed"
)

const (
	//
	// Misc related error codes start at 0 and go to 99
	//

	// UnableToOpenConfigErrorCode is the error code associated with UnableToOpenConfig
	UnableToOpenConfigErrorCode = iota
	// UnableToGetConfigErrorCode is the error code associated with UnableToGetConfig
	UnableToGetConfigErrorCode
	// UnableToLoadConfigErrorCode is the error code associated with UnableToLoadConfig
	UnableToLoadConfigErrorCode
	// UnableToLoadSecretsErrorCode is the error code associated with UnableToLoadSecrets
	UnableToLoadSecretsErrorCode
	// UnableToGetDBConnStrErrorCode is the error code associated with UnableToGetDBConnStr
	UnableToGetDBConnStrErrorCode
	// UnableToOpenDBConnErrorCode is the error code associated with UnableToOpenDBConn
	UnableToOpenDBConnErrorCode
	// DBRowScanErrorCode is the error code associated with DBRowScan
	DBRowScanErrorCode
	// DBQueryErrorCode is the error code associated with DBQueryError
	DBQueryErrorCode
	// UnableToCreateHTTPHandlerErrorCode is the error code associated with UnableToCreateHTTPHandler
	UnableToCreateHTTPHandlerErrorCode
	// JSONMarshalingErrorCode is the error code associated with JSONMarshaling
	JSONMarshalingErrorCode
	// MalformedURLErrorCode is the error code associated with MalformedURL
	MalformedURLErrorCode
)

const (
	//
	// User related error codes start at 1000 and go to 1999
	//

	// UserTypeConversionErrorCode is the error code associated with UserTypeConversion
	UserTypeConversionErrorCode = iota + 1000
	// UserGETErrorCode is the error code associated with UserGETError
	UserGETErrorCode
)
