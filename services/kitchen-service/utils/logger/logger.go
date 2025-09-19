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

type Logger struct {
	serviceName string
}

func New(serviceName string) *Logger {
	return &Logger{serviceName: serviceName}
}

func (l *Logger) Info(action, message, requestID string) {
	l.log("INFO", action, message, requestID, nil)
}

func (l *Logger) Debug(action, message, requestID string) {
	l.log("DEBUG", action, message, requestID, nil)
}

func (l *Logger) Error(action, message, requestID string, err error) {
	errorData := map[string]interface{}{
		"msg":   err.Error(),
		"stack": fmt.Sprintf("%+v", err),
	}
	l.log("ERROR", action, message, requestID, errorData)
}

func (l *Logger) log(level, action, message, requestID string, errorData map[string]interface{}) {
	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Service:   l.serviceName,
		Action:    action,
		Message:   message,
		Hostname:  getHostname(),
		RequestID: requestID,
		Error:     errorData,
	}

	jsonData, _ := json.Marshal(logEntry)
	fmt.Println(string(jsonData))
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
