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

	// Register only our Twilio collector in the custom registry
	registry.MustRegister(twilioCollector)

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// Start HTTP server for metrics using our custom registry
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	logrus.Fatal(http.ListenAndServe(":8080", mux))
}
