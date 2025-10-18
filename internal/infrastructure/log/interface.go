package log

import "context"

type Logger interface {
	Info(ctx context.Context, message string, tags map[string]string)
	Error(ctx context.Context, message string, tags map[string]string)
	Warn(ctx context.Context, message string, tags map[string]string)
}
