package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// InitLogger initializes the global logger based on the environment
func InitLogger(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// AddSource: true, // Uncomment if you want file:line in logs
	}

	if env == "production" {
		opts.Level = slog.LevelInfo
		// JSON format is standard for production log aggregators (ELK, Datadog)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text format is friendlier for local development
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}