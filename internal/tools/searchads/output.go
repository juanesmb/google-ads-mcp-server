package searchads

// ToolOutput captures the structured response returned to the MCP client.
type ToolOutput struct {
	Ads           []AdOutput `json:"ads"`
	NextPageToken string     `json:"next_page_token,omitempty"`
	TotalCount    int64      `json:"total_count"`
}

// AdOutput mirrors the normalized ad representation returned to clients.
type AdOutput struct {
	ID                   string          `json:"id"`
	ResourceName         string          `json:"resource_name"`
	Name                 string          `json:"name,omitempty"`
	Type                 string          `json:"type"`
	Status               string          `json:"status"`
	FinalURLs            []string        `json:"final_urls,omitempty"`
	ApprovalStatus       string          `json:"approval_status,omitempty"`
	CampaignID           string          `json:"campaign_id"`
	CampaignName         string          `json:"campaign_name"`
	CampaignResourceName string          `json:"campaign_resource_name"`
	AdGroupID            string          `json:"ad_group_id"`
	AdGroupName          string          `json:"ad_group_name"`
	AdGroupResourceName  string          `json:"ad_group_resource_name"`
	ExpandedTextAd       *ExpandedTextAd `json:"expanded_text_ad,omitempty"`
	ResponsiveSearchAd   *ResponsiveSearchAd `json:"responsive_search_ad,omitempty"`
	CallOnlyAd           *CallOnlyAd     `json:"call_only_ad,omitempty"`
	Metrics              AdMetrics       `json:"metrics"`
}

// ExpandedTextAd represents fields specific to expanded text ads.
type ExpandedTextAd struct {
	HeadlinePart1 string   `json:"headline_part1,omitempty"`
	HeadlinePart2 string   `json:"headline_part2,omitempty"`
	HeadlinePart3 string   `json:"headline_part3,omitempty"`
	Description   string   `json:"description,omitempty"`
	Description2  string   `json:"description2,omitempty"`
	Path1          string   `json:"path1,omitempty"`
	Path2          string   `json:"path2,omitempty"`
}

// ResponsiveSearchAd represents fields specific to responsive search ads.
type ResponsiveSearchAd struct {
	Headlines          []string `json:"headlines,omitempty"`
	Descriptions       []string `json:"descriptions,omitempty"`
	Path1              string   `json:"path1,omitempty"`
	Path2              string   `json:"path2,omitempty"`
}

// CallOnlyAd represents fields specific to call-only ads.
type CallOnlyAd struct {
	Headline1           string `json:"headline1,omitempty"`
	Headline2           string `json:"headline2,omitempty"`
	Description1        string `json:"description1,omitempty"`
	Description2        string `json:"description2,omitempty"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	CallTracked         bool   `json:"call_tracked,omitempty"`
	DisableCallConversion bool `json:"disable_call_conversion,omitempty"`
}

// AdMetrics represents comprehensive metrics for an ad.
type AdMetrics struct {
	Clicks                             int64   `json:"clicks"`
	Impressions                        int64   `json:"impressions"`
	CTR                                float64 `json:"ctr"`
	AverageCPC                         int64   `json:"average_cpc_micros"` // in micros
	CostMicros                         int64   `json:"cost_micros"`
	Conversions                        float64 `json:"conversions"`
	ConversionsValue                   float64 `json:"conversions_value"`
	CostPerConversion                  float64 `json:"cost_per_conversion"`
	ConversionRate                     float64 `json:"conversion_rate"`
	AllConversions                     float64 `json:"all_conversions"`
	AllConversionsValue                float64 `json:"all_conversions_value"`
	AllConversionsFromInteractionsRate float64 `json:"all_conversions_from_interactions_rate"`
	CostPerAllConversions              float64 `json:"cost_per_all_conversions"`
	Interactions                       int64   `json:"interactions"`
	EngagementRate                     float64 `json:"engagement_rate"`
	SearchImpressionShare              float64 `json:"search_impression_share"`
	SearchRankLostImpressionShare      float64 `json:"search_rank_lost_impression_share"`
	VideoViews                         int64   `json:"video_views,omitempty"`
	VideoViewRate                      float64 `json:"video_view_rate,omitempty"`
	AverageCPV                         int64   `json:"average_cpv_micros,omitempty"` // in micros
}

