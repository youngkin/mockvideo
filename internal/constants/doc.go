// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package constants provides common, constant, definitions for things like error codes and log field names.

As a whole this package attempts to demonstrate best practices around things like:

1. 	Use of constants for common string values used throughout the application (e.g., log fields)
2.	Definition of error codes and strings to be used throughout the application. The point of common
	error codes and error strings (unformatted) is to enable searching in log aggregators like
	ELK and Splunk.
*/
package constants
