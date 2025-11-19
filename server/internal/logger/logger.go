package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(lvl string) *slog.Logger {
	var lvlParsed slog.Level
	if err := lvlParsed.UnmarshalText([]byte(lvl)); err != nil {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		logger.Warn("unknown log level, setting INFO as the default", slog.String("provided", lvl))
		lvlParsed = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvlParsed}))
}
