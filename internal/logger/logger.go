package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	debugMode   bool
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	debugLogger = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(os.Stderr, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func SetDebugMode(enabled bool) {
	debugMode = enabled
	if debugMode {
		Debug("Debug mode enabled")
	}
}

func IsDebugMode() bool {
	return debugMode
}

func Debug(format string, args ...interface{}) {
	if debugMode {
		_ = debugLogger.Output(2, fmt.Sprintf(format, args...))
	}
}

func Info(format string, args ...interface{}) {
	infoLogger.Printf(format, args...)
}

func Error(format string, args ...interface{}) {
	_ = errorLogger.Output(2, fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", fmt.Sprintf(format, args...))
	}
}

// Request logging function for HTTP requests
func LogRequest(method, path, remoteAddr string) {
	if debugMode {
		Debug("HTTP %s %s from %s", method, path, remoteAddr)
	}
}

// Response logging function for HTTP responses
func LogResponse(method, path string, statusCode int, duration string) {
	if debugMode {
		Debug("HTTP %s %s -> %d (%s)", method, path, statusCode, duration)
	}
}
