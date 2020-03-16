// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handlers

import "net/http"

// HealthFunc returns current health/status of the service
func HealthFunc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("customerd is healthy"))
}
