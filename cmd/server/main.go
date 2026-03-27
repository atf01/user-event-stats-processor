package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-event-stats-processor/internal/api"
	"user-event-stats-processor/internal/config"
	"user-event-stats-processor/internal/processor"
	"user-event-stats-processor/internal/store"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println(cfg.RabbitMQ.URL)
	sStore, err := store.NewScyllaStore(cfg.Scylla)
	if err != nil {
		log.Fatalf("Failed to connect to ScyllaDB: %v", err)
	}

	wp := processor.NewWorkerPool(cfg.RabbitMQ.URL, cfg.WorkerPool.BufferSize)
	wp.Start(cfg.WorkerPool.Workers)

	h := api.NewHandler(sStore, wp)
	mux := http.NewServeMux()
	mux.HandleFunc("/events", h.PostEvent)
	mux.HandleFunc("/stats", h.GetStats)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.App.Port), Handler: mux}

	go func() {
		log.Printf("🚀 API Server on :%d", cfg.App.Port)
		srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	wp.Stop()
}
