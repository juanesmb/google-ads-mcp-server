package listadaccounts

// Account represents a normalized Google Ads customer account.
type Account struct {
	CustomerID   string
	CustomerName string
	CurrencyCode string
	TimeZone     string
	Status       string
	ResourceName string
}

// Result aggregates the accounts outcome along with pagination metadata.
type Result struct {
	Accounts          []Account
	NextPageToken     string
	TotalResultsCount int64
}
