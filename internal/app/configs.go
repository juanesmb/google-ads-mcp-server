package app

import (
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
	CustomerID     string
	DeveloperToken string
	ClientID       string
	PrivateKeyID   string
	PrivateKey     string
}

func readConfigs() Configs {
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

	return Configs{
		ServerConfig: ServerConfig{
			BindAddress: bindAddress,
			Port:        port,
			Path:        path,
		},
		GoogleAdsConfig: GoogleAdsConfig{
			CustomerID:     os.Getenv("GOOGLE_ADS_CUSTOMER_ID"),
			DeveloperToken: os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN"),
			ClientID:       os.Getenv("GOOGLE_ADS_CLIENT_ID"),
			PrivateKeyID:   os.Getenv("GOOGLE_ADS_PRIVATE_KEY_ID"),
			PrivateKey:     os.Getenv("GOOGLE_ADS_PRIVATE_KEY"),
		},
	}
}
