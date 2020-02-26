package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/mockvideo/internal/customers"
	"github.com/youngkin/mockvideo/internal/platform/logging"
)

// logger is used to control code-under-test logging behavior
var logger *log.Entry

func init() {
	logger = logging.GetLogger()
	// Uncomment for more verbose logging
	// logger.Logger.SetLevel(log.DebugLevel)
	// Supress all application logging
	logger.Logger.SetLevel(log.PanicLevel)
	// Uncomment for non-tty logging
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	//  })
}

func TestGetCustomers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	custHandler, err := New(db, logger)
	if err != nil {
		t.Fatalf("error '%s' was not expected when getting a customer handler", err)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "streetAddress", "city", "state", "country"}).
		AddRow(1, "porgy tirebiter", "123 anyStreet", "anyCity", "anyState", "anyCountry").
		AddRow(2, "mickey dolenz", "123 Laurel Canyon", "LA", "CA", "USA")

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	testSrv := httptest.NewServer(http.HandlerFunc(custHandler.ServeHTTP))
	defer testSrv.Close()

	url := testSrv.URL
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("an error '%s' was not expected calling customerd server", err)
	}
	defer resp.Body.Close()

	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("an error '%s' was not expected reading response body", err)
	}

	expected := customers.Customers{
		Customers: []customers.Customer{
			{
				ID:            1,
				Name:          "porgy tirebiter",
				StreetAddress: "123 anyStreet",
				City:          "anyCity",
				State:         "anyState",
				Country:       "anyCountry",
			},
			{
				ID:            2,
				Name:          "mickey dolenz",
				StreetAddress: "123 Laurel Canyon",
				City:          "LA",
				State:         "CA",
				Country:       "USA",
			},
		},
	}

	mExpected, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("an error '%s' was not expected Marshaling %+v", err, expected)
	}

	if bytes.Compare(mExpected, actual) != 0 {
		t.Errorf("expected %+v, got %+v", mExpected, actual)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCustomersError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	custHandler, err := New(db, logger)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting a customer handler", err)
	}

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnError(fmt.Errorf("some error"))

	testSrv := httptest.NewServer(http.HandlerFunc(custHandler.ServeHTTP))
	defer testSrv.Close()

	url := testSrv.URL
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("an error '%s' was not expected calling customerd server", err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != 500 {
		t.Errorf("expected StatusCode = 500, got %d", status)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCustomersRowScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	custHandler, err := New(db, logger)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting a customer handler", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT id, name, streetAddress, city, state, country FROM customer").
		WillReturnRows(rows)

	testSrv := httptest.NewServer(http.HandlerFunc(custHandler.ServeHTTP))
	defer testSrv.Close()

	url := testSrv.URL
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("an error '%s' was not expected calling customerd server", err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != 500 {
		t.Errorf("expected StatusCode = 500, got %d", status)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
