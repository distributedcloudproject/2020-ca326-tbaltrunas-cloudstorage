package utils

import (
	"os"
	"io"
	"log"

	"github.com/hashicorp/logutils"
)

var (
	logger *log.Logger
	fileDir string = "logs"
	LogLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR", "CRITICAL"}
    defaultWriter = os.Stderr
	defaultLevel = "WARN"
	prefix = ""
	flags = log.Ldate | log.Ltime | log.Lshortfile
)

func init() {
	logger = NewLoggerFromEnv()
}

// GetLogger returns a global logger variable, or creates a new default logger.
func GetLogger() *log.Logger {
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

// NewLoggerFromWriterLevel creates a new global logger using the given writer and level.
// Level should be one of the levels found in the LogLevels variable.
// The old logger variable is overriden.
func NewLoggerFromWriterLevel(writer io.Writer, level string) *log.Logger {
	logger = log.New(writer, prefix, flags)
	filter := &logutils.LevelFilter{
		Levels: LogLevels,
		MinLevel: logutils.LogLevel(level),
		Writer: writer,
	}
	logger.SetOutput(filter)
	logger.Println("[DEBUG] Created a new logger from writer and level.")
	return logger
}

// NewLoggerFromWriter creates a new logger from a writer, using the default level.
func NewLoggerFromWriter(writer io.Writer) *log.Logger {
	logger = log.New(writer, prefix, flags)
	filter := &logutils.LevelFilter{
		Levels: LogLevels,
		MinLevel: logutils.LogLevel(defaultLevel),
		Writer: writer,
	}
	logger.SetOutput(filter)
	logger.Println("[DEBUG] Created a new logger from writer.")
	return logger
}

// NewLoggerFromLevel creates a new logger using a certain log level, using the default writer.
func NewLoggerFromLevel(level string) *log.Logger {
	logger = log.New(defaultWriter, "", log.Ldate | log.Ltime | log.Lshortfile)
	filter := &logutils.LevelFilter{
		Levels: LogLevels,
		MinLevel: logutils.LogLevel(level),
		Writer: defaultWriter,
	}
	logger.SetOutput(filter)
	logger.Println("[DEBUG] Created a new logger from level.")
	return logger
}

// NewLoggerFromEnv creates a new logger from environment variables.
func NewLoggerFromEnv() *log.Logger {
	lvl, okLvl := os.LookupEnv("CLOUD_LOG_LEVEL")
	logFile, okFile := os.LookupEnv("CLOUD_LOG_FILE")
	var logLevel string
	var logWriter io.Writer

	if okLvl {
		logLevel = lvl
	} else {
		logLevel = defaultLevel
	}

	if okFile {
		f, err := os.Create(logFile)
		if err != nil {
			GetLogger().Printf("[ERROR] %v.", err)
			return GetLogger()
		}
		logWriter = f
	} else {
		logWriter = defaultWriter
	}

	return NewLoggerFromWriterLevel(logWriter, logLevel)
}

func newDefaultLogger() *log.Logger {
	logger = log.New(defaultWriter, prefix, flags)
	filter := &logutils.LevelFilter{
		Levels: LogLevels,
		MinLevel: logutils.LogLevel(defaultLevel),
		Writer: defaultWriter,
	}
	logger.SetOutput(filter)
	logger.Println("[DEBUG] Created a new default logger.")
	return logger
}
