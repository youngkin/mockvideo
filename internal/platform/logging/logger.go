package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/platform/constants"
)

var logger *log.Entry

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
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
