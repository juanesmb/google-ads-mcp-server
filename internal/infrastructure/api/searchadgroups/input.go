package searchadgroups

// Filters captures the parameters used to search for ad groups.
type Filters struct {
	CustomerID      string
	AdGroupIDs      []string
	AdGroupNames    []string
	Statuses        []string
	CampaignIDs     []string
	CampaignNames   []string
	DateRangeStart  string
	DateRangeEnd    string
}

