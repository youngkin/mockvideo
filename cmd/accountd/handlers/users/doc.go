// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package users contains the implementations of the HTTP handlers for the 'users' resource. There
are several resources that are part of an account. These include the users that are part of an account,
and the account that has one or more associated users.

The code in this package attempts to showcase several best practices including:

1.	Structured logging (i.e., 'logger.WithFields(...)')
2.	Log 'hygiene'. Lower level functions don't log. The do return
	errors when necessary and allow the calling function to decide
	if it wants to log the error or propagate the error up the stack.
3.	Error handling:
	i.		Early returns
	ii.		Use of error codes vs. text strings
	iii.	Addition of info to errors to help better understand the
			context the error occurred in.
4. 	Request validation - e.g., verify proper URL path construction
5.	Proper use of HTTP status codes
6.	Use of Prometheus to capture metrics

The tests and supporting code demonstrate the following:

1.  Table driven tests using 'Tests' and 'CustTests' structs and appropriate
	test instance definitions using struct literals in each test function
2.	Sub-tests. These are useful to get more detailed information from your test
	executions.
3.	Use of unit test "helper" functions that can fail tests on behalf of calling code.
	See [Go Advanced Testing Tips](https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859) for details.



The 'users' sub-package will do basic CRUD operations on the users resource. Here are the associated
resource URLs (prepended with '/accountd'):

		/users
		/users/{id}

Supported HTTP Verbs:

		GET, PUT (update), POST (create), DELETE

All verbs operate on a 'User' object encouded in JSON. It has the following structure:

		{
			id: {int}
			accountid: {int}
			name: {string}
			email: {string}
			role: {int} // Valid values for 'role' are 0 (primary), 1 (unrestricted), 2 (restricted)
			password: {string}
		}

Here's an example of the above:

		{
			id: 1
			accountid: 1
			name: "Brian Wilson"
			email: "goodvibrations@beachboys.com"
			role: 1 // Valid values for 'role' are 0 (primary), 1 (unrestricted), 2 (restricted)
			password: "helpmerhonda"
		}

The following paragraphs show example requests. Note that the 'hostname' of each request is 'accountd.kube'. This
matches the Kubernetes Ingress definition for this service. Other deployments may have different 'hostname's.

PUT' and 'POST' operations take a JSON body that describes the 'User' to be created (POST) or
updated (PUT). Here's an example POST request:

		curl -i -X POST http://accountd.kube/users -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

Note that the resource path for POST is '/users' and that the JSON doesn't include an 'id' field. Having the path include a
resource id and the JSON having an 'id' field are indicative of an update and should be handled via PUT. Including that information
in a POST request will result in a '400' (BadRequest) status. The 'id' of the newly created User can be found in the "Location" header
field. HTTP status of 201 indicates that the resource was successfully created.

Here's an example of a PUT request:

		curl -i -X PUT http://accountd.kube/users/1 -H "Content-Type: application/json" -d "{\"id\":1,\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"inmyroom\"}"

A successful request will result in an HTTP status of 200.

Here's are examples of GET requests (the second requests a 'User' identified by '1')

		curl -i http://accountd.kube/users
		curl -i http://accountd.kube/users/1

JSON similar to the example above will be returned in the response body with a status 200.

Here's an example of a DELETE request:

		curl -i -X DELETE http://accountd.kube/users/1

A 200 HTTP status indicates a successful result.

Other HTTP status codes indicate various errors. These are:

1. 400 Bad Request - This indicates there was a problem with the request and it was not accepted. These request should not be retried.
2. 404 Not Found - This indicates that the requested user, whether for GET or PUT, could not be found
3. 500 Internal Server Error - This indicates that there was a problem with the server fulfilling the request. It does not indicate that the request was invalid. It's possible the problem could be resolved if the request is retried.
4. 501 Not Implemented - The request is not supported (e.g., a HEAD request).
5. 502 Bad Gateway - This is not returned directly by the service. It is returned by an upstream proxy or Kubernetes ingress. The request can be retried.
6. 503 Service Unavailable - This may be returned if the server is overloaded. If so, there will beha a 'Retry-After' header indicating how much time should pass before the request is retried.
*/
package users
