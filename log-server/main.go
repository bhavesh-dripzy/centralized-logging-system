package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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

// --- INTERFACE for Log Storage ---
type Storage interface {
	Save(LogEntry)
	Query(params map[string]string) []LogEntry
	Metrics() map[string]interface{}
	Clear() // for test use
}

// --- In-memory Storage Implementation ---
type InMemoryStore struct {
	mu            sync.Mutex
	logs          []LogEntry
	totalLogs     int
	categoryCount map[string]int
	severityCount map[string]int
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		logs:          []LogEntry{},
		categoryCount: make(map[string]int),
		severityCount: make(map[string]int),
	}
}

func (s *InMemoryStore) Save(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, entry)
	s.totalLogs++
	s.categoryCount[entry.EventCategory]++
	s.severityCount[strings.ToUpper(entry.Severity)]++
}

func (s *InMemoryStore) Query(params map[string]string) []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := []LogEntry{}

	for _, log := range s.logs {
		match := true
		for key, val := range params {
			val = strings.ToLower(val)
			switch key {
			case "service", "event.category":
				if strings.ToLower(log.EventCategory) != val {
					match = false
				}
			case "level", "severity":
				if strings.ToLower(log.Severity) != val {
					match = false
				}
			case "username":
				if strings.ToLower(log.Username) != val {
					match = false
				}
			case "is.blacklisted":
				if val == "true" && !log.IsBlacklisted {
					match = false
				}
				if val == "false" && log.IsBlacklisted {
					match = false
				}
			}
		}
		if match {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

func (s *InMemoryStore) Metrics() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return map[string]interface{}{
		"total_logs":  s.totalLogs,
		"by_category": s.categoryCount,
		"by_severity": s.severityCount,
	}
}

func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = []LogEntry{}
	s.totalLogs = 0
	s.categoryCount = make(map[string]int)
	s.severityCount = make(map[string]int)
}

// --- Use Storage Interface Globally ---
var store Storage = NewInMemoryStore()

// --- HTTP HANDLERS ---
func ingestLogHandler(w http.ResponseWriter, r *http.Request) {
	var entry LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	store.Save(entry)
	log.Printf("Log stored: %s - %s\n", entry.Username, entry.EventCategory)
	w.WriteHeader(http.StatusOK)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	filtered := store.Query(params)

	// Sort if requested
	if strings.ToLower(params["sort"]) == "timestamp" {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Timestamp < filtered[j].Timestamp
		})
	}

	// Limit if requested
	if limitStr, ok := params["limit"]; ok {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(filtered) {
			filtered = filtered[:limit]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := store.Metrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// --- MAIN ---
func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/ingest", ingestLogHandler)
	http.HandleFunc("/logs", queryHandler)
	http.HandleFunc("/metrics", metricsHandler)

	log.Println("Log Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
