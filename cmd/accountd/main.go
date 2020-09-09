// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpcuser "github.com/youngkin/mockvideo/cmd/accountd/grpc/users"
	handlers "github.com/youngkin/mockvideo/cmd/accountd/http"
	"github.com/youngkin/mockvideo/cmd/accountd/http/users"
	"github.com/youngkin/mockvideo/cmd/accountd/internal/config"
	"github.com/youngkin/mockvideo/cmd/accountd/services"
	"github.com/youngkin/mockvideo/internal/db"
	userdb "github.com/youngkin/mockvideo/internal/db"
	mverr "github.com/youngkin/mockvideo/internal/errors"
	"github.com/youngkin/mockvideo/internal/logging"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
	"google.golang.org/grpc"

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

func init() {
	// PROMETHEUS NOTE:
	// As metrics get defined, e.g., such as 'users.UserRqstDur', they must be
	// added here. 'prometheus.MustRegister()' can only be called once at
	// program initialization. Metrics should be defined in the packages that
	// use them.
	prometheus.MustRegister(users.UserRqstDur, db.DBRqstDur)
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func main() {
	configFileName := flag.String("configFile",
		"/opt/mockvideo/accountd/config/config",
		"specifies the location of the accountd service configuration")
	secretsDir := flag.String("secretsDir",
		"/opt/mockvideo/accountd/secrets",
		"specifies the location of the accountd secrets")
	protocolType := flag.String("protocol", "http", "specifies whether the service will use http or grpc. Options are 'http' or 'grpc'.")
	flag.Parse()

	logger := logging.GetLogger().WithField(logging.Application, logging.User)

	//
	// Get configuration
	//
	configFile, err := os.Open(*configFileName)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.ErrorCode:      mverr.UnableToOpenConfigErrorCode,
			logging.ErrorDetail:    err.Error(),
		}).Fatal(mverr.UnableToOpenConfigMsg)
		os.Exit(1)
	}

	configs, err := config.LoadConfig(configFile)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.ErrorCode:      mverr.UnableToLoadConfigErrorCode,
			logging.ErrorDetail:    err.Error(),
		}).Fatal(mverr.UnableToLoadConfigMsg)
		os.Exit(1)
	}

	secrets, err := config.LoadSecrets(*secretsDir)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ConfigFileName: *secretsDir,
			logging.ErrorCode:      mverr.UnableToLoadSecretsErrorCode,
			logging.ErrorDetail:    err.Error(),
		}).Fatal(mverr.UnableToLoadSecretsMsg)
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
		logging.ConfigFileName: *configFileName,
		logging.SecretsDirName: *secretsDir,
	}).Info("accountd service starting")

	//
	// Setup DB connection
	//
	connStr, err := getDBConnectionStr(configs, secrets)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.UnableToGetDBConnStrErrorCode,
			logging.ErrorDetail: err.Error(),
		}).Fatal(mverr.UnableToGetDBConnStrMsg)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.UnableToOpenDBConnErrorCode,
			logging.ErrorDetail: err.Error(),
			logging.DBHost:      configs["dbHost"],
			logging.DBPort:      configs["dbPort"],
			logging.DBName:      configs["dbName"],
		}).Fatal("OPEN: " + mverr.UnableToOpenDBConnMsg)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.UnableToOpenDBConnErrorCode,
			logging.ErrorDetail: err.Error(),
			logging.DBHost:      configs["dbHost"],
			logging.DBPort:      configs["dbPort"],
			logging.DBName:      configs["dbName"],
		}).Fatal("PING: " + mverr.UnableToOpenDBConnMsg)
		os.Exit(1)
	}

	maxBulkOps := 10
	maxBulkOpsStr, ok := configs["maxConcurrentBulkOperations"]
	if !ok {
		logger.Info("max bulk operations configuration unavailable (configs[maxConcurrentBulkOperations]), defaulting to 10")
	} else {
		maxBulkOps, err = strconv.Atoi(maxBulkOpsStr)
		if err != nil {
			logger.Warnf("maxConcurrentBulkOperations <%s> invalid, defaulting to %d", maxBulkOpsStr, maxBulkOps)
		}
	}

	//
	// Setup Repositories and UseCases
	//
	userTable, err := userdb.NewTable(db)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.UnableToCreateRepositoryErrorCode,
			logging.ErrorDetail: "unable to create a userdb.Table instance",
		}).Fatal(mverr.UnableToCreateRepositoryMsg)
		os.Exit(1)
	}
	userSvc, err := services.NewUserSvc(userTable, logger, maxBulkOps)
	if err != nil {
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.UnableToCreateUserSvcErrorCode,
			logging.ErrorDetail: "unable to create a services.UserUseCase instance",
		}).Fatal(mverr.UnableToCreateUserSvcMsg)
		os.Exit(1)
	}

	//
	// Setup endpoints and start service
	//
	port, ok := configs["port"]
	if !ok {
		logger.Info("port configuration unavailable (configs[port]), defaulting to 5000")
		port = "5000"
	}
	port = ":" + port

	switch *protocolType {
	case "http":
		s, err := startHTTPServer(userSvc, logger, maxBulkOps, port)
		if err != nil {
			logger.WithFields(log.Fields{
				logging.ErrorCode:   mverr.UnableToCreateHTTPHandlerErrorCode,
				logging.ErrorDetail: err.Error(),
			}).Fatal(mverr.UnableToCreateHTTPHandlerMsg)
			os.Exit(1)
		}
		logger.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.SecretsDirName: *secretsDir,
			logging.Port:           port,
			logging.LogLevel:       log.GetLevel().String(),
			logging.DBHost:         configs["dbHost"],
			logging.DBPort:         configs["dbPort"],
			logging.DBName:         configs["dbName"],
		}).Info("accountd HTTP service running")

		handleTermSignalHTTP(s, logger, 10)

	case "grpc":
		s, err := startGRPCServer(userSvc, logger, maxBulkOps, port)
		if err != nil {
			logger.WithFields(log.Fields{
				logging.ErrorCode:   mverr.UnableToCreateRPCServerErrorCode,
				logging.ErrorDetail: err.Error(),
			}).Fatal(mverr.UnableToCreateRPCServerErrorMsg)
			os.Exit(1)
		}
		logger.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.SecretsDirName: *secretsDir,
			logging.Port:           port,
			logging.LogLevel:       log.GetLevel().String(),
			logging.DBHost:         configs["dbHost"],
			logging.DBPort:         configs["dbPort"],
			logging.DBName:         configs["dbName"],
		}).Info("accountd gRPC service running")

		handleTermSignalGRPC(s, logger)

	default:
		logger.WithFields(log.Fields{
			logging.ErrorCode:   mverr.InvalidProtocolTypeErrorCode,
			logging.ErrorDetail: fmt.Sprintf("invalid protocolType, %s, provided", *protocolType),
		}).Fatal(mverr.InvalidProtocolTypeErrorMsg)
		os.Exit(1)
	}
}

//
// Helper funcs
//

// handleTermSignalHTTP provides a mechanism to catch SIGTERMs and gracefully
// shutdown the HTTP service endpoint.
func handleTermSignalHTTP(s *http.Server, logger *log.Entry, timeout int) {
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

// handleTermSignalGRPC provides a mechanism to catch SIGTERMs and gracefully
// shutdown the gRPC service endpoint.
func handleTermSignalGRPC(s *grpc.Server, logger *log.Entry) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	logger.Info("Server shutting down")

	s.GracefulStop()
	logger.Info("Server stopped")
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

func startHTTPServer(userSvc *services.UserSvc, logger *log.Entry, maxBulkOps int, port string) (*http.Server, error) {
	usersHandler, err := users.NewUserHandler(userSvc, logger, maxBulkOps)
	if err != nil {
		return nil, err
	}

	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()
	mux.Handle("/users", usersHandler)  // Desired to prevent redirects. Can remove if redirects for '/users/' are OK
	mux.Handle("/users/", usersHandler) // Required to properly route requests to '/users/{id}. Don't understand why the above route isn't sufficient
	mux.Handle("/accountdhealth", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			logging.ErrorCode: mverr.MalformedURLErrorCode,
			logging.Path:      r.URL.Path,
		}).Error(mverr.MalformedURLMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mverr.MalformedURLMsg))
	})

	s := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			// TODO: improve logging (e.g., 'WithFields...')
			logger.Fatal(err)
		}
	}()

	return s, nil
}

func startGRPCServer(userSvc *services.UserSvc, logger *log.Entry, maxBulkOps int, port string) (*grpc.Server, error) {
	usersServer, err := grpcuser.NewUserServer(userSvc, logger)
	conn, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	pb.RegisterUserServerServer(s, usersServer)

	go func() {
		defer conn.Close()

		if err := s.Serve(conn); err != nil {
			// TODO: improve logging (e.g., 'WithFields...')
			logger.Fatal(err)
		}
	}()

	return s, nil
}
