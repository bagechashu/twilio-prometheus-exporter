package main

import (
    "strconv"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/sirupsen/logrus"
)

type TwilioCollector struct {
    client          *TwilioClient
    balanceMetric   *prometheus.GaugeVec
}

func NewTwilioCollector() *TwilioCollector {
    return &TwilioCollector{
        client: NewTwilioClient(),
        balanceMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "twilio_account_balance",
            Help: "Current balance of Twilio account.",
        }, []string{"currency", "account_sid"}),
    }
}

func (tc *TwilioCollector) Describe(ch chan<- *prometheus.Desc) {
    tc.balanceMetric.Describe(ch)
}

func (tc *TwilioCollector) Collect(ch chan<- prometheus.Metric) {
    // Извлечение баланса аккаунта
    balance, currency, accountSID, err := tc.client.FetchBalance()
    if err != nil {
        logrus.WithError(err).Error("Failed to fetch balance")
        return
    }

    // Обновление метрики баланса
    tc.UpdateBalanceMetric(balance, currency, accountSID)

    // Экспортирование метрик в канал
    tc.balanceMetric.Collect(ch)
}

// UpdateBalanceMetric обновляет метрику баланса аккаунта
func (tc *TwilioCollector) UpdateBalanceMetric(balance, currency, accountSID string) {
    balanceFloat, err := strconv.ParseFloat(balance, 64)
    if err != nil {
        logrus.WithError(err).Error("Failed to parse balance")
        return
    }
    tc.balanceMetric.WithLabelValues(currency, accountSID).Set(balanceFloat)
    logrus.WithFields(logrus.Fields{
        "balance":    balance,
        "currency":   currency,
        "accountSID": accountSID,
    }).Info("Twilio account balance")
}