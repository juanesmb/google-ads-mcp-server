package searchcampaigns

// Campaign represents a normalized Google Ads campaign.
type Campaign struct {
	ID                      string
	ResourceName            string
	Name                    string
	Status                  string
	AdvertisingChannelType  string
	BiddingStrategyType     string
	BudgetAmountMicros      int64
	OptimizationScore       float64
	Metrics                 CampaignMetrics
}

// CampaignMetrics represents metrics for a campaign.
type CampaignMetrics struct {
	Clicks       int64
	Impressions  int64
	CTR          float64
	AverageCPC   int64 // in micros
	CostMicros   int64
}

// Result aggregates the campaigns outcome along with pagination metadata.
type Result struct {
	Campaigns          []Campaign
	NextPageToken      string
	TotalResultsCount  int64
}

