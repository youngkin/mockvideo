package customers

// TODO
//	1.	TODO: What to do about tests? cmd/customerd/handlers/customers_test.go already handles all the required
//		TODO: tests, and it **NEEDS** to. Adding tests here would be a duplicate. But some of what's done here is
//		arguably "whitebox" testing (e.g., Query error path testing such as 'sql.ErrNoRows' in 'GetCustomer()').
//		Also, if testing this capability is left to clients, who knows if they're actually going to test?
import (
	"database/sql"
	"fmt"

	"github.com/juju/errors"
)

// Customer represents the data about a customer
type Customer struct {
	HREF          string `json:"href"`
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	StreetAddress string `json:"StreetAddress,omitempty"`
	City          string `json:"City,omitempty"`
	State         string `json:"State,omitempty"`
	Country       string `json:"Country,omitempty"`
}

// Customers is a collection (slice) of Customer
type Customers struct {
	Customers []*Customer `json:"Customers"`
}

// GetAllCustomers will return all customers known to the application
func GetAllCustomers(db *sql.DB) (*Customers, error) {
	results, err := db.Query("SELECT id, name, streetAddress, city, state, country FROM customer")
	if err != nil {
		return &Customers{}, errors.Annotate(err, "error querying DB")
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
			return &Customers{}, errors.Annotate(err, "error scanning result set")
		}

		custs.Customers = append(custs.Customers, &customer)
	}

	return &custs, nil
}

// GetCustomer will return the customer identified by 'id' or a nil customer if there
// wasn't a matching customer.
func GetCustomer(db *sql.DB, id int) (*Customer, error) {
	q := fmt.Sprintf("SELECT id, name, streetAddress, city, state, country FROM customer WHERE id=%d", id)
	row := db.QueryRow(q)
	cust := &Customer{}
	err := row.Scan(&cust.ID,
		&cust.Name,
		&cust.StreetAddress,
		&cust.City,
		&cust.State,
		&cust.Country)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Annotate(err, "error scanning customer row")
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return cust, nil
}
