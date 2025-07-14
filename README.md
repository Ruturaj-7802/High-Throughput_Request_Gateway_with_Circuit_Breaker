# High-Throughput Request Gateway with Circuit Breaker (Go)

## Overview

This project implements a **lightweight gateway** in Go that forwards requests to backend services using a **round-robin load balancer** and a **per-backend circuit breaker** to ensure fault isolation.

> This is a **Go implementation** of the assignment originally suggested for Python. The core principles remain intact.

---

## Core Features

### âœ“ REST Proxy API

- **Endpoint:**

  ```
  GET /v1/proxy/{service_name}
  ```

- Routes requests to one of the backend replicas registered for the service name.

---

### Config-Driven Routing

- Reads service â†’ list of backend URLs from a YAML config file.
- Example `config.yaml`:

```yaml
product:
  - http://localhost:9001
  - http://localhost:9002
```

---

### Round-Robin Routing

- Distributes incoming requests across backend replicas.
- Tracks round-robin index **per service**.

---

### Circuit Breaker Logic

Each backend URL is protected by a **circuit breaker**, which:

| State        | Behavior                                                                 |
|--------------|--------------------------------------------------------------------------|
| **Closed**   | All requests go through normally. Failures are tracked.                  |
| **Open**     | Circuit "trips" after `N` consecutive failures. Requests are blocked.    |
| **Half-Open**| After cooldown, one test request is sent. If successful â†’ Closed. Else â†’ Open. |

- **Failure Threshold:** 3
- **Open Timeout:** 10 seconds (for demo purposes)
- Circuit breaker state and metrics are stored **in memory**.

---

### Mock Backends

Use a helper server to simulate backend responses:

```bash
go run backend/main.go 9001
go run backend/main.go 9002
```

---

## How to Run

### 1. Clone & Build

```bash
git clone https://github.com/yourusername/request-gateway
cd request-gateway
go mod tidy
```

---

### 2. Start Backend Servers

In two terminals:

```bash
go run backend/main.go 9001
go run backend/main.go 9002
```

---

### 3. Start Gateway

```bash
go run main.go
```

---

### 4. Test It

```bash
curl http://localhost:8080/v1/proxy/product
```

- It will alternate between `9001` and `9002`.
- Stop one backend (e.g. `9002`) to observe circuit breaker behavior.

---

## Logs & Metrics

- Circuit state transitions are logged:

```bash
[CIRCUIT] http://localhost:9002 â†’ OPEN (threshold reached)
[CIRCUIT] http://localhost:9002 â†’ HALF-OPEN (timeout passed)
[CIRCUIT] http://localhost:9002 â†’ CLOSED (recovered)
```

- Per-backend metrics printed on each request:

```bash
[CIRCUIT STATUS] http://localhost:9002
  Total:    10
  Success:   6
  Failures:  4
  State:     Closed
```

---

## Project Structure

```
.
  backend/             # mock backend server
    main.go

  circuitbreaker/      # in-memory circuit breaker logic
    breaker.go

  config/
    config.yaml        # routing rules

  main.go              # entrypoint gateway
  README.md            # this file
```

---

## Design Notes

- **In-Memory Breakers:** Simple per-backend circuit breaker stored using a map.
- **Mutex Locks:** Ensures thread-safety for concurrent updates.
- **Extensibility:** Easily pluggable with service discovery or Prometheus if needed.

---

## FAQ

**Q: Can I test failure scenarios?**  
Yes. Kill or block one backend. You'll observe:
- Circuit moving to `OPEN`
- Requests skipping it
- Eventually `HALF-OPEN â†’ CLOSED` if recovered

**Q: Why Go instead of Python?**  
This implementation explores the same principles (routing, circuit breaking, resilience) in a performant, concurrent language.

---

## Author

- Ruturaj Patil
- GitHub: [ruturaj-7802](https://github.com/ruturaj-7802)
