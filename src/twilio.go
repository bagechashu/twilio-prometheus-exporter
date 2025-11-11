package main

import (
	"github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go"

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

// NOTE: Call and Message events are now handled via HTTP webhooks.
// For real-time call and message monitoring, configure Twilio to send Status Callbacks to:
// - Message Status Callbacks: POST /webhooks/message
// - Call Status Events: POST /webhooks/call
// This eliminates the high-cardinality risk of polling individual call/message records.
