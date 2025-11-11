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
func (tc *TwilioClient) FetchBalance() ([]ApiV2010Balance, error) {
	balanceObj, err := tc.client.Api.FetchBalance(nil)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("FetchBalance: %+v", balanceObj)

	// Convert from openapi.ApiV2010Balance to ApiV2010Balance
	localBalance := ApiV2010Balance{
		AccountSid: balanceObj.AccountSid,
		Balance:    balanceObj.Balance,
		Currency:   balanceObj.Currency,
	}

	return []ApiV2010Balance{localBalance}, nil
}

// FetchUsageRecords extracts usage records from Twilio
func (tc *TwilioClient) FetchUsageRecordsToday() ([]ApiV2010UsageRecordToday, error) {
	// Create parameters for the request including record limit and EndDate
	params := &openapi.ListUsageRecordTodayParams{
		Limit: &tc.config.RecordLimit,
	}

	// Get usage records from Twilio
	usageRecords, err := tc.client.Api.ListUsageRecordToday(params)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("FetchUsageRecordsToday: %+v", usageRecords)

	// Convert from openapi.ApiV2010UsageRecordToday to ApiV2010UsageRecordToday
	localUsageRecords := make([]ApiV2010UsageRecordToday, len(usageRecords))
	for i, record := range usageRecords {
		localUsageRecords[i] = ApiV2010UsageRecordToday{
			AccountSid: record.AccountSid,
			ApiVersion: record.ApiVersion,
			Category:   record.Category,
			CountUnit:  record.CountUnit,
			Usage:      record.Usage,
			UsageUnit:  record.UsageUnit,
		}
	}

	return localUsageRecords, nil
}

// FetchCall extracts call information from Twilio
func (tc *TwilioClient) FetchCalls() ([]ApiV2010Call, error) {
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
	logrus.Debugf("FetchCalls: %+v", calls)

	// Convert from openapi.ApiV2010Call to ApiV2010Call
	localCalls := make([]ApiV2010Call, len(calls))
	for i, call := range calls {
		localCalls[i] = ApiV2010Call{
			AccountSid: call.AccountSid,
			To:         call.To,
			Status:     call.Status,
			ApiVersion: call.ApiVersion,
		}
	}

	return localCalls, nil
}

// FetchMessage extracts message information from Twilio
func (tc *TwilioClient) FetchMessages() ([]ApiV2010Message, error) {
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
	messages, err := tc.client.Api.ListMessage(params)

	if err != nil {
		return nil, err
	}
	logrus.Debugf("FetchMessages: %+v", messages)

	// Convert from openapi.ApiV2010Message to ApiV2010Message
	localMessages := make([]ApiV2010Message, len(messages))
	for i, message := range messages {
		localMessages[i] = ApiV2010Message{
			To:         message.To,
			AccountSid: message.AccountSid,
			Status:     message.Status,
			ErrorCode:  message.ErrorCode,
			ApiVersion: message.ApiVersion,
		}
	}

	return localMessages, nil
}
