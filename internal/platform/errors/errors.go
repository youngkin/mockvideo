package errors

const (
	// BadConfigFileExit indicates there was a problem opening the configuration file
	BadConfigFileExit = iota + 1
	// UnableToGetConfigExit indicates there was a problem obtaining the application configuration
	UnableToGetConfigExit
	// BadDBConnectionExit indicates there was a problem obtaining a database connection
	BadDBConnectionExit
	// UnableToGetPortConfigExit indicates the port configuration was unavailable
	UnableToGetPortConfigExit
)
