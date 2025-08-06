package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

type RawLog struct {
	Timestamp       string `json:"timestamp"`
	Hostname        string `json:"hostname"`
	EventSourceType string `json:"event.source.type"`
	EventCategory   string `json:"event.category"`
	Message         string `json:"message"`
}

type EnrichedLog struct {
	Timestamp     string `json:"timestamp"`
	EventCategory string `json:"event.category"`
	Username      string `json:"username"`
	Hostname      string `json:"hostname"`
	Severity      string `json:"severity"`
	RawMessage    string `json:"raw.message"`
	IsBlacklisted bool   `json:"is.blacklisted"`
}

var blacklist = map[string]bool{
	"root":     true,
	"motadata": true,
}

func main() {
	fmt.Println("Log Collector started on :9000")

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Error starting TCP listener: %v", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var rawLog RawLog
		err := json.Unmarshal(scanner.Bytes(), &rawLog)
		if err != nil {
			log.Println("Invalid log format:", err)
			continue
		}

		// Process each log concurrently
		go processLog(rawLog)
	}
}

func processLog(raw RawLog) {
	enriched := EnrichedLog{
		Timestamp:     raw.Timestamp,
		EventCategory: raw.EventCategory,
		Hostname:      raw.Hostname,
		RawMessage:    raw.Message,
		Username:      extractUsername(raw.Message),
		Severity:      extractSeverity(raw.Message),
	}

	// Enrich with blacklist info
	enriched.IsBlacklisted = blacklist[enriched.Username]

	// Forward to log server
	forwardToLogServer(enriched)
}

func extractUsername(msg string) string {
	re := regexp.MustCompile(`user (\w+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

func extractSeverity(msg string) string {
	if strings.Contains(msg, "<86>") {
		return "INFO"
	}
	return "UNKNOWN"
}

func forwardToLogServer(logEntry EnrichedLog) {
	jsonData, _ := json.Marshal(logEntry)
	resp, err := http.Post("http://log-server:8080/ingest", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Failed to send log to server:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		log.Println("Server rejected log:", resp.Status)
	}
}
