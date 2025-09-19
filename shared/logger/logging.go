package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Message   string                 `json:"message"`
	Hostname  string                 `json:"hostname"`
	RequestID string                 `json:"request_id,omitempty"`
	Error     map[string]interface{} `json:"error,omitempty"`
}

// Logger interface
type Logger interface {
	Info(service, action, message, requestID string)
	Error(service, action, message, requestID string, err error)
	Debug(service, action, message, requestID string)
}

type consoleLogger struct{}

func NewConsoleLogger() Logger {
	return &consoleLogger{}
}

func (l *consoleLogger) Info(service, action, message, requestID string) {
	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Service:   service,
		Action:    action,
		Message:   message,
		Hostname:  getHostname(),
		RequestID: requestID,
	}
	outputLog(logEntry)
}
func (l *consoleLogger) Debug(service, action, message, requestID string) {}

func (l *consoleLogger) Error(service, action, message, requestID string, err error) {
	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "ERROR",
		Service:   service,
		Action:    action,
		Message:   message,
		Hostname:  getHostname(),
		RequestID: requestID,
	}
	if err != nil {
		logEntry.Error = map[string]interface{}{
			"msg":   err.Error(),
			"stack": fmt.Sprintf("%+v", err),
		}
	}
	outputLog(logEntry)
}

func outputLog(entry LogEntry) {
	jsonData, _ := json.Marshal(entry)
	fmt.Println(string(jsonData))
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
