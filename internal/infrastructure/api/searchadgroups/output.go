package searchadgroups

// AdGroup represents a normalized Google Ads ad group.
type AdGroup struct {
	ID                     string
	ResourceName           string
	Name                   string
	Status                 string
	Type                   string
	CampaignID             string
	CampaignName           string
	CampaignResourceName   string
	Metrics                AdGroupMetrics
}

// AdGroupMetrics represents metrics for an ad group.
type AdGroupMetrics struct {
	Clicks      int64   // Total clicks
	Impressions int64   // Total impressions
	CTR         float64 // Click-through rate
	AverageCPC  int64   // Average cost per click in micros
	CostMicros  int64   // Total cost in micros
}

// Result aggregates the ad groups outcome along with pagination metadata.
type Result struct {
	AdGroups          []AdGroup
	NextPageToken     string
	TotalResultsCount int64
}

