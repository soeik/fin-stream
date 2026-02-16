# FinStream Engine 🚀

**High-Performance Financial Data Aggregator & Middleware**

FinStream Engine is a low-latency streaming middleware built in Go. It ingests raw, high-frequency WebSocket feeds from financial exchanges (e.g., Binance), processes them in-memory using concurrent pipelines, and delivers optimized, aggregated data streams to frontend applications.

## 🎯 The Challenge
Public crypto exchange APIs generate massive event volumes (10k+ ticks/sec). Direct streaming to the frontend causes:
1. **Browser Bottlenecks:** Rendering engines choke on high-frequency updates.
2. **Network Overhead:** Massive bandwidth consumption for redundant data.
3. **Calculation Complexity:** Heavy lifting for metrics (Moving Averages, Volatility) shouldn't happen on the client side.

## 🏗 Architectural Design

The system follows **Clean Architecture** principles and utilizes a decoupled **Producer-Consumer** model.

### 1. Ingestion Layer
* **Resilient WS Client:** Implements **Exponential Backoff** for robust reconnection logic.
* **Context-Driven Lifecycle:** Uses `context.Context` for cancellation propagation and **Graceful Shutdown**, ensuring no data is lost during deployments.

### 2. Processing Layer (The Engine)
* **Worker Pool Pattern:** A fixed pool of goroutines handles deserialization and business logic, preventing "goroutine leaks" and CPU spikes.
* **Zero-Allocation Focus:** Leverages `sync.Pool` to reuse JSON structs. This drastically reduces **GC (Garbage Collector) pressure** and stabilizes P99 latency.
* **Concurrency Primitives:** High-speed data passing via buffered channels.

### 3. State Management (Aggregation)
* **Sliding Window Algorithm:** Computes real-time metrics (e.g., 5s Average Price) in-memory with $O(1)$ complexity.
* **Thread Safety:** Optimized state access using `sync.RWMutex`, favoring high-frequency reads from the Delivery Layer.

### 4. Delivery Layer
* **Throttling & Batching:** Instead of "fire-and-forget" forwarding, the engine batches updates and flushes them to clients at fixed intervals (e.g., every 100ms).
* **Traffic Optimization:** Reduces WebSocket frame overhead by merging multiple price updates into a single payload.

## 🛠 Tech Stack
* **Runtime:** Go (Golang)
* **Patterns:** Worker Pools, Pipelines, Singleton, Observer.
* **Concurrency:** Goroutines, Channels, `sync.Pool`, `errgroup`.
* **Libraries:** [Gorilla WebSocket](https://github.com), [Prometheus Go Client](https://github.com).

## 🚀 Getting Started

### Local Setup
```bash
go mod download
go run cmd/main.go
