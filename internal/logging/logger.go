// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/constants"
)

var logger *log.Entry

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Only log the DEBUG severity or above.
	log.SetLevel(log.InfoLevel)
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}

	logger = log.WithFields(log.Fields{
		constants.HostName: hostName,
	})

}

// GetLogger gets the common logger for the customer service
func GetLogger() *log.Entry {
	return logger
}
