package customers

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
