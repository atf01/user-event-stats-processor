# User Event Stats Processor

A high-performance, asynchronous event processing engine built in **Go**. This service ingests high-frequency user activity from **RabbitMQ** and aggregates statistics in **ScyllaDB** using atomic operations.

## 🚀 Architectural Rationale

### Why Asynchronous?
To handle high-burst traffic (e.g., flash sales or viral events), a synchronous API would block the caller and risk timeouts. By using an async pattern, we decouple ingestion from processing. The system acknowledges receipt immediately, ensuring a smooth user experience regardless of database load.

### Why RabbitMQ?
RabbitMQ acts as a reliable buffer. 
- **Load Leveling:** It protects ScyllaDB from spikes by queueing messages and allowing our 50 concurrent workers to consume them at a steady, manageable rate.
- **Durability:** Messages are persisted, ensuring that no user events are lost even if the consumer service restarts.

### Why ScyllaDB?
ScyllaDB was chosen for its masterless, distributed architecture which excels at write-heavy workloads.
- **Native Counters:** We utilize ScyllaDB `counter` types for distributed, atomic increments. This eliminates the need for complex distributed locks (like Redis Redlock) or "Read-Modify-Write" cycles.
- **Fixed-Point Arithmetic:** To support decimal values (like `float64` for purchases) within integer-only counters, the system scales values by a factor of 100 at the application layer. This preserves precision while maintaining the performance of native CQL counters.


---

## 🛠 Tech Stack
- **Language:** Go 1.21+
- **Message Broker:** RabbitMQ (AMQP 0.9.1)
- **Database:** ScyllaDB (Cassandra-compatible)
- **Configuration:** Viper / Godotenv

---

## 📦 How to Run

### 1. Prerequisites
Ensure you have **Docker** and **Go** installed.

### 2. Infrastructure Setup
Spin up the required services using Docker Compose:
```bash
docker-compose up -d
