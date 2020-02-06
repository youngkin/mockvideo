package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/youngkin/mockvideo/cmd/customerd/handlers"
	"github.com/youngkin/mockvideo/internal/platform/config"
	"github.com/youngkin/mockvideo/internal/platform/errors"
	"github.com/youngkin/mockvideo/internal/platform/logging"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

// TODO:
//	1.	TODO: HTTP connection config (e.g., timeouts)
//	2.	TODO: Don't panic, add error handling
//	3.	TODO: Use logging levels
//	4.	TODO: Config parms (configMap?), monitor for changes restarting if necessary
//	5.	TODO: Prometheus
//	6.	TODO: Kube logging
//	7.	TODO: ELK stack for logging
//	8.	TODO: Graphana for Prometheus
//	9.	TODO: Helm/Kube deployment (with P9S support)
//	10.	TODO: Create build system that will compile and create docker image
//	11.	TODO: Use https
//	9.	DONE: Travis CI

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the INFO severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	log.WithFields(log.Fields{
		logging.Application: logging.Customer,
	})
	// TODO: REMOVE port := flag.Int("port", 5999, "the port to start the customer service on")
	configFileName := flag.String("configFile",
		"/opt/mockvideo/custd",
		"specifies the location of the custd service configuration")
	flag.Parse()

	configFile, err := os.Open(*configFileName)
	if err != nil {
		log.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.Error:          err.Error(),
		}).Fatal("Error opening config file")
		os.Exit(errors.BadConfigFileExit)
	}
	config, err := config.LoadConfig(configFile)
	if err != nil {
		log.WithFields(log.Fields{
			logging.ConfigFileName: *configFileName,
			logging.Error:          err.Error(),
		}).Fatal("Error loading config file")
		os.Exit(errors.UnableToGetConfigExit)
	}

	loglevel, ok := config["logLevel"]
	if !ok {
		log.Warn("Log level unavailable, defaulting to INFO")
	}
	if ok {
		level, err := strconv.Atoi(loglevel)
		if err != nil {
			log.Warnf("Log level <%s> invalid, defaulting to INFO", loglevel)
		} else {
			log.SetLevel(log.Level(level))
		}
	}

	// TODO: Get from Kubernetes secret
	db, err := sql.Open("mysql", "admin:2girls1cat@tcp(10.0.0.100:3306)/mockvideo")

	if err != nil {
		log.WithFields(log.Fields{
			logging.Error: err.Error(),
		}).Fatal("Error opening DB connection")
		os.Exit(errors.BadDBConnectionExit)
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	customersHandler, err := handlers.New(db)
	if err != nil {
		//TODO Handle? Probably just to cover the case of a coding error where db == nil
	}
	healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()

	mux.Handle("/customers", customersHandler)
	mux.Handle("/custdhealth", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	port, ok := config["port"]
	if !ok {
		log.Printf("port configuration unavailable (config[\"port\"])\n")
		os.Exit(errors.UnableToGetPortConfigExit)
	}

	log.WithFields(log.Fields{
		logging.Port:           port,
		logging.ConfigFileName: configFileName,
		logging.LogLevel:       "TODO, where to get this from", //TODO
	}).Info("customerd service starting")
	log.Fatal(http.ListenAndServe(port, mux))
}
