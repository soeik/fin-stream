# 🚀 FinStream Engine: Ultra-Low Latency HFT Dashboard

**Full-stack Streaming Middleware & High-Frequency Visualization Suite**

FinStream is a professional-grade streaming engine designed to ingest, aggregate, and visualize high-frequency financial data (Binance `aggTrade`) with sub-millisecond overhead. It eliminates the "Browser Bottleneck" by shifting heavy lifting to a concurrent Go backend and utilizing binary protocols for delivery.

## ⚡ Key Performance Metrics (Measured on MacBook Air M1/M2)
*   **Backend Memory:** **~16 MB RAM** (Handling 20+ concurrent instrument streams).
*   **Backend CPU:** **< 1%** (60 FPS broadcast for 20 symbols).
*   **Frontend Efficiency:** **~0.8% Scripting load** (60 FPS rendering via uPlot + Protobuf).
*   **Network Gain:** **~60% traffic reduction** compared to JSON via Binary Protobuf frames.

## 🏗 System Architecture

### 1. Ingestion & Aggregation (Go Backend)
*   **Concurrency Model:** High-speed pipeline using **Worker Pools** and **Buffered Channels** to prevent backpressure.
*   **Sliding Window Analytics:** Real-time computation of **Moving Average**, **Min/Max (Volatility)**, and **Live Price** over a configurable **300s window**.
*   **Memory Efficiency:** Zero-allocation focus with efficient slice management for trade history, maintaining a stable 16MB footprint.

### 2. Delivery Layer (Binary Protocol)
*   **Protocol Buffers (Protobuf):** Replaced heavy JSON with strictly typed binary payloads. This eliminates `JSON.parse` overhead and drastically reduces GC pressure in the browser.
*   **Hub/Broadcaster:** Thread-safe delivery to multiple WebSocket clients using `sync.RWMutex`, ensuring high-frequency read stability.

### 3. Visualization (Canvas-based Frontend)
*   **uPlot Integration:** Leveraging the world's fastest time-series library for **60 FPS** smooth rendering on HTML5 Canvas.
*   **Volatility Heatmap:** Dynamic UI sorting using **CSS Grid Order**. Most volatile assets (highest % range in 300s) automatically "float" to the top of the dashboard.
*   **Intelligent Scales:** Auto-scaling Y-axis with **Volatility Corridors** (Min/Max bands) to visualize market depth and noise.

## 🛠 Tech Stack
*   **Backend:** Go (Golang), [Gorilla WebSocket](https://github.com), [Google Protobuf](https://github.com).
*   **Frontend:** Vanilla JS (ES6+), [uPlot](https://github.com), [protobuf.js](https://github.com).
*   **Patterns:** Observer, Pipeline, Concurrent Aggregator, Binary Streaming, Worker Pools.

## 🚀 Getting Started

### Backend
```bash
# From project root
go mod download
go run cmd/engine/main.go
