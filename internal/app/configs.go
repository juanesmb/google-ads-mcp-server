package app

import (
	"fmt"
	"os"
	"strings"
)

type Configs struct {
	ServerConfig ServerConfig
}

type ServerConfig struct {
	BindAddress string
	Port        string
	Path        string
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
	}
}
