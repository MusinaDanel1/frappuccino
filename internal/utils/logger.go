package utils

import (
	"fmt"
	"log/slog"
	"os"
)

/*
This Go code sets up a logger that writes log entries to a specified file. It uses Go's built-in log/slog package for structured logging. Here's a breakdown:

SetupLogger(logFilePath string): This function creates and configures a logger that writes to the specified log file.
*/

func SetupLogger(logFilePath string) (*slog.Logger, error) {
	if _, err := os.Stat(logFilePath); os.IsExist(err) {
		os.RemoveAll(logFilePath)
	}
	_, err := os.Create(logFilePath)
	if err != nil {
		fmt.Printf("Failed fo create a file %s: %v\n", logFilePath, err)
	}
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	fileHandler := slog.NewJSONHandler(file, nil)

	logger := slog.New(fileHandler)

	return logger, nil
}
