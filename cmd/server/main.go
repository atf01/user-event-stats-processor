package main

import (
	"encoding/json"
	"log"
	"user-event-stats-processor/internal/config"
	"user-event-stats-processor/internal/models"
	"user-event-stats-processor/internal/store"

	"github.com/streadway/amqp"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize Scylla with error checking
	sStore, err := store.NewScyllaStore(cfg.Scylla)
	if err != nil {
		log.Fatalf("🛑 Failed to initialize ScyllaDB: %v", err)
	}
	defer sStore.Close()

	// 2. Connect to RabbitMQ with error checking
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("🛑 Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 3. Open Channel with error checking
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("🛑 Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 4. Start Consuming
	msgs, err := ch.Consume(
		cfg.RabbitMQ.QueueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("🛑 Failed to register a consumer: %v", err)
	}

	log.Println("📥 Consumer is live. Listening for events...")

	// 5. Process loop
	for d := range msgs {
		var e models.Event
		if err := json.Unmarshal(d.Body, &e); err != nil {
			log.Printf("⚠️ Error decoding event: %v", err)
			continue
		}

		// Atomic update in Scylla
		if err := sStore.Update(e.UserID, e.EventType, e.Value); err != nil {
			log.Printf("❌ Failed to update Scylla for user %s: %v", e.UserID, err)
		} else {
			log.Printf("✅ Processed %s for user %s", e.EventType, e.UserID)
		}
	}
}
