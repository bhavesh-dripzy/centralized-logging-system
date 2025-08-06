package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestLogHandler(t *testing.T) {
	// Clear store before test
	store.Clear()

	entry := LogEntry{
		Timestamp:     "2025-08-06T07:47:31Z",
		EventCategory: "login.audit",
		Username:      "root",
		Hostname:      "aiops9242",
		Severity:      "INFO",
		RawMessage:    "<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0)",
		IsBlacklisted: false,
		EventSource:   "linux",
	}

	payload, _ := json.Marshal(entry)
	req := httptest.NewRequest("POST", "/ingest", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ingestLogHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	results := store.Query(map[string]string{"username": "root"})
	if len(results) != 1 {
		t.Fatalf("Expected 1 log in store, got %d", len(results))
	}
	if results[0].Username != entry.Username {
		t.Errorf("Expected username %s, got %s", entry.Username, results[0].Username)
	}
}

func TestQueryHandler_Filtering(t *testing.T) {
	// Use store abstraction instead of logStore
	store.Clear()

	store.Save(LogEntry{
		Timestamp:     "2025-08-06T07:47:31Z",
		EventCategory: "login.audit",
		Username:      "admin",
		Hostname:      "host1",
		Severity:      "INFO",
		RawMessage:    "msg1",
		IsBlacklisted: false,
		EventSource:   "linux",
	})
	store.Save(LogEntry{
		Timestamp:     "2025-08-06T07:47:32Z",
		EventCategory: "login.audit",
		Username:      "root",
		Hostname:      "host2",
		Severity:      "ERROR",
		RawMessage:    "msg2",
		IsBlacklisted: true,
		EventSource:   "linux",
	})

	req := httptest.NewRequest("GET", "/logs?username=root&is.blacklisted=true", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(queryHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var result []LogEntry
	json.NewDecoder(rr.Body).Decode(&result)

	if len(result) != 1 {
		t.Fatalf("Expected 1 matching log, got %d", len(result))
	}
	if result[0].Username != "root" || !result[0].IsBlacklisted {
		t.Errorf("Filtered log mismatch: %+v", result[0])
	}
}

func TestMetricsHandler(t *testing.T) {
	// Clear store before test
	store.Clear()

	// Add 2 log entries
	logs := []LogEntry{
		{
			Timestamp:     "2025-08-06T07:47:31Z",
			EventCategory: "login.audit",
			Username:      "admin",
			Hostname:      "host1",
			Severity:      "INFO",
			RawMessage:    "msg1",
			IsBlacklisted: false,
			EventSource:   "linux",
		},
		{
			Timestamp:     "2025-08-06T07:48:00Z",
			EventCategory: "login.audit",
			Username:      "root",
			Hostname:      "host2",
			Severity:      "ERROR",
			RawMessage:    "msg2",
			IsBlacklisted: true,
			EventSource:   "linux",
		},
	}

	for _, entry := range logs {
		store.Save(entry)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(metricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var metrics map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &metrics)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if metrics["total_logs"] != float64(2) {
		t.Errorf("Expected total_logs to be 2, got %v", metrics["total_logs"])
	}

	categoryCounts := metrics["by_category"].(map[string]interface{})
	if categoryCounts["login.audit"] != float64(2) {
		t.Errorf("Expected login.audit category count 2, got %v", categoryCounts["login.audit"])
	}

	severityCounts := metrics["by_severity"].(map[string]interface{})
	if severityCounts["INFO"] != float64(1) || severityCounts["ERROR"] != float64(1) {
		t.Errorf("Expected severity INFO:1 and ERROR:1, got %+v", severityCounts)
	}
}
