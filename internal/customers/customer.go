package customers

// TODO
//	1.	TODO: What to do about tests? cmd/customerd/handlers/customers_test.go already handles all the required
//		TODO: tests, and it **NEEDS** to. Adding tests here would be a duplicate.
import (
	"database/sql"

	"github.com/juju/errors"
)

// Customer represents the data about a customer
type Customer struct {
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	StreetAddress string `json:"StreetAddress,omitempty"`
	City          string `json:"City,omitempty"`
	State         string `json:"State,omitempty"`
	Country       string `json:"Country,omitempty"`
}

// Customers is a collection (slice) of Customer
type Customers struct {
	Customers []Customer `json:"Customers"`
}

// GetAllCustomers will return all customers known to the application
func GetAllCustomers(db *sql.DB) (Customers, error) {
	results, err := db.Query("SELECT id, name, streetAddress, city, state, country FROM customer")
	if err != nil {
		return Customers{}, errors.Annotate(err, "error querying DB")
	}

	custs := Customers{}
	for results.Next() {
		var customer Customer

		err = results.Scan(&customer.ID,
			&customer.Name,
			&customer.StreetAddress,
			&customer.City,
			&customer.State,
			&customer.Country)
		if err != nil {
			return Customers{}, errors.Annotate(err, "error scanning result set")
		}

		customer = Customer{
			ID:            customer.ID,
			Name:          customer.Name,
			StreetAddress: customer.StreetAddress,
			City:          customer.City,
			State:         customer.State,
			Country:       customer.Country,
		}
		custs.Customers = append(custs.Customers, customer)
	}

	return custs, nil
}
