package searchcampaigns

// Filters captures the parameters used to search for campaigns.
type Filters struct {
	CustomerID      string
	CampaignIDs     []string
	CampaignNames   []string
	Statuses        []string
	DateRangeStart  string
	DateRangeEnd    string
}

