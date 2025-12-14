# Log Cheatsheet

**Repository:** https://github.com/charmbracelet/log  
**Purpose:** Beautiful, human-readable structured logging with customizable styling

## Basic Usage

### Simple Logging
```go
import "github.com/charmbracelet/log"

func main() {
    // Basic logging
    log.Debug("Debug message")
    log.Info("Information message")
    log.Warn("Warning message")
    log.Error("Error message")
    log.Fatal("Fatal message") // Exits program
    
    // With structured fields
    log.Info("User logged in", "user", "john", "session", "abc123")
    log.Error("Database connection failed", "err", err, "host", "localhost")
    log.Warn("High memory usage", "usage", "85%", "threshold", "80%")
}
```

### Logger Creation
```go
// Default logger (logs to stderr)
logger := log.New(os.Stderr)

// With options
logger := log.NewWithOptions(os.Stderr, log.Options{
    ReportCaller:    true,           // Show file:line
    ReportTimestamp: true,           // Show timestamp
    TimeFormat:      time.Kitchen,   // Time format
    Prefix:          "ZeroUI", // Log prefix
})

// Set log level
logger.SetLevel(log.DebugLevel)
```

## Log Levels

### Available Levels
```go
log.DebugLevel  // -4: Detailed debugging info
log.InfoLevel   //  0: General information  
log.WarnLevel   //  4: Warning conditions
log.ErrorLevel  //  8: Error conditions
log.FatalLevel  // 12: Fatal errors (exits program)

// Setting levels
log.SetLevel(log.DebugLevel)  // Show all levels
log.SetLevel(log.InfoLevel)   // Hide debug messages
log.SetLevel(log.WarnLevel)   // Show only warnings and errors
log.SetLevel(log.ErrorLevel)  // Show only errors and fatal
```

### Level Checking
```go
if log.GetLevel() <= log.DebugLevel {
    // Expensive debug operations
    data := collectDebugData()
    log.Debug("Debug data", "data", data)
}

// Helper methods
logger.DebugEnabled() // Returns true if debug logging is enabled
```

## Structured Logging

### Key-Value Pairs
```go
// Basic key-value logging
log.Info("User action", "user", "alice", "action", "login")
log.Error("Operation failed", "op", "save", "file", "config.json", "err", err)

// Multiple pairs
log.Info("Request processed",
    "method", "GET",
    "path", "/api/users",
    "status", 200,
    "duration", time.Since(start),
    "ip", "192.168.1.1",
)
```

### Context Loggers
```go
// Create sub-logger with persistent fields
userLogger := log.With("user", "john", "session", "abc123")
userLogger.Info("Logged in")         // Includes user and session
userLogger.Warn("Invalid operation") // Includes user and session
userLogger.Error("Permission denied") // Includes user and session

// Chain multiple contexts
requestLogger := log.With("request_id", requestID)
userRequestLogger := requestLogger.With("user", userID)
userRequestLogger.Info("Processing request")
```

### Complex Field Types
```go
// Different value types
log.Info("Complex data",
    "string", "text",
    "int", 42,
    "float", 3.14,
    "bool", true,
    "duration", 5*time.Second,
    "time", time.Now(),
    "struct", struct{ Name string }{"test"},
    "slice", []string{"a", "b", "c"},
    "map", map[string]int{"key": 123},
)

// Error handling
if err != nil {
    log.Error("Operation failed",
        "operation", "file_read",
        "file", filename,
        "error", err,
        "retry_count", retries,
    )
}
```

## Styling and Customization

### Custom Styles
```go
import "github.com/charmbracelet/lipgloss"

// Get default styles
styles := log.DefaultStyles()

// Customize level styles
styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
    SetString("ERROR").
    Padding(0, 1, 0, 1).
    Background(lipgloss.Color("204")).
    Foreground(lipgloss.Color("0"))

styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
    SetString("WARN").
    Padding(0, 1, 0, 1).
    Background(lipgloss.Color("220")).
    Foreground(lipgloss.Color("0"))

styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
    SetString("INFO").
    Padding(0, 1, 0, 1).
    Background(lipgloss.Color("86")).
    Foreground(lipgloss.Color("0"))

// Apply styles
logger := log.New(os.Stderr)
logger.SetStyles(styles)
```

### Complete Style Customization
```go
customStyles := &log.Styles{
    // Timestamp styling
    Timestamp: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
    
    // Caller (file:line) styling
    Caller: lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
    
    // Prefix styling
    Prefix: lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("13")),
    
    // Message styling
    Message: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
    
    // Key styling (for key-value pairs)
    Key: lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
    
    // Value styling
    Value: lipgloss.NewStyle().Foreground(lipgloss.Color("37")),
    
    // Separator between key-value pairs
    Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
    
    // Level-specific styles
    Levels: map[log.Level]lipgloss.Style{
        log.DebugLevel: lipgloss.NewStyle().
            SetString("DEBUG").
            Foreground(lipgloss.Color("8")),
        log.InfoLevel: lipgloss.NewStyle().
            SetString("INFO ").
            Foreground(lipgloss.Color("12")),
        log.WarnLevel: lipgloss.NewStyle().
            SetString("WARN ").
            Foreground(lipgloss.Color("11")),
        log.ErrorLevel: lipgloss.NewStyle().
            SetString("ERROR").
            Foreground(lipgloss.Color("9")),
        log.FatalLevel: lipgloss.NewStyle().
            SetString("FATAL").
            Foreground(lipgloss.Color("1")),
    },
}

logger.SetStyles(customStyles)
```

## Output Formats

### Text Format (Default)
```go
// Default human-readable format
logger := log.New(os.Stderr)
logger.Info("User login", "user", "alice", "ip", "192.168.1.1")
// Output: INFO User login user=alice ip=192.168.1.1
```

### JSON Format
```go
// JSON output for machine processing
logger := log.New(os.Stderr)
logger.SetFormatter(log.JSONFormatter)
logger.Info("User login", "user", "alice", "ip", "192.168.1.1")
// Output: {"level":"info","msg":"User login","time":"2023-...","user":"alice","ip":"192.168.1.1"}
```

### Logfmt Format
```go
// Logfmt format
logger := log.New(os.Stderr)
logger.SetFormatter(log.LogfmtFormatter)
logger.Info("User login", "user", "alice", "ip", "192.168.1.1")
// Output: level=info msg="User login" time=2023-... user=alice ip=192.168.1.1
```

## File Logging

### Log to File
```go
// Create or append to log file
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal("Failed to open log file", "err", err)
}
defer file.Close()

// Create logger with file output
logger := log.NewWithOptions(file, log.Options{
    ReportTimestamp: true,
    TimeFormat:      time.RFC3339,
    Prefix:          "zeroui",
})
```

### Multiple Outputs
```go
// Log to both file and stderr
logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal("Failed to open log file", "err", err)
}

multiWriter := io.MultiWriter(os.Stderr, logFile)
logger := log.New(multiWriter)
```

### Rotating Logs
```go
// Using external rotation (example with lumberjack)
import "gopkg.in/natefinch/lumberjack.v2"

lumber := &lumberjack.Logger{
    Filename:   "app.log",
    MaxSize:    1,    // MB
    MaxBackups: 3,
    MaxAge:     28,   // days
    Compress:   true,
}

logger := log.New(lumber)
```

## Advanced Features

### Custom Formatters
```go
// Custom formatter function
func customFormatter(keyvals ...interface{}) string {
    var b strings.Builder
    
    // Extract level and message
    level := keyvals[1].(log.Level)
    msg := keyvals[3].(string)
    
    b.WriteString(fmt.Sprintf("[%s] %s", level.String(), msg))
    
    // Add key-value pairs
    for i := 4; i < len(keyvals); i += 2 {
        if i+1 < len(keyvals) {
            key := keyvals[i]
            val := keyvals[i+1]
            b.WriteString(fmt.Sprintf(" %s=%v", key, val))
        }
    }
    
    b.WriteString("\n")
    return b.String()
}

logger := log.New(os.Stderr)
logger.SetFormatter(log.FormatterFunc(customFormatter))
```

### Conditional Logging
```go
func logIfError(err error, msg string, keyvals ...interface{}) {
    if err != nil {
        args := append(keyvals, "error", err)
        log.Error(msg, args...)
    }
}

func logIfSlow(duration time.Duration, threshold time.Duration, op string) {
    if duration > threshold {
        log.Warn("Slow operation detected",
            "operation", op,
            "duration", duration,
            "threshold", threshold,
        )
    }
}
```

### Performance Considerations
```go
// Lazy evaluation for expensive operations
log.Debug("Expensive debug info", "data", func() interface{} {
    return collectExpensiveDebugData()
})

// Check level before expensive operations
if logger.DebugEnabled() {
    data := expensiveDebugOperation()
    logger.Debug("Debug data collected", "data", data)
}
```

## ZeroUI Integration Examples

### Application Logger Setup
```go
func setupLogger(logLevel string, logFile string) *log.Logger {
    var writer io.Writer = os.Stderr
    
    // Setup file output if specified
    if logFile != "" {
        file, err := os.OpenFile(logFile, 
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            log.Fatal("Cannot open log file", "file", logFile, "error", err)
        }
        writer = io.MultiWriter(os.Stderr, file)
    }
    
    logger := log.NewWithOptions(writer, log.Options{
        ReportCaller:    logLevel == "debug",
        ReportTimestamp: true,
        TimeFormat:      "15:04:05",
        Prefix:          "zeroui",
    })
    
    // Set log level
    switch strings.ToLower(logLevel) {
    case "debug":
        logger.SetLevel(log.DebugLevel)
    case "info":
        logger.SetLevel(log.InfoLevel)
    case "warn":
        logger.SetLevel(log.WarnLevel)
    case "error":
        logger.SetLevel(log.ErrorLevel)
    default:
        logger.SetLevel(log.InfoLevel)
    }
    
    // Custom ZeroUI styling
    styles := log.DefaultStyles()
    styles.Key = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))     // Blue
    styles.Value = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))   // Light gray
    styles.Prefix = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("13"))  // Purple
    
    logger.SetStyles(styles)
    return logger
}
```

### Structured Logging for Operations
```go
func LogConfigOperation(logger *log.Logger, op, app, key string, oldVal, newVal interface{}, success bool, duration time.Duration) {
    level := log.Info
    message := "Configuration operation completed"
    
    if !success {
        level = log.Error
        message = "Configuration operation failed"
    }
    
    logger.Log(level, message,
        "operation", op,
        "application", app,
        "key", key,
        "old_value", oldVal,
        "new_value", newVal,
        "success", success,
        "duration", duration,
    )
}

func LogApplicationAction(logger *log.Logger, action, app string, details map[string]interface{}) {
    args := []interface{}{
        "action", action,
        "application", app,
    }
    
    // Add details as key-value pairs
    for k, v := range details {
        args = append(args, k, v)
    }
    
    logger.Info("Application action", args...)
}
```

### Error Logging with Context
```go
type ContextLogger struct {
    *log.Logger
    context map[string]interface{}
}

func NewContextLogger(base *log.Logger) *ContextLogger {
    return &ContextLogger{
        Logger:  base,
        context: make(map[string]interface{}),
    }
}

func (cl *ContextLogger) WithContext(key string, value interface{}) *ContextLogger {
    newContext := make(map[string]interface{})
    for k, v := range cl.context {
        newContext[k] = v
    }
    newContext[key] = value
    
    return &ContextLogger{
        Logger:  cl.Logger,
        context: newContext,
    }
}

func (cl *ContextLogger) log(level log.Level, msg string, keyvals ...interface{}) {
    // Merge context with keyvals
    args := make([]interface{}, 0, len(cl.context)*2+len(keyvals))
    
    for k, v := range cl.context {
        args = append(args, k, v)
    }
    args = append(args, keyvals...)
    
    cl.Logger.Log(level, msg, args...)
}

func (cl *ContextLogger) Info(msg string, keyvals ...interface{}) {
    cl.log(log.InfoLevel, msg, keyvals...)
}

func (cl *ContextLogger) Error(msg string, keyvals ...interface{}) {
    cl.log(log.ErrorLevel, msg, keyvals...)
}

// Usage
appLogger := NewContextLogger(logger).WithContext("app", "vscode")
appLogger.Info("Configuration loaded", "file", "settings.json")
// Outputs: INFO Configuration loaded app=vscode file=settings.json
```

### Audit Logging
```go
func setupAuditLogger(auditFile string) *log.Logger {
    file, err := os.OpenFile(auditFile, 
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Fatal("Cannot create audit log", "file", auditFile, "error", err)
    }
    
    auditLogger := log.NewWithOptions(file, log.Options{
        ReportTimestamp: true,
        TimeFormat:      time.RFC3339,
        Prefix:          "AUDIT",
    })
    
    // Use JSON format for audit logs
    auditLogger.SetFormatter(log.JSONFormatter)
    return auditLogger
}

func AuditConfigChange(logger *log.Logger, user, app, key string, oldVal, newVal interface{}) {
    logger.Info("Configuration changed",
        "user", user,
        "application", app,
        "key", key,
        "old_value", oldVal,
        "new_value", newVal,
        "timestamp", time.Now(),
        "action", "config_change",
    )
}
```

### Performance Logging
```go
func LogPerformanceMetrics(logger *log.Logger, operation string, metrics map[string]interface{}) {
    args := []interface{}{"operation", operation}
    
    for metric, value := range metrics {
        args = append(args, metric, value)
    }
    
    logger.Info("Performance metrics", args...)
}

// Usage
func (ct *ZeroUI) Toggle(app, key string) error {
    start := time.Now()
    defer func() {
        metrics := map[string]interface{}{
            "duration": time.Since(start),
            "app":      app,
            "key":      key,
        }
        LogPerformanceMetrics(ct.logger, "toggle", metrics)
    }()
    
    // ... toggle logic
    return nil
}
```