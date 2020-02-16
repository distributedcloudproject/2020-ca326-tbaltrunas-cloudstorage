package utils

import (
	// "fmt"
	"os"
	"io"
	// "time"
	"log"

	"github.com/hashicorp/logutils"
)

var (
	logger *log.Logger
	fileDir string = "logs"
	LogLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR", "CRITICAL"}
	defaultLogLevel = "INFO"
)

// NewLogger creates a new global logger using the given writer and level.
// Level should be one of the levels found in the LogLevels variable.
// The old logger variable is overriden.
func NewLogger(writer io.Writer, level string) *log.Logger {
	logger = log.New(writer, "", log.Ldate | log.Ltime | log.Lshortfile)
	filter := &logutils.LevelFilter{
		Levels: LogLevels,
		MinLevel: logutils.LogLevel(level),
		Writer: writer,
	}
	logger.SetOutput(filter)
	logger.Println("[DEBUG] Created a new logger.")
	return logger
}

// GetLogger returns a global logger variable, or creates a new default logger.
func GetLogger() *log.Logger {
	if logger == nil {
		writer := os.Stderr
		logger = log.New(writer, "", log.Ldate | log.Ltime | log.Lshortfile)
		filter := &logutils.LevelFilter{
			Levels: LogLevels,
			MinLevel: logutils.LogLevel(defaultLogLevel),
			Writer: writer,
		}
		logger.SetOutput(filter)
		logger.Println("[DEBUG] Created default logger.")
	}
	return logger
}
