package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// WebhookMetrics holds Prometheus counters for real-time Twilio events
type WebhookMetrics struct {
	messagesDeliveredTotal *prometheus.CounterVec // labels: none (simple counter)
	messagesFailedTotal    *prometheus.CounterVec // labels: error_code
	callsCompletedTotal    *prometheus.CounterVec // labels: call_status
	callsFailedTotal       *prometheus.CounterVec // labels: disconnect_reason
}

// NewWebhookMetrics creates and registers webhook event counters
func NewWebhookMetrics() *WebhookMetrics {
	return &WebhookMetrics{
		messagesDeliveredTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "twilio_messages_delivered_total",
			Help: "Total number of successfully delivered messages via webhook callbacks.",
		}, []string{}),
		messagesFailedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "twilio_messages_failed_total",
			Help: "Total number of failed/undelivered messages via webhook callbacks.",
		}, []string{"error_code"}),
		callsCompletedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "twilio_calls_completed_total",
			Help: "Total number of completed calls via webhook callbacks.",
		}, []string{"call_status"}),
		callsFailedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "twilio_calls_failed_total",
			Help: "Total number of failed calls via webhook callbacks.",
		}, []string{"disconnect_reason"}),
	}
}

// MessageStatusCallback represents a Twilio message status callback payload
type MessageStatusCallback struct {
	MessageSid    string
	AccountSid    string
	From          string
	To            string
	MessageStatus string
	ErrorCode     string
}

// CallStatusCallback represents a Twilio call status callback payload
type CallStatusCallback struct {
	CallSid          string
	AccountSid       string
	From             string
	To               string
	CallStatus       string
	CallDuration     string
	Timestamp        string
	RecordingUrl     string
	Reason           string
	DisconnectReason string
}

// ValidateTwilioWebhookSignature validates the request came from Twilio
// See: https://www.twilio.com/docs/usage/webhooks/webhooks-security
func ValidateTwilioWebhookSignature(r *http.Request, authToken string) bool {
	// Get the signature from the X-Twilio-Signature header
	signature := r.Header.Get("X-Twilio-Signature")
	if signature == "" {
		logrus.Warn("Missing X-Twilio-Signature header in webhook request")
		return false
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read webhook request body")
		return false
	}
	defer r.Body.Close()

	// Restore the body for further processing
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Reconstruct the full URL
	fullURL := "https://" + r.Host + r.RequestURI

	// Parse the request body as form data
	formData, err := url.ParseQuery(string(body))
	if err != nil {
		logrus.WithError(err).Error("Failed to parse webhook form data")
		return false
	}

	// Sort form keys for consistent ordering
	var keys []string
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Reconstruct the data string from sorted keys
	var dataStr strings.Builder
	dataStr.WriteString(fullURL)
	for _, k := range keys {
		dataStr.WriteString(k)
		for _, v := range formData[k] {
			dataStr.WriteString(v)
		}
	}

	// Compute HMAC-SHA1
	h := hmac.New(sha1.New, []byte(authToken))
	h.Write([]byte(dataStr.String()))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Compare signatures in constant time
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		logrus.WithFields(logrus.Fields{
			"received": signature,
			"expected": expectedSignature,
		}).Warn("Webhook signature validation failed")
		return false
	}

	return true
}

// HandleMessageStatusCallback processes Twilio message status callback
func (wm *WebhookMetrics) HandleMessageStatusCallback(w http.ResponseWriter, r *http.Request) {
	// Validate webhook signature
	if !ValidateTwilioWebhookSignature(r, "") { // Note: authToken should come from config
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		logrus.WithError(err).Error("Failed to parse message callback form")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Extract callback data
	callback := MessageStatusCallback{
		MessageSid:    r.FormValue("MessageSid"),
		AccountSid:    r.FormValue("AccountSid"),
		From:          r.FormValue("From"),
		To:            r.FormValue("To"),
		MessageStatus: r.FormValue("MessageStatus"),
		ErrorCode:     r.FormValue("ErrorCode"),
	}

	// Log the callback
	logrus.WithFields(logrus.Fields{
		"message_sid": callback.MessageSid,
		"status":      callback.MessageStatus,
		"error_code":  callback.ErrorCode,
	}).Debug("Message status callback received")

	// Update metrics based on status
	switch callback.MessageStatus {
	case "delivered":
		wm.messagesDeliveredTotal.WithLabelValues().Inc()
	case "failed", "undelivered":
		errorCode := callback.ErrorCode
		if errorCode == "" {
			errorCode = "unknown"
		}
		wm.messagesFailedTotal.WithLabelValues(errorCode).Inc()
	}

	// Return 200 OK to acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// HandleCallStatusCallback processes Twilio call status callback
func (wm *WebhookMetrics) HandleCallStatusCallback(w http.ResponseWriter, r *http.Request) {
	// Validate webhook signature
	if !ValidateTwilioWebhookSignature(r, "") { // Note: authToken should come from config
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		logrus.WithError(err).Error("Failed to parse call callback form")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Extract callback data
	callback := CallStatusCallback{
		CallSid:          r.FormValue("CallSid"),
		AccountSid:       r.FormValue("AccountSid"),
		From:             r.FormValue("From"),
		To:               r.FormValue("To"),
		CallStatus:       r.FormValue("CallStatus"),
		CallDuration:     r.FormValue("CallDuration"),
		Timestamp:        r.FormValue("Timestamp"),
		RecordingUrl:     r.FormValue("RecordingUrl"),
		Reason:           r.FormValue("Reason"),
		DisconnectReason: r.FormValue("CallbackSource"),
	}

	// Log the callback
	logrus.WithFields(logrus.Fields{
		"call_sid":   callback.CallSid,
		"status":     callback.CallStatus,
		"duration":   callback.CallDuration,
		"disconnect": callback.DisconnectReason,
	}).Debug("Call status callback received")

	// Update metrics based on status
	switch callback.CallStatus {
	case "completed":
		wm.callsCompletedTotal.WithLabelValues(callback.CallStatus).Inc()
	case "failed", "no-answer", "busy":
		wm.callsFailedTotal.WithLabelValues(callback.DisconnectReason).Inc()
	}

	// Return 200 OK to acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// RegisterWebhookMetrics registers webhook metrics with the given registry
func (wm *WebhookMetrics) Register(registry *prometheus.Registry) error {
	if err := registry.Register(wm.messagesDeliveredTotal); err != nil {
		return err
	}
	if err := registry.Register(wm.messagesFailedTotal); err != nil {
		return err
	}
	if err := registry.Register(wm.callsCompletedTotal); err != nil {
		return err
	}
	if err := registry.Register(wm.callsFailedTotal); err != nil {
		return err
	}
	return nil
}

// CollectWebhookMetrics returns a function that collects webhook metrics
func (wm *WebhookMetrics) Collect(ch chan<- prometheus.Metric) {
	wm.messagesDeliveredTotal.Collect(ch)
	wm.messagesFailedTotal.Collect(ch)
	wm.callsCompletedTotal.Collect(ch)
	wm.callsFailedTotal.Collect(ch)
}
