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
  ```
  and refer to placeholders and add the accordingly in your file:
  ```bash
  RABBITMQ_USER={your-user-admin}
  RABBITMQ_PASS={your-user-pass}
  RABBITMQ_HOST={localhost}:{port}
  SCYLLA_USER={db-username}
  SCYLLA_PASS={db-password}
  ```


### 2. Infrastructure Setup
Spin up the core services in the background:
 ```bash
 docker-compose up -d
```


### 3. run the scyllaDB configuration script
run the script under /scripts of name init-scylla.sh:
```bash
docker exec -i scylla-db sh < ./scripts/init-scylla.sh
```
this step ensures that the keyspace & needed tables are created

### 4. now run both /cmd/consumer/main.go & /cmd/server/main.go
```bash
go run /cmd/consumer/main.go
go run /cmd/server/main.go
```

### 5. test against data race conditions
```bash
go run -race cmd/consumer/main.go
go run -race cmd/server/main.go
```

### 6. a small test to make sure everything is up and running
Once both the Producer and Consumer are running, use these commands to simulate real-world user activity.

1. Track a Page View (Simple Increment)
Simulate a user clicking on a product or viewing a page.
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_101",
    "event_type": "page_view",
    "value": 1
  }'
```

2. Track a Purchase (Fixed-Point Arithmetic)
Simulate a transaction. Note that the system will store this as 2999 in ScyllaDB to maintain precision.
```bash
curl -X POST http://localhost:8080/events -d '{"user_id": "ahmed_123", "event_type": "click", "value": 1}'
```

3. High-Concurrency Stress Test (Bash Loop)
To see the 50-worker pool in action, run this quick loop to fire 100 events in rapid succession:

```bash

for i in {1..100}; do
  curl -s -X POST http://localhost:8080/events \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": \"user_$i\", \"event_type\": \"stress_test\", \"value\": $i}" > /dev/null &
done
```
4. test get user stats:
```bash
curl "http://localhost:8080/stats?user_id=ahmed_123"
``` 
## ⚖️ Tradeoffs & Consistency Model

### Eventual Consistency over Strict Consistency
In this system, I have intentionally favored **Availability** and **Partition Tolerance** (the 'A' and 'P' in CAP) over strict, immediate Consistency. 

**The Rationale:**
- **Dashboard Use-Case:** This service is designed to power analytical dashboards. In such systems, a delay of a few milliseconds (or even seconds) between a `POST` (event ingestion) and a `GET` (stats retrieval) is acceptable. 
- **User Experience:** It is more critical that the Ingestor (Producer) is always available to accept user events than it is for the Dashboard to reflect that specific event the microsecond it occurs.
- **High Throughput:** By choosing **Eventual Consistency**, we eliminate the need for distributed locking or synchronous "Wait-for-All" replication, allowing the system to handle thousands of concurrent writes without bottlenecking.
---
