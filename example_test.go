package config_test

import (
	"fmt"
	"github.com/Yalantis/go-config"
	"log"
	"time"
)

func init() {
	// setup ENV
	_ = config.Setenv("VERSION", "0.0.1")
	_ = config.Setenv("REPLICAS_0_POSTGRES_USER", "replica0")
	_ = config.Setenv("REPLICAS_0_POSTGRES_PORT", "5433")
	_ = config.Setenv("REPLICAS_1_USER", "replica1")
	_ = config.Setenv("REPLICAS_1_PORT", "5433")
	_ = config.Setenv("REDIS_ADDR", "127.0.0.1:6377,127.0.0.1:6378,127.0.0.1:6379")
	_ = config.Setenv("READ_TIMEOUT", "30s")
}

func ExampleInit() {
	type Server struct {
		Addr string `json:"addr" envconfig:"SERVER_ADDR" default:"localhost:8080"`
	}

	type Postgres struct {
		Host     string `json:"host"     envconfig:"POSTGRES_HOST"     default:"localhost"`
		Port     string `json:"port"     envconfig:"POSTGRES_PORT"     default:"5432"`
		User     string `json:"user"     envconfig:"POSTGRES_USER"     default:"postgres"`
		Password string `json:"password" envconfig:"POSTGRES_PASSWORD" default:"12345"`
	}

	type Redis struct {
		Addrs []string `json:"addrs" envconfig:"REDIS_ADDR" default:"localhost:6379"`
	}

	type NATS struct {
		ServerURL               string          `json:"server_url"                envconfig:"NATS_SERVER_URL"             default:"nats://localhost:4222"`
		MaxReconnectionAttempts int             `json:"max_reconnection_attempts" envconfig:"NATS_MAX_RECONNECT_ATTEMPTS" default:"5"`
		ReconnectInterval       config.Duration `json:"reconnect_interval"        envconfig:"NATS_RECONNECT_INTERVAL"     default:"2s"`
	}

	type Websocket struct {
		Port string `json:"port" envconfig:"WEBSOCKET_PORT" default:"9876"`
	}

	var cfg struct {
		Version   string     `envconfig:"VERSION" default:"0"`
		Server    Server     `json:"server"`
		Postgres  Postgres   `json:"postgres"`
		Replicas  []Postgres `json:"replicas" envprefix:"REPLICAS"`
		Redis     Redis      `json:"redis"`
		NATS      NATS       `json:"nats"`
		Websocket Websocket  `json:"websocket"`
	}

	if err := config.Init(&cfg, "testdata/config.json"); err != nil {
		log.Fatalln(err)
	}

	fmt.Println(cfg)
	// Output: {0.0.1 {localhost:8080} {localhost 5432 postgres 12345} [{localhost 5433 replica0 12345} {localhost 5433 replica1 12345}] {[127.0.0.1:6377 127.0.0.1:6378 127.0.0.1:6379]} {nats://localhost:4222 5 2000000000} {9876}}
}

func ExampleInitTimeout() {
	var cfg struct {
		ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT"  default:"1s"`
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	}

	if err := config.Init(&cfg, ""); err != nil {
		log.Fatalln(err)
	}

	fmt.Println(cfg)
	// Output: {30s 10s}
}
