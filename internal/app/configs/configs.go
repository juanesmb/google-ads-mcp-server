package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
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

// readGoogleAdsConfig reads unified Google Ads configuration from local file or Google Secret Manager
func readGoogleAdsConfig() (GoogleAdsConfig, error) {
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

		return GoogleAdsConfig(configData), nil
	}

	// If no local config file found, try Google Secret Manager (for production)
	return readGoogleAdsConfigFromSecretManager()
}

// readGoogleAdsConfigFromSecretManager reads unified Google Ads configuration from Google Secret Manager
func readGoogleAdsConfigFromSecretManager() (GoogleAdsConfig, error) {
	// Get the project ID from environment
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return GoogleAdsConfig{}, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is required for production")
	}

	// Create the secret manager client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return GoogleAdsConfig{}, fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer client.Close()

	// Build the secret name
	secretPath := fmt.Sprintf("projects/%s/secrets/GOOGLE_ADS_CONFIG/versions/latest", projectID)

	// Access the secret
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return GoogleAdsConfig{}, fmt.Errorf("failed to access GOOGLE_ADS_CONFIG secret: %w", err)
	}

	// Parse the JSON configuration
	var configData GoogleAdsConfigData
	if err := json.Unmarshal(result.Payload.Data, &configData); err != nil {
		return GoogleAdsConfig{}, fmt.Errorf("failed to parse Google Ads config JSON: %w", err)
	}

	return GoogleAdsConfig(configData), nil
}
