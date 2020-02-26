package constants

// TODO:
//	1.	Refactor into errors with text and numeric code?

const (
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

	// UnableToCreateHTTPHandler indications that there was a problem creating an http handler
	UnableToCreateHTTPHandler = "Unable to create HTTP handler"
	// DBQueryError indications that there was a problem executing a DB query operation
	DBQueryError = "DB query failed"

	// JSONMarshalingError indicates that there was a problem un/marshaling JSON
	JSONMarshalingError = "JSON Marshaling Error"
)
