// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logging

//
// **NOTE** When adding constants, please order them in alphabetical sequence
//

//
// All constants below refer to the standard names for the various
// fields used in log messages.
//
const (
	Application    string = "Application"
	ConfigFileName string = "ConfigFileName"

	DBHost string = "DBHost"
	DBName string = "DBName"
	DBPort string = "DBPort"

	ErrorCode   string = "ErrorCode"
	ErrorDetail string = "ErrorDetail"

	HostName   string = "HostName"
	HTTPStatus string = "HTTPStatus"

	LogLevel      string = "LogLevel"
	Method        string = "HTTPMethod"
	MessageDetail string = "MsgDetail"

	Path string = "URLPath"
	Port string = "Port"

	RemoteAddr     string = "RemoteAddr"
	RPCFunc        string = "RPCFunc"
	ServiceName    string = "ServiceName"
	SecretsDirName string = "SecretsDirName"

	UserID string = "UserID"

	TestName string = "TestName"
)

const (
	// User is the standard name for the user service
	User string = "user"
)
