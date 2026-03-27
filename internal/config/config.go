package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Scylla     ScyllaConfig     `mapstructure:"scylla"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	WorkerPool WorkerPoolConfig `mapstructure:"worker_pool"`
}

type AppConfig struct {
	Port int `mapstructure:"port"`
}

type ScyllaConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	User     string   `mapstructure:"user"`
	Pass     string   `mapstructure:"pass"`
	Keyspace string   `mapstructure:"keyspace"`
}

type RabbitMQConfig struct {
	URL       string `mapstructure:"url"`
	User      string `mapstructure:"user"`
	Pass      string `mapstructure:"pass"`
	QueueName string `mapstructure:"queue_name"`
}

type WorkerPoolConfig struct {
	Workers    int `mapstructure:"workers"`
	BufferSize int `mapstructure:"buffer_size"`
}

func LoadConfig() *Config {
	// 1. Force load the .env file into the process environment
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found; falling back to system environment variables")
	}
	v := viper.New()
	v.SetConfigFile("config.yaml")
	v.AutomaticEnv()

	// 1. Bind the ENV vars
	v.BindEnv("rabbitmq.user", "RABBITMQ_USER")
	v.BindEnv("rabbitmq.pass", "RABBITMQ_PASS")

	if err := v.ReadInConfig(); err != nil {
		log.Printf("No config file found, using ENV")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unmarshal error: %v", err)
	}

	// 2. CRITICAL FIX: Ensure User/Pass aren't empty before building
	user := cfg.RabbitMQ.User
	pass := cfg.RabbitMQ.Pass

	// If ENV didn't provide them, fallback to guest (default)
	if user == "" {
		user = "guest"
	}
	if pass == "" {
		pass = "guest"
	}

	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "127.0.0.1:5672"
	}

	// Rebuild the URL correctly
	cfg.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s/", user, pass, host)

	return &cfg
}
