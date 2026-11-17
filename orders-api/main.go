package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(context.Background(), config.postgresURL())
	if err != nil {
		return fmt.Errorf("configure postgres pool: %w", err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.redisAddress(),
		Password: config.redisPassword,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("close redis client: %v", err)
		}
	}()

	api := NewAPI(NewPostgresStore(pool), NewRedisQueue(redisClient), config.version, config.pod)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	signals, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve orders API: %w", err)
		}
	case <-signals.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("gracefully shut down orders API: %w", err)
		}
		if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve orders API: %w", err)
		}
	}
	return nil
}

type config struct {
	postgresHost     string
	postgresPort     int
	postgresDB       string
	postgresUser     string
	postgresPassword string
	redisHost        string
	redisPort        int
	redisPassword    string
	version          string
	pod              string
}

func loadConfig() (config, error) {
	var cfg config
	var err error
	if cfg.postgresHost, err = requiredEnv("POSTGRES_HOST"); err != nil {
		return config{}, err
	}
	if cfg.postgresPort, err = requiredPort("POSTGRES_PORT"); err != nil {
		return config{}, err
	}
	if cfg.postgresDB, err = requiredEnv("POSTGRES_DB"); err != nil {
		return config{}, err
	}
	if cfg.postgresUser, err = requiredEnv("POSTGRES_USER"); err != nil {
		return config{}, err
	}
	if cfg.postgresPassword, err = requiredEnv("POSTGRES_PASSWORD"); err != nil {
		return config{}, err
	}
	if cfg.redisHost, err = requiredEnv("REDIS_HOST"); err != nil {
		return config{}, err
	}
	if cfg.redisPort, err = requiredPort("REDIS_PORT"); err != nil {
		return config{}, err
	}
	if cfg.pod, err = requiredEnv("HOSTNAME"); err != nil {
		return config{}, err
	}
	cfg.redisPassword = os.Getenv("REDIS_PASSWORD")
	cfg.version = os.Getenv("APP_VERSION")
	if cfg.version == "" {
		cfg.version = "dev"
	}
	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s must be set", key)
	}
	return value, nil
}

func requiredPort(key string) (int, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be a valid TCP port", key)
	}
	return port, nil
}

func (cfg config) postgresURL() string {
	return (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.postgresUser, cfg.postgresPassword),
		Host:   net.JoinHostPort(cfg.postgresHost, strconv.Itoa(cfg.postgresPort)),
		Path:   cfg.postgresDB,
	}).String()
}

func (cfg config) redisAddress() string {
	return net.JoinHostPort(cfg.redisHost, strconv.Itoa(cfg.redisPort))
}
