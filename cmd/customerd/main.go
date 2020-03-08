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
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"
	"github.com/youngkin/mockvideo/internal/platform/config"
	"github.com/youngkin/mockvideo/internal/platform/constants"
	"github.com/youngkin/mockvideo/internal/platform/logging"

	log "github.com/sirupsen/logrus"
)

// TODO:
//	-2	TODO: Use error codes in errors (e.g., nil db pointer passed text message should also include the error code)
//	-1	TODO: Use config struct
//	0.	TODO: Rethink errors, see TODO in errors.go
//	1.	TODO: HTTP connection configs (e.g., timeouts)
//  1.  TODO: DB configs, use secrets for user name and password
//	6.	TODO: Kube logging
//	7.	TODO: ELK stack for logging
//	10.	TODO: Create build system that will compile and create docker image
//	11.	TODO: Use https
//	4.	TODO: Config parms (configMap?), monitor for changes restarting if necessary
//	5.	ONGOING: Prometheus
//	2.	DONE: Don't panic, add error handling
//	3.	DONE: Use logging levels
//	8.	DONE: Graphana for Prometheus
//	9.	DONE: Helm/Kube deployment (with P9S support)
//	9.	DONE: Travis CI

func main() {
	configFileName := flag.String("configFile",
		"/opt/mockvideo/custd/config/config",
		"specifies the location of the custd service configuration")

	secretsDir := flag.String("secretsDir",
		"/opt/mockvideo/custd/secrets",
		"specifies the location of the custd secrets")
	flag.Parse()

	logger := logging.GetLogger().WithField(constants.Application, constants.Customer)

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
	configs, err := config.LoadConfig(configFile)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.AppError:       constants.UnableToLoadConfig,
			constants.ErrorDetail:    err.Error(),
		}).Fatal("Error loading config data")
	}

	secrets, err := config.LoadSecrets(*secretsDir)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *secretsDir,
			constants.AppError:       constants.UnableToLoadSecrets,
			constants.ErrorDetail:    err.Error(),
		}).Fatal("Error loading secrets data")
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

	//
	// Setup DB connection
	//
	connStr, err := getDBConnectionStr(configs, secrets)
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
	customersHandler, err := handlers.New(db, logger)
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
	mux.HandleFunc("/sleeper", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			constants.ServiceName: "sleeper",
		}).Info("handling request")
		time.Sleep(10 * time.Second)
	})

	port, ok := configs["port"]
	if !ok {
		logger.Info("port configuration unavailable (configs[port]), defaulting to 5000")
		port = "5000"
	}
	port = ":" + port

	s := &http.Server{Addr: port, Handler: mux}

	go func() {
		logger.WithFields(log.Fields{
			constants.ConfigFileName: *configFileName,
			constants.SecretsDirName: *secretsDir,
			constants.Port:           port,
			constants.LogLevel:       log.GetLevel().String(),
			constants.DBHost:         configs["dbHost"],
			constants.DBPort:         configs["dbPort"],
			constants.DBName:         configs["dbName"],
		}).Info("customerd service starting")

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
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	<-term

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
	// E.g., "username:userpassword@tcp(10.0.0.100:3306)/mockvideo"
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

	return sb.String(), nil
}
