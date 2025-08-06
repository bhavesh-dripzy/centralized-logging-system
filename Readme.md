# Centralized Logging System with Golang Microservices and Docker

## Overview

This project implements a fully functional centralized logging system using Go microservices and Docker. It simulates Linux system logs from client services, collects and enriches them using a log collector service, and stores them in a central logging server with a queryable HTTP API.

The key components are:

* `client-linux-login`: Simulates system login logs and sends them via TCP.
* `log-collector`: Parses, enriches, and forwards logs to the central server.
* `log-server`: Stores logs and exposes HTTP APIs to ingest, query, and fetch metrics.

---

## Architecture

```
+------------------+           +------------------+           +------------------+
| client-linux-login |  --->   |   log-collector   |  --->   |    log-server     |
+------------------+    TCP    +------------------+   HTTP   +------------------+
                                                  \
                                                   +-- GET /logs
                                                   +-- GET /metrics
```

---

## Prerequisites

* Docker
* Docker Compose
* Go (for local testing)

---

## Getting Started

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd logging_system
```

### 2. Run the Full System with Docker

```bash
docker-compose up --build
```

This will start all three services:

* `log-server` on port `8080`
* `log-collector` on port `9000`
* `client-linux-login` will begin sending logs every 2 seconds

---

## API Reference

### 1. Ingest Logs (internal use only)

```
POST /ingest
Content-Type: application/json
```

**Payload Example:**

```json
{
  "timestamp": "2025-08-06T07:47:31Z",
  "event.category": "login.audit",
  "username": "root",
  "hostname": "aiops9242",
  "severity": "INFO",
  "raw.message": "<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0)",
  "is.blacklisted": false,
  "event.source.type": "linux"
}
```

### 2. Query Logs

```
GET /logs
```

**Query Parameters:**

* `username=root`
* `event.category=login.audit`
* `level=info`
* `is.blacklisted=true`
* `limit=10`
* `sort=timestamp`

**Example:**

```bash
curl "http://localhost:8080/logs?username=root&is.blacklisted=true&limit=5&sort=timestamp"
```

### 3. View Metrics

```
GET /metrics
```

**Response Example:**

```json
{
  "total_logs": 29,
  "by_category": {
    "login.audit": 29
  },
  "by_severity": {
    "INFO": 29
  }
}
```

---

## Project Structure

```
logging_system/
├── client-linux-login/
│   ├── main.go
│   └── Dockerfile
│
├── log-collector/
│   ├── main.go
│   ├── parser_test.go
│   └── Dockerfile
│
├── log-server/
│   ├── main.go
│   ├── server_test.go
│   └── Dockerfile
│
├── docker-compose.yml
└── README.md
```

---

## Testing

### Unit Tests

Run tests inside each service directory:

```bash
cd log-server
go test -v

cd ../log-collector
go test -v
```

### Sample Output

```
=== RUN   TestIngestLogHandler
--- PASS: TestIngestLogHandler (0.00s)
=== RUN   TestQueryHandler_Filtering
--- PASS: TestQueryHandler_Filtering (0.00s)
=== RUN   TestMetricsHandler
--- PASS: TestMetricsHandler (0.00s)
PASS
```

---

## Docker Containers

### docker-compose.yml

```yaml
version: '3.8'

services:
  log-server:
    build: ./log-server
    ports:
      - "8080:8080"

  log-collector:
    build: ./log-collector
    ports:
      - "9000:9000"
    depends_on:
      - log-server

  client-linux-login:
    build: ./client-linux-login
    depends_on:
      - log-collector
```

---

## Blacklist Configuration

In `log-collector/main.go`, the blacklist is hardcoded:

```go
var blacklist = map[string]bool{
    "root": true,
    "motadata": true,
}
```

Logs from these users will be marked as blacklisted.

---

## Log Parsing Example

Input message:

```
<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)
```

Parsed fields:

* severity: INFO
* hostname: aiops9242
* username: root
* category: login.audit

---

## Notes

* The system uses in-memory storage, so logs are ephemeral.
* Add persistence using a database like MongoDB or PostgreSQL if needed.
* Change log generation rate in `client-linux-login/main.go` via `time.Sleep()`.

---

## Summary

This system demonstrates end-to-end log generation, collection, enrichment, storage, and querying across microservices using Go and Docker. It handles concurrency via goroutines and is designed for extensibility.

To stop the system:

```bash
docker-compose down
```
