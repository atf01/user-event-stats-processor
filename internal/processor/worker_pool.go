package processor

import (
	"encoding/json"
	"log"
	"sync"
	"user-event-stats-processor/internal/models"

	"github.com/streadway/amqp"
)

type WorkerPool struct {
	jobQueue chan models.Event
	conn     *amqp.Connection
	channel  *amqp.Channel
	wg       sync.WaitGroup
}

func NewWorkerPool(amqpURL string, bufferSize int) *WorkerPool {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	_, err = ch.QueueDeclare(
		"events_queue", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	return &WorkerPool{
		jobQueue: make(chan models.Event, bufferSize),
		conn:     conn,
		channel:  ch,
	}
}

func (workers *WorkerPool) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		workers.wg.Add(1)
		go func() {
			defer workers.wg.Done()
			for event := range workers.jobQueue {
				body, _ := json.Marshal(event)
				err := workers.channel.Publish(
					"",
					"events_queue",
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					})
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
				}
			}
		}()
	}
}

func (wp *WorkerPool) Enqueue(event models.Event) bool {
	select {
	case wp.jobQueue <- event:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.jobQueue)
	wp.wg.Wait()
	wp.channel.Close()
	wp.conn.Close()
}
