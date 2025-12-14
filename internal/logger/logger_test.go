package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLevelParsingBlocksLowerEvents(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(&Config{Level: "warn", Format: "json", Output: buf, EnableTracing: false})

	log.Info("hidden message")
	if buf.Len() != 0 {
		t.Fatalf("expected no output for info at warn level, got %s", buf.String())
	}

	log.Warn("visible warning")

	var event map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if event["level"] != "warn" {
		t.Fatalf("expected warn level, got %v", event["level"])
	}
}

func TestConsoleAndJSONFormatting(t *testing.T) {
	jsonBuf := &bytes.Buffer{}
	jsonLogger := New(&Config{Level: "info", Format: "json", Output: jsonBuf, EnableTracing: false})
	jsonLogger.Info("structured", Field{Key: "user", Value: "alice"})

	var event map[string]interface{}
	if err := json.Unmarshal(jsonBuf.Bytes(), &event); err != nil {
		t.Fatalf("failed to unmarshal json log: %v", err)
	}
	if event["user"] != "alice" {
		t.Fatalf("expected user field to be preserved, got %v", event["user"])
	}

	consoleBuf := &bytes.Buffer{}
	consoleLogger := New(&Config{Level: "info", Format: "console", Output: consoleBuf, EnableTracing: false})
	consoleLogger.Info("hello world", Field{Key: "user", Value: "bob"})

	consoleOutput := consoleBuf.String()
	if !strings.Contains(consoleOutput, "hello world") {
		t.Fatalf("console output missing message: %s", consoleOutput)
	}
	if !strings.Contains(strings.ToUpper(consoleOutput), "INF") {
		t.Fatalf("console output missing level: %s", consoleOutput)
	}
}

func TestContextualFieldsAndRedaction(t *testing.T) {
	buf := &bytes.Buffer{}
	scoped := New(&Config{Level: "debug", Format: "json", Output: buf, EnableTracing: true})

	scoped = scoped.With(
		Field{Key: "app", Value: "demo"},
		Field{Key: "secret", Value: "token", Redact: true},
	)

	scoped.Info("with context", Field{Key: "id", Value: 42})

	var event map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("failed to decode contextual log: %v", err)
	}

	if event["app"] != "demo" {
		t.Fatalf("expected app field, got %v", event["app"])
	}
	if event["secret"] != "[REDACTED]" {
		t.Fatalf("expected secret to be redacted, got %v", event["secret"])
	}
	if id, ok := event["id"].(float64); !ok || id != 42 {
		t.Fatalf("expected id field to be 42, got %v", event["id"])
	}
	if phase, ok := event["trace_phase"].(string); !ok || phase == "" {
		t.Fatalf("expected tracing metadata to be present, got %v", event["trace_phase"])
	}
}
