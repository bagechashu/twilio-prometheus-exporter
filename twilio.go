package main

import (
	"os"
    "github.com/sirupsen/logrus"
    "github.com/twilio/twilio-go"
)

type TwilioClient struct {
    client *twilio.RestClient
}

func NewTwilioClient() *TwilioClient {
    accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")
    client := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: accountSid,
        Password: authToken,
    })
    return &TwilioClient{client: client}
}

// FetchBalance извлекает баланс аккаунта из Twilio, а также возвращает валюту и account SID
func (tc *TwilioClient) FetchBalance() (balance, currency, accountSID string, err error) {
    balanceObj, err := tc.client.Api.FetchBalance(nil)
    if err != nil {
        logrus.WithError(err).Error("Failed to fetch balance")
        return "", "", "", err
    }
    return *balanceObj.Balance, *balanceObj.Currency, *balanceObj.AccountSid, nil
}
