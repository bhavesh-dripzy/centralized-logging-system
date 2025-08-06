package main

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

type LogEntry struct {
	Timestamp     string `json:"timestamp"`
	EventCategory string `json:"event.category"`
	Username      string `json:"username"`
	Hostname      string `json:"hostname"`
	Severity      string `json:"severity"`
	RawMessage    string `json:"raw.message"`
	IsBlacklisted bool   `json:"is.blacklisted"`
	EventSource   string `json:"event.source.type"`
}

// Dummy parser — paste your real one here
func parseLogMessage(msg string) LogEntry {
	severity := "INFO"
	if strings.Contains(msg, "<") && strings.Contains(msg, ">") {
		start := strings.Index(msg, "<") + 1
		end := strings.Index(msg, ">")
		code := msg[start:end]
		_ = code // map code to severity if needed
	}

	hostname := ""
	username := ""
	category := "login.audit" // fallback default

	// Extract hostname
	hostRegex := regexp.MustCompile(`<\d+>\s*([^\s]+)`)
	if matches := hostRegex.FindStringSubmatch(msg); len(matches) > 1 {
		hostname = matches[1]
	}

	// Extract username
	userRegex := regexp.MustCompile(`user (\w+)`)
	if matches := userRegex.FindStringSubmatch(msg); len(matches) > 1 {
		username = matches[1]
	}

	// Infer category
	if strings.Contains(msg, "session opened") {
		category = "login.audit"
	} else if strings.Contains(msg, "session closed") {
		category = "logout.audit"
	}

	return LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		EventCategory: category,
		Username:      username,
		Hostname:      hostname,
		Severity:      severity,
		EventSource:   "linux",
	}
}

func TestParseLogMessage(t *testing.T) {
	message := "<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)"
	expected := LogEntry{
		Hostname:      "aiops9242",
		Username:      "root",
		EventCategory: "login.audit",
		Severity:      "INFO",
		EventSource:   "linux",
	}

	result := parseLogMessage(message)

	if result.Hostname != expected.Hostname {
		t.Errorf("Expected hostname %s, got %s", expected.Hostname, result.Hostname)
	}
	if result.Username != expected.Username {
		t.Errorf("Expected username %s, got %s", expected.Username, result.Username)
	}
	if result.EventCategory != expected.EventCategory {
		t.Errorf("Expected category %s, got %s", expected.EventCategory, result.EventCategory)
	}
	if result.Severity != expected.Severity {
		t.Errorf("Expected severity %s, got %s", expected.Severity, result.Severity)
	}
	if result.EventSource != expected.EventSource {
		t.Errorf("Expected event source %s, got %s", expected.EventSource, result.EventSource)
	}
}
