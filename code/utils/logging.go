package utils

import (
	"fmt"
	"os"
	"time"
	"log"
)

var (
	logger *log.Logger
	fileDir string = "logs"
)

func GetLogger() *log.Logger {
	if logger == nil {
		err := os.MkdirAll(fileDir, os.ModeDir)
		t := time.Now()
		logFile := fmt.Sprintf("%v/%v.log", fileDir, t.Format(time.RFC1123Z))
		f, err := os.Create(logFile)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		logger = log.New(f, "[LOG] ", log.Ldate | log.Ltime | log.Lshortfile)
	}
	return logger
}
