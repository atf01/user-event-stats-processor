# User Event Stats Processor

A high-concurrency, asynchronous event processing engine built in **Go**. This service is designed to ingest high-frequency user activity (clicks, views, purchases) via **RabbitMQ** and aggregate statistics in **ScyllaDB** using atomic operations and fixed-point arithmetic.

---

## 🏗 Architectural Rationale

### 1. Why Asynchronous?
To handle high-burst traffic (e.g., flash sales or global events), a synchronous API would block the caller and risk timeouts. By using an async pattern, we decouple **Ingestion** from **Processing**. The system acknowledges receipt immediately, ensuring a high-throughput user experience regardless of database pressure.

### 2. Why RabbitMQ?
RabbitMQ acts as a reliable buffer and load-leveler:
- **Backpressure Management:** It protects ScyllaDB from write-spikes by queueing messages, allowing our worker pool to consume them at a steady, sustainable rate.
- **Durability:** Messages are persisted, ensuring no user events are lost if a consumer node restarts.
- **Decoupling:** The Producer (Ingestor) remains independent of the Consumer (Worker), allowing for independent scaling and deployment.

### 3. Why ScyllaDB?
ScyllaDB was chosen for its masterless, distributed architecture which excels at write-heavy workloads:
- **Native Counters:** We utilize ScyllaDB `counter` types for distributed, atomic increments. This eliminates the need for expensive distributed locks (like Redis Redlock) or "Read-Modify-Write" cycles.
- **High Availability:** Its "Shared-Nothing" architecture ensures no single point of failure for our statistics engine.

### 4. Fixed-Point Arithmetic Strategy
ScyllaDB counters only support integers. To support decimal values (like `$15.50` for purchases) without losing precision:
- **Application Layer:** Scales values by a factor of 100 (e.g., `15.50 -> 1550`).
- **Database Layer:** Increments the integer counter.
- **Retrieval:** The value is divided by 100 to restore the original decimal precision.

---

## 🛠 Tech Stack
- **Language:** Go 1.21+
- **Message Broker:** RabbitMQ (AMQP 0.9.1)
- **Database:** ScyllaDB (Cassandra-compatible NoSQL)
- **Concurrency:** Worker Pool Pattern (50 Concurrent Goroutines)
- **Configuration:** Viper / Godotenv

---

## 📦 Getting Started

### 1. Prerequisites
- Docker & Docker Compose
- Go 1.21 or higher
- create your own .env file to configure your secrets properly
  ```bash
  cp .env.example .env

### 2. Infrastructure Setup
Spin up the core services in the background:
 ```bash
 docker-compose up -d

### 3. run the scyllaDB configuration script
run the script under /scripts of name init-scylla.sh:
```bash
docker exec -i scylla-db sh < ./scripts/init-scylla.sh
