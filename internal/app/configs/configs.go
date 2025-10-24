package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Configs struct {
	ServerConfig    ServerConfig
	GoogleAdsConfig GoogleAdsConfig
}

type ServerConfig struct {
	BindAddress string
	Port        string
	Path        string
}

type GoogleAdsConfig struct {
	CustomerID         string
	DeveloperToken     string
	ServiceAccountJSON string
}

// GoogleAdsConfigData represents the unified configuration structure
type GoogleAdsConfigData struct {
	CustomerID         string `json:"customer_id"`
	DeveloperToken     string `json:"developer_token"`
	ServiceAccountJSON string `json:"service_account_json"`
}

func ReadConfigs() Configs {
	// Non-sensitive server configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("MCP_SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	path := os.Getenv("MCP_SERVER_PATH")
	if path == "" {
		path = "/mcp"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	bindAddress := fmt.Sprintf("%s:%s", host, port)

	// Read unified Google Ads configuration - try local file first, then Google Secret Manager
	googleAdsConfig, err := readGoogleAdsConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to read Google Ads configuration: %v", err))
	}

	return Configs{
		ServerConfig: ServerConfig{
			BindAddress: bindAddress,
			Port:        port,
			Path:        path,
		},
		GoogleAdsConfig: googleAdsConfig,
	}
}

// readGoogleAdsConfig reads unified Google Ads configuration from local file or environment variable
func readGoogleAdsConfig() (GoogleAdsConfig, error) {
	// First, try to read from local config file (for local development)
	localConfigPath := "internal/app/configs/google-ads-config.json"
	if _, err := os.Stat(localConfigPath); err == nil {
		data, err := os.ReadFile(localConfigPath)
		if err != nil {
			return GoogleAdsConfig{}, fmt.Errorf("failed to read local config file: %w", err)
		}

		var configData GoogleAdsConfigData
		if err := json.Unmarshal(data, &configData); err != nil {
			return GoogleAdsConfig{}, fmt.Errorf("failed to parse local config JSON: %w", err)
		}

		return configData.ToGoogleAdsConfig(), nil
	}

	// If no local config file found, try environment variable (for production)
	// Google Cloud Run automatically populates GOOGLE_ADS_CONFIG with secret value
	return readGoogleAdsConfigFromEnv()
}

// readGoogleAdsConfigFromEnv reads unified Google Ads configuration from environment variable
func readGoogleAdsConfigFromEnv() (GoogleAdsConfig, error) {
	configJSON := os.Getenv("GOOGLE_ADS_CONFIG")
	if configJSON == "" {
		return GoogleAdsConfig{}, fmt.Errorf("GOOGLE_ADS_CONFIG environment variable is required for production")
	}

	// Parse the JSON configuration
	var configData GoogleAdsConfigData
	if err := json.Unmarshal([]byte(configJSON), &configData); err != nil {
		return GoogleAdsConfig{}, fmt.Errorf("failed to parse GOOGLE_ADS_CONFIG JSON: %w", err)
	}

	return configData.ToGoogleAdsConfig(), nil
}

// ToGoogleAdsConfig converts GoogleAdsConfigData to GoogleAdsConfig
func (d GoogleAdsConfigData) ToGoogleAdsConfig() GoogleAdsConfig {
	return GoogleAdsConfig(d)
}
