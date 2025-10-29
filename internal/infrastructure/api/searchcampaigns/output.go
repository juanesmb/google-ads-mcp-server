package searchcampaigns

// Campaign represents a normalized Google Ads campaign.
type Campaign struct {
	ID                     string
	ResourceName           string
	Name                   string
	Status                 string
	AdvertisingChannelType string
	BiddingStrategyType    string
	BudgetAmountMicros     int64
	OptimizationScore      float64
	Metrics                CampaignMetrics
}

// CampaignMetrics represents metrics for a campaign.
type CampaignMetrics struct {
	Clicks                             int64   // Total clicks
	Impressions                        int64   // Total impressions
	CTR                                float64 // Click-through rate
	AverageCPC                         int64   // Average cost per click in micros
	CostMicros                         int64   // Total cost in micros
	Conversions                        float64 // Conversions
	ConversionsValue                   float64 // Total conversion value
	CostPerConversion                  float64 // Cost per conversion in currency units
	ConversionRate                     float64 // Conversion rate (percentage)
	AllConversions                     float64 // All conversions (including estimated)
	AllConversionsValue                float64 // Total value of all conversions
	AllConversionsFromInteractionsRate float64 // All conversions rate from interactions
	AllConversionsValuePerCost         float64 // All conversions value per cost
	CostPerAllConversions              float64 // Cost per all conversions in currency units
	Interactions                       int64   // Total interactions (clicks + engagements)
	EngagementRate                     float64 // Engagement rate (percentage)
	SearchImpressionShare              float64 // Search impression share (percentage)
	SearchRankLostImpressionShare      float64 // Search rank lost impression share (percentage)
}

// Result aggregates the campaigns outcome along with pagination metadata.
type Result struct {
	Campaigns         []Campaign
	NextPageToken     string
	TotalResultsCount int64
}
