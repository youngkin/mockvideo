package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"
	"github.com/youngkin/mockvideo/cmd/customerd/logging"
	"github.com/youngkin/mockvideo/internal/platform/config"
	"github.com/youngkin/mockvideo/internal/platform/constants"

	log "github.com/sirupsen/logrus"
)

// TODO:
//	1.	TODO: HTTP connection config (e.g., timeouts)
//  1.  TODO: DB config, use secrets for user name and password
//	4.	TODO: Config parms (configMap?), monitor for changes restarting if necessary
//	6.	TODO: Kube logging
//	7.	TODO: ELK stack for logging
//	10.	TODO: Create build system that will compile and create docker image
//	11.	TODO: Use https
//	5.	ONGOING: Prometheus
//	2.	DONE: Don't panic, add error handling
//	3.	DONE: Use logging levels
//	8.	DONE: Graphana for Prometheus
//	9.	DONE: Helm/Kube deployment (with P9S support)
//	9.	DONE: Travis CI

func main() {
	configFileName := flag.String("configFile",
		"/opt/mockvideo/custd/custd",
		"specifies the location of the custd service configuration")
	flag.Parse()

	logger := logging.GetLogger()

	//
	// Get configuration
	//
	configFile, err := os.Open(*configFileName)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.AppError:       constants.UnableToOpenConfig,
			constants.ErrorDetail:    err.Error(),
		}).Fatal("Error opening config file")
	}
	config, err := config.LoadConfig(configFile)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.AppError:       constants.UnableToLoadConfig,
			constants.ErrorDetail:    err.Error(),
		}).Fatal("Error loading config file")
	}

	//
	// Set configuration
	//
	loglevel, ok := config["logLevel"]
	if !ok {
		logger.Warnf("Log level unavailable, defaulting to %s", log.GetLevel().String())
	} else {
		level, err := strconv.Atoi(loglevel)
		if err != nil {
			logger.Warnf("Log level <%s> invalid, defaulting to %s", loglevel, log.GetLevel().String())
		} else {
			log.SetLevel(log.Level(level))
		}
	}

	port, ok := config["port"]
	if !ok {
		logger.Info("port configuration unavailable (config[port]), defaulting to 5000")
		port = "5000"
	}
	port = ":" + port

	//
	// Setup DB connection
	//
	connStr, err := getDBConnectionStr(config)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.AppError:    constants.UnableToGetDBConnStr,
			constants.ErrorDetail: err.Error(),
		}).Fatal("Error constructing DB connection string")
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.AppError:    constants.UnableToOpenDBConn,
			constants.ErrorDetail: err.Error(),
		}).Fatal("Error opening DB connection")
	}
	defer db.Close()

	//
	// Setup endpoints and start service
	//
	customersHandler, err := handlers.New(db)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.AppError:    constants.UnableToCreateHTTPHandler,
			constants.ErrorDetail: err.Error(),
		}).Fatal("Error creating 'customers' HTTP handler")
	}
	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()

	mux.Handle("/customers", customersHandler)
	mux.Handle("/custdhealth", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	logger.WithFields(log.Fields{
		constants.ConfigFileName: *configFileName,
		constants.Port:           port,
		constants.LogLevel:       log.GetLevel().String(),
	}).Info("customerd service starting")
	logger.Fatal(http.ListenAndServe(port, mux))
}

func getDBConnectionStr(config map[string]string) (string, error) {
	// E.g., "username:userpassword@tcp(10.0.0.100:3306)/mockvideo"
	var sb strings.Builder

	sb.WriteString("admin") // TODO: Replace with secret
	sb.WriteString(":")
	sb.WriteString("2girls1cat") // TODO: Replace with secret
	sb.WriteString("@tcp(")

	dbHost, ok := config["dbHost"]
	if !ok {
		return "", errors.NotAssignedf("DB hostname/address, identified by 'dbHost', not found in configuration")
	}
	sb.WriteString(dbHost)
	sb.WriteString(":")

	dbPort, ok := config["dbPort"]
	if !ok {
		return "", errors.NotAssignedf("DB port, identified by 'dbPort', not found in configuration")
	}
	sb.WriteString(dbPort)

	sb.WriteString(")/")

	dbName, ok := config["dbName"]
	if !ok {
		return "", errors.NotAssignedf("DB Name, identified by 'dbName', not found in configuration")
	}
	sb.WriteString(dbName)

	return sb.String(), nil
}
