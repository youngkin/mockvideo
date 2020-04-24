// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/youngkin/mockvideo/src/cmd/accountd/handlers"
	"github.com/youngkin/mockvideo/src/cmd/accountd/handlers/users"
	"github.com/youngkin/mockvideo/src/internal/platform/config"
	"github.com/youngkin/mockvideo/src/internal/platform/constants"
	"github.com/youngkin/mockvideo/src/internal/platform/logging"

	log "github.com/sirupsen/logrus"
)

/*
This file, the 'main' function in particular, attempt to convey some best practices
pertaining to:

1. 	Obtaining configuration via command line flags and from the project's common 'config' capability.
2.	Using structured logging for use with log view/search apps like ELK and Splunk
3.	HTTP service configuration related to gracefully handling slow or unresponsive clients (e.g., write timeout)
3.	Graceful shutdown in response to SIGTERM
4.	Use of a MySQL 'database.sql.driver' implementation
	i.	Uses 'interpolateParams=true' to avoid multiple round-trips when using placeholders (i.e., '?') in a
		`db.Query()` or `db.Exec()` call
	ii.	Uses 'parseTime=true' to allow unmarshaling DATE DATETIME directly into Golang time.Time variables.
*/

// TODO:
//	1.	TODO: Load limiting (e.g., leaky bucket, and/or Itsio/Traefik?)
//	2.	TODO: Circuit breakers on DB
//	11.	TODO: Use https
//	4.	TODO: Config parms (configMap?), monitor for changes restarting if necessary
//	5.	TODO: ONGOING: Prometheus, instrument database calls

func main() {
	configFileName := flag.String("configFile",
		"/opt/mockvideo/accountd/config/config",
		"specifies the location of the accountd service configuration")
	secretsDir := flag.String("secretsDir",
		"/opt/mockvideo/accountd/secrets",
		"specifies the location of the accountd secrets")
	flag.Parse()

	logger := logging.GetLogger().WithField(constants.Application, constants.User)

	//
	// Get configuration
	//
	configFile, err := os.Open(*configFileName)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.ErrorCode:      constants.UnableToOpenConfigErrorCode,
			constants.ErrorDetail:    err.Error(),
		}).Fatal(constants.UnableToOpenConfig)
		os.Exit(1)
	}

	configs, err := config.LoadConfig(configFile)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.ErrorCode:      constants.UnableToLoadConfigErrorCode,
			constants.ErrorDetail:    err.Error(),
		}).Fatal(constants.UnableToLoadConfig)
		os.Exit(1)
	}

	secrets, err := config.LoadSecrets(*secretsDir)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *secretsDir,
			constants.ErrorCode:      constants.UnableToLoadSecretsErrorCode,
			constants.ErrorDetail:    err.Error(),
		}).Fatal(constants.UnableToLoadSecrets)
		os.Exit(1)
	}

	loglevel, ok := configs["logLevel"]
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

	logger.WithFields(log.Fields{
		constants.ConfigFileName: *configFileName,
		constants.SecretsDirName: *secretsDir,
	}).Info("accountd service starting")

	//
	// Setup DB connection
	//
	connStr, err := getDBConnectionStr(configs, secrets)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToGetDBConnStrErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Fatal(constants.UnableToGetDBConnStr)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToOpenDBConnErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Fatal(constants.UnableToOpenDBConn)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToOpenDBConnErrorCode,
			constants.ErrorDetail: err.Error(),
			constants.DBHost:      configs["dbHost"],
			constants.DBPort:      configs["dbPort"],
			constants.DBName:      configs["dbName"],
		}).Fatal(constants.UnableToOpenDBConn + ": database unreachable")
		os.Exit(1)
	}

	maxBulkOps := 10
	maxBulkOpsStr, ok := configs["maxConcurrentBulkOperations"]
	if !ok {
		logger.Info("max bulk operations configuration unavailable (configs[maxConcurrentBulkOperations]), defaulting to 10")
		maxBulkOpsStr = "10"
	} else {
		maxBulkOps, err = strconv.Atoi(maxBulkOpsStr)
		if err != nil {
			logger.Warnf("maxConcurrentBulkOperations <%s> invalid, defaulting to %d", maxBulkOpsStr, maxBulkOps)
		}
	}

	//
	// Setup endpoints and start service
	//
	usersHandler, err := users.NewUserHandler(db, logger, maxBulkOps)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToCreateHTTPHandlerErrorCode,
			constants.ErrorDetail: err.Error(),
			constants.DBHost:      configs["dbHost"],
			constants.DBPort:      configs["dbPort"],
			constants.DBName:      configs["dbName"],
		}).Fatal(constants.UnableToCreateHTTPHandler)
		os.Exit(1)
	}

	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()
	mux.Handle("/users", usersHandler)  // Desired to prevent redirects. Can remove if redirects for '/users/' are OK
	mux.Handle("/users/", usersHandler) // Required to properly route requests to '/users/{id}. Don't understand why the above route isn't sufficient
	mux.Handle("/accountdhealth", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.ErrorDetail: errors.New(constants.MalformedURL),
		}).Info("handling request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.MalformedURL))
	})

	port, ok := configs["port"]
	if !ok {
		logger.Info("port configuration unavailable (configs[port]), defaulting to 5000")
		port = "5000"
	}
	port = ":" + port

	s := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	go func() {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.SecretsDirName: *secretsDir,
			constants.Port:           port,
			constants.LogLevel:       log.GetLevel().String(),
			constants.DBHost:         configs["dbHost"],
			constants.DBPort:         configs["dbPort"],
			constants.DBName:         configs["dbName"],
		}).Info("accountd service running")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	handleTermSignal(s, logger, 10)
}

//
// Helper funcs
//

// handleTermSignal provides a mechanism to catch SIGTERMs and gracefully
// shutdown the service.
func handleTermSignal(s *http.Server, logger *log.Entry, timeout int) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	logger.Infof("Server shutting down with timeout: %d", timeout)

	if err := s.Shutdown(ctx); err != nil {
		logger.Warnf("Server shutting down with error: %s", err)
	} else {
		logger.Info("Server stopped")
	}

}

func getDBConnectionStr(configs, secrets map[string]string) (string, error) {
	// E.g., "username:userpassword@tcp(10.0.0.100:3306)/mockvideo?interpolateParams=true"
	var sb strings.Builder

	dbuser, ok := secrets["dbuser"]
	if !ok {
		return "", errors.NotAssignedf("DB user name, identified by 'dbuser', not found in secrets")
	}
	sb.WriteString(dbuser)
	sb.WriteString(":")
	dbpassword, ok := secrets["dbpassword"]
	if !ok {
		return "", errors.NotAssignedf("DB user password, identified by 'dbpassword', not found in secrets")
	}
	sb.WriteString(dbpassword)
	sb.WriteString("@tcp(")

	dbHost, ok := configs["dbHost"]
	if !ok {
		return "", errors.NotAssignedf("DB hostname/address, identified by 'dbHost', not found in configuration")
	}
	sb.WriteString(dbHost)
	sb.WriteString(":")

	dbPort, ok := configs["dbPort"]
	if !ok {
		return "", errors.NotAssignedf("DB port, identified by 'dbPort', not found in configuration")
	}
	sb.WriteString(dbPort)

	sb.WriteString(")/")

	dbName, ok := configs["dbName"]
	if !ok {
		return "", errors.NotAssignedf("DB Name, identified by 'dbName', not found in configuration")
	}
	sb.WriteString(dbName)

	sb.WriteString("?interpolateParams=true")

	return sb.String(), nil
}
