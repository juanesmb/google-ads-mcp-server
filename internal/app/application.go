package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google-ads-mcp/internal/app/configs"
	"google-ads-mcp/internal/infrastructure/middleware"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Start() {
	cfgs := configs.ReadConfigs()

	server := initServer(cfgs)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	mux := http.NewServeMux()
	mux.Handle(cfgs.ServerConfig.Path, handler)

	wrappedHandler := middleware.LoggingHandler(mux)

	httpServer := &http.Server{
		Addr:    cfgs.ServerConfig.BindAddress,
		Handler: wrappedHandler,
	}

	log.Printf("Google Ads MCP server (streamable HTTP) listening on path %s (bind %s)", cfgs.ServerConfig.Path, cfgs.ServerConfig.BindAddress)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			serverErrCh <- err
		}
	}()

	select {
	case <-shutdownCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
