// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handlers

import "net/http"

/*
Always create at least a simple 'health' endpoint. A more sophisticated health endpoint could
provide metrics associated number of succesful and unsucessful service requests as well as
perhaps an overall health indication taken from several, service appropriate, metrics.
*/

// HealthFunc returns current health/status of the service
func HealthFunc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("customerd is healthy"))
}
