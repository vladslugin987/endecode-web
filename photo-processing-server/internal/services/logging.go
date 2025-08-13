package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// BroadcastFunc is a function type for broadcasting log messages
type BroadcastFunc func(message string)

// Logger represents our logging service (port of ConsoleState.kt)
type Logger struct {
	logs        []LogEntry
	mutex       sync.RWMutex
	logger      *logrus.Logger
	broadcaster BroadcastFunc
}

// LogEntry represents a single log entry
type LogEntry struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	return &Logger{
		logs:        make([]LogEntry, 0),
		mutex:       sync.RWMutex{},
		logger:      logger,
		broadcaster: nil,
	}
}

// SetLevel sets the log level from string (debug, info, warn, error)
func (l *Logger) SetLevel(level string) {
    switch level {
    case "debug":
        l.logger.SetLevel(logrus.DebugLevel)
    case "info":
        l.logger.SetLevel(logrus.InfoLevel)
    case "warn", "warning":
        l.logger.SetLevel(logrus.WarnLevel)
    case "error":
        l.logger.SetLevel(logrus.ErrorLevel)
    default:
        l.logger.SetLevel(logrus.InfoLevel)
    }
}

// SetWebSocketBroadcaster sets the WebSocket broadcaster function
func (l *Logger) SetWebSocketBroadcaster(broadcaster BroadcastFunc) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.broadcaster = broadcaster
}

// Log adds a message to logs and prints to console (equivalent to ConsoleState.log())
func (l *Logger) Log(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	entry := LogEntry{
		Message:   message,
		Timestamp: time.Now(),
		Level:     "INFO",
	}
	
	l.logs = append(l.logs, entry)
	
	// Print to console (equivalent to println(message) in Kotlin)
	fmt.Println(message)
	l.logger.Info(message)
	
	// Broadcast to WebSocket clients if broadcaster is set
	if l.broadcaster != nil {
		l.broadcaster(message)
	}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.Log(message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	entry := LogEntry{
		Message:   message,
		Timestamp: time.Now(),
		Level:     "ERROR",
	}
	
	l.logs = append(l.logs, entry)
	
	fmt.Printf("ERROR: %s\n", message)
	l.logger.Error(message)
	
	// Broadcast to WebSocket clients if broadcaster is set
	if l.broadcaster != nil {
		l.broadcaster(fmt.Sprintf("ERROR: %s", message))
	}
}

// Success logs a success message with checkmark (like Kotlin version)
func (l *Logger) Success(filename string) {
	message := fmt.Sprintf("%s: Success ✔", filename)
	l.Log(message)
}

// Processing logs a processing message
func (l *Logger) Processing(message string) {
	fullMessage := fmt.Sprintf("▷ %s", message)
	l.Log(fullMessage)
}

// GetLogs returns all logs (equivalent to ConsoleState.logs)
func (l *Logger) GetLogs() []LogEntry {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	// Return a copy to prevent concurrent modification
	result := make([]LogEntry, len(l.logs))
	copy(result, l.logs)
	return result
}

// Clear removes all logs (equivalent to ConsoleState.clear())
func (l *Logger) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	l.logs = make([]LogEntry, 0)
	l.logger.Info("Logs cleared")
}

// GetLogMessages returns just the messages as strings (for compatibility)
func (l *Logger) GetLogMessages() []string {
	logs := l.GetLogs()
	messages := make([]string, len(logs))
	for i, log := range logs {
		messages[i] = log.Message
	}
	return messages
}

// Global logger instance (similar to Kotlin object)
var globalLogger *Logger
var once sync.Once

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	once.Do(func() {
		globalLogger = NewLogger()
	})
	return globalLogger
}

// Helper functions for global access (similar to ConsoleState usage in Kotlin)

// Log logs a message using the global logger
func Log(message string) {
	GetGlobalLogger().Log(message)
}

// LogSuccess logs a success message using the global logger
func LogSuccess(filename string) {
	GetGlobalLogger().Success(filename)
}

// LogError logs an error message using the global logger
func LogError(message string) {
	GetGlobalLogger().Error(message)
}

// LogProcessing logs a processing message using the global logger
func LogProcessing(message string) {
	GetGlobalLogger().Processing(message)
}