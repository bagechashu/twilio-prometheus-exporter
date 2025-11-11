package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Config struct {
	TwilioAccountSID string
	TwilioAuthToken  string
	LogLevel         string
	RecordLimit      int
	StartDate        string
	Duration         string
	SkipMissing      bool
}

func initConfig() Config {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	recordLimitStr := os.Getenv("RECORD_LIMIT")
	recordLimit, err := strconv.Atoi(recordLimitStr)
	if err != nil {
		recordLimit = 200
	}
	startDate := os.Getenv("START_DATE")
	if startDate == "" {
		startDate = "-10m"
	}
	duration := os.Getenv("DURATION")
	if duration == "" {
		duration = "5m"
	}
	skipMissing, err := strconv.ParseBool(os.Getenv("SKIP_MISSING"))
	if err != nil {
		skipMissing = false
	}
	return Config{
		TwilioAccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		LogLevel:         os.Getenv("LOG_LEVEL"),
		RecordLimit:      recordLimit,
		StartDate:        startDate,
		Duration:         duration,
		SkipMissing:      skipMissing,
	}
}

func main() {
	config := initConfig()

	// Set log format to JSON
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Set log level
	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logrus.Warnf("Invalid LOG_LEVEL value '%s', using default level 'info'", config.LogLevel)
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	// Create a custom registry without default collectors
	registry := prometheus.NewRegistry()

	// Create a new Twilio collector using the configuration
	twilioCollector := NewTwilioCollector(config)

	// Register Twilio collector in the custom registry
	registry.MustRegister(twilioCollector)

	// Create webhook metrics for real-time event tracking
	webhookMetrics := NewWebhookMetrics()
	if err := webhookMetrics.Register(registry); err != nil {
		logrus.WithError(err).Fatal("Failed to register webhook metrics")
	}

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Webhook endpoints for real-time event capture
	mux.HandleFunc("/webhooks/message", webhookMetrics.HandleMessageStatusCallback)
	mux.HandleFunc("/webhooks/call", webhookMetrics.HandleCallStatusCallback)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	logrus.Info("Starting Twilio Prometheus exporter on :8080")
	logrus.Info("  /metrics - Prometheus metrics endpoint")
	logrus.Info("  /webhooks/message - Twilio message status callbacks")
	logrus.Info("  /webhooks/call - Twilio call status callbacks")
	logrus.Info("  /health - Health check endpoint")

	logrus.Fatal(http.ListenAndServe(":8080", mux))
}
