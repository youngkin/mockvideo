package customers

// Customer represents the data about a customer
type Customer struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	StreetAddress string `json:"streetAddress",omitempty`
	City          string `json:"city",omitempty`
	State         string `json:"state",omitempty`
	Country       string `json:"country",omitempty`
}
