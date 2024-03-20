package main

import (
	"github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go"

	"time"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioClient struct {
	client *twilio.RestClient
	config Config
}

func NewTwilioClient(accountSID, authToken string, config Config) *TwilioClient {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})
	return &TwilioClient{client: client, config: config} // Save the configuration in the TwilioClient instance
}

// FetchBalance extracts the account balance from Twilio, along with currency and account SID
func (tc *TwilioClient) FetchBalance() ([]openapi.ApiV2010Balance, error) {
	balanceObj, err := tc.client.Api.FetchBalance(nil)
	if err != nil {
		return nil, err
	}
	return []openapi.ApiV2010Balance{*balanceObj}, nil
}

// FetchUsageRecords extracts usage records from Twilio
func (tc *TwilioClient) FetchUsageRecordsToday() ([]openapi.ApiV2010UsageRecordToday, error) {
	// Create parameters for the request including record limit and EndDate
	params := &openapi.ListUsageRecordTodayParams{
		Limit: &tc.config.RecordLimit,
	}

	// Get usage records from Twilio
	usageRecords, err := tc.client.Api.ListUsageRecordToday(params)
	if err != nil {
		return nil, err
	}

	return usageRecords, nil
}

// FetchCall extracts call information from Twilio
func (tc *TwilioClient) FetchCalls() ([]openapi.ApiV2010Call, error) {
	// Convert string to time duration
	offset, err := time.ParseDuration(tc.config.StartDate)
	if err != nil {
		return nil, err
	}

	// Get current time
	currentTime := time.Now().UTC()

	// Calculate start time as current time minus offset
	startTime := currentTime.Add(offset)

	// Convert DURATION to time duration
	duration, err := time.ParseDuration(tc.config.Duration)
	if err != nil {
		return nil, err
	}

	// Calculate endTime as the sum of START_DATE and DURATION
	endTime := startTime.Add(duration)

	// Create parameters for the request
	params := &openapi.ListCallParams{
		Limit:     &tc.config.RecordLimit,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	// Get list of calls from Twilio
	calls, err := tc.client.Api.ListCall(params)
	if err != nil {
		return nil, err
	}

	return calls, nil
}

// FetchMessage extracts message information from Twilio
func (tc *TwilioClient) FetchMessages() ([]openapi.ApiV2010Message, error) {
	// Get start time and duration from configuration
	startDate, err := time.ParseDuration(tc.config.StartDate)
	if err != nil {
		// Handle error parsing start time
		logrus.Errorf("Error parsing start time: %v", err)
		return nil, err
	}

	duration, err := time.ParseDuration(tc.config.Duration)
	if err != nil {
		// Handle error parsing duration
		logrus.Errorf("Error parsing duration: %v", err)
		return nil, err
	}

	// Calculate time interval based on START_DATE and DURATION
	now := time.Now()
	endDate := now.Add(startDate)
	dateSentAfter := endDate.Add(-duration)

	// Create parameters for the request
	params := &openapi.ListMessageParams{
		Limit:          &tc.config.RecordLimit,
		DateSentBefore: &endDate,
		DateSentAfter:  &dateSentAfter,
	}

	// Get list of messages from Twilio
	calls, err := tc.client.Api.ListMessage(params)

	if err != nil {
		return nil, err
	}

	return calls, nil
}
