package searchads

// Ad represents a normalized Google Ads ad.
type Ad struct {
	ID                  string
	ResourceName        string
	Name                string
	Type                string
	Status              string
	FinalURLs           []string
	ApprovalStatus      string
	CampaignID          string
	CampaignName        string
	CampaignResourceName string
	AdGroupID           string
	AdGroupName         string
	AdGroupResourceName  string
	ExpandedTextAd      *ExpandedTextAd
	ResponsiveSearchAd  *ResponsiveSearchAd
	CallOnlyAd          *CallOnlyAd
	Metrics             AdMetrics
}

// ExpandedTextAd represents fields specific to expanded text ads.
type ExpandedTextAd struct {
	HeadlinePart1 string
	HeadlinePart2 string
	HeadlinePart3 string
	Description   string
	Description2  string
	Path1          string
	Path2          string
}

// ResponsiveSearchAd represents fields specific to responsive search ads.
type ResponsiveSearchAd struct {
	Headlines    []string
	Descriptions []string
	Path1         string
	Path2         string
}

// CallOnlyAd represents fields specific to call-only ads.
type CallOnlyAd struct {
	Headline1             string
	Headline2             string
	Description1          string
	Description2          string
	PhoneNumber           string
	CallTracked           bool
	DisableCallConversion bool
}

// AdMetrics represents comprehensive metrics for an ad.
type AdMetrics struct {
	Clicks                             int64
	Impressions                        int64
	CTR                                float64
	AverageCPC                         int64 // in micros
	CostMicros                         int64
	Conversions                        float64
	ConversionsValue                   float64
	CostPerConversion                  float64
	ConversionRate                     float64
	AllConversions                     float64
	AllConversionsValue                float64
	AllConversionsFromInteractionsRate float64
	CostPerAllConversions              float64
	Interactions                       int64
	EngagementRate                     float64
	SearchImpressionShare              float64
	SearchRankLostImpressionShare      float64
	VideoViews                         int64
	VideoViewRate                      float64
	AverageCPV                         int64 // in micros
}

// Result aggregates the ads outcome along with pagination metadata.
type Result struct {
	Ads             []Ad
	NextPageToken   string
	TotalResultsCount int64
}

