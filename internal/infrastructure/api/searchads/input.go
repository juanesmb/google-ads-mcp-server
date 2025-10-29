package searchads

// Filters captures the parameters used to search for ads.
type Filters struct {
	CustomerID      string
	CampaignIDs     []string
	CampaignNames   []string
	AdGroupIDs      []string
	AdGroupNames    []string
	Statuses        []string
	AdTypes         []string
	DateRangeStart  string
	DateRangeEnd    string
}

