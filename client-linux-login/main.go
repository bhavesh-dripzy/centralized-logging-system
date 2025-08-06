package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type LogMessage struct {
	Timestamp       string `json:"timestamp"`
	Hostname        string `json:"hostname"`
	EventSourceType string `json:"event.source.type"`
	EventCategory   string `json:"event.category"`
	Message         string `json:"message"`
}

func main() {
	// Connect to the log collector
	conn, err := net.Dial("tcp", "log-collector:9000")
	if err != nil {
		panic("Failed to connect to log collector: " + err.Error())
	}
	defer conn.Close()

	fmt.Println("Connected to log collector...")

	for {
		log := LogMessage{
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			Hostname:        "aiops9242",
			EventSourceType: "linux",
			EventCategory:   "login.audit",
			Message:         "<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)",
		}

		data, _ := json.Marshal(log)
		// Send log followed by newline (for scanner in collector)
		fmt.Fprintf(conn, string(data)+"\n")

		fmt.Println("Sent log at", log.Timestamp)
		time.Sleep(2 * time.Second)
	}
}
