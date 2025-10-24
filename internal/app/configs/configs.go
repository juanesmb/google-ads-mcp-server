package configs

import (
	"context"
	"fmt"
	"io/ioutil"
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

func ReadConfigs() Configs {
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

	// Read service account JSON - try local file first, then Google Secret Manager
	serviceAccountJSON, err := readServiceAccountJSON()
	if err != nil {
		panic(fmt.Sprintf("failed to read service account JSON: %v", err))
	}

	return Configs{
		ServerConfig: ServerConfig{
			BindAddress: bindAddress,
			Port:        port,
			Path:        path,
		},
		GoogleAdsConfig: GoogleAdsConfig{
			CustomerID:         os.Getenv("GOOGLE_ADS_CUSTOMER_ID"),
			DeveloperToken:     os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN"),
			ServiceAccountJSON: serviceAccountJSON,
		},
	}
}

// readServiceAccountJSON reads service account JSON from local file or Google Secret Manager
func readServiceAccountJSON() (string, error) {
	// First, try to read from local file (for local development)
	// Look for any JSON file in the configs directory
	configDir := "internal/app/configs"
	files, err := ioutil.ReadDir(configDir)
	if err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				localFilePath := fmt.Sprintf("%s/%s", configDir, file.Name())
				data, err := ioutil.ReadFile(localFilePath)
				if err != nil {
					continue // Try next file
				}
				return string(data), nil
			}
		}
	}

	// If no local JSON file found, try Google Secret Manager (for production)
	return readSecretFromManager("GOOGLE_ADS_SERVICE_ACCOUNT_JSON")
}

// readSecretFromManager reads a secret from Google Secret Manager
func readSecretFromManager(secretName string) (string, error) {
	// Get the project ID from environment or use default
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "understory-mcps" // Default project ID
	}

	// Create the secret manager client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer client.Close()

	// Build the secret name
	secretPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName)

	// Access the secret
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}
