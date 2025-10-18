package local

import (
	"context"
	"log/slog"
	"os"
)

type LogService struct {
	client *slog.Logger
}

func NewLogger() *LogService {
	return &LogService{
		client: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l LogService) Info(ctx context.Context, message string, tags map[string]string) {
	l.client.Info(message, l.formatTags(tags)...)
}

func (l LogService) Error(ctx context.Context, message string, tags map[string]string) {
	l.client.Error(message, l.formatTags(tags)...)
}

func (l LogService) Warn(ctx context.Context, message string, tags map[string]string) {
	l.client.Warn(message, l.formatTags(tags)...)
}

func (l LogService) formatTags(tags map[string]string) []any {
	var logTags []any
	for key, value := range tags {
		logTags = append(logTags, key, value)
	}

	return logTags
}
