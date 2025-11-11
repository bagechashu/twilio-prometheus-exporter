package main

type ApiV2010Balance struct {
	// The unique SID identifier of the Account.
	AccountSid *string `json:"account_sid,omitempty"`
	// The balance of the Account, in units specified by the unit parameter. Balance changes may not be reflected immediately. Child accounts do not contain balance information
	Balance *string `json:"balance,omitempty"`
	// The units of currency for the account balance
	Currency *string `json:"currency,omitempty"`
}

type ApiV2010UsageRecordToday struct {
	// The SID of the [Account](https://www.twilio.com/docs/iam/api/account) that accrued the usage.
	AccountSid *string `json:"account_sid,omitempty"`
	// The API version used to create the resource.
	ApiVersion *string `json:"api_version,omitempty"`
	// // Usage records up to date as of this timestamp, formatted as YYYY-MM-DDTHH:MM:SS+00:00. All timestamps are in GMT
	// AsOf     *string `json:"as_of,omitempty"`
	Category *string `json:"category,omitempty"`
	// The number of usage events, such as the number of calls.
	// Count *string `json:"count,omitempty"`
	// // The units in which `count` is measured, such as `calls` for calls or `messages` for SMS.
	CountUnit *string `json:"count_unit,omitempty"`
	// // A plain-language description of the usage category.
	// Description *string `json:"description,omitempty"`
	// The last date for which usage is included in the UsageRecord. The date is specified in GMT and formatted as `YYYY-MM-DD`.
	// EndDate *string `json:"end_date,omitempty"`
	// // The total price of the usage in the currency specified in `price_unit` and associated with the account.
	// Price *float32 `json:"price,omitempty"`
	// // The currency in which `price` is measured, in [ISO 4127](https://www.iso.org/iso/home/standards/currency_codes.htm) format, such as `usd`, `eur`, and `jpy`.
	// PriceUnit *string `json:"price_unit,omitempty"`
	// // The first date for which usage is included in this UsageRecord. The date is specified in GMT and formatted as `YYYY-MM-DD`.
	// StartDate *string `json:"start_date,omitempty"`
	// // A list of related resources identified by their URIs. For more information, see [List Subresources](https://www.twilio.com/docs/usage/api/usage-record#list-subresources).
	// SubresourceUris *map[string]interface{} `json:"subresource_uris,omitempty"`
	// // The URI of the resource, relative to `https://api.twilio.com`.
	// Uri *string `json:"uri,omitempty"`
	// The amount used to bill usage and measured in units described in `usage_unit`.
	Usage *string `json:"usage,omitempty"`
	// The units in which `usage` is measured, such as `minutes` for calls or `messages` for SMS.
	UsageUnit *string `json:"usage_unit,omitempty"`
}

type ApiV2010Call struct {
	// // The unique string that we created to identify this Call resource.
	// Sid *string `json:"sid,omitempty"`
	// // The date and time in UTC that this resource was created specified in [RFC 2822](https://www.ietf.org/rfc/rfc2822.txt) format.
	// DateCreated *string `json:"date_created,omitempty"`
	// // The date and time in UTC that this resource was last updated, specified in [RFC 2822](https://www.ietf.org/rfc/rfc2822.txt) format.
	// DateUpdated *string `json:"date_updated,omitempty"`
	// // The SID that identifies the call that created this leg.
	// ParentCallSid *string `json:"parent_call_sid,omitempty"`
	// // The SID of the [Account](https://www.twilio.com/docs/iam/api/account) that created this Call resource.
	AccountSid *string `json:"account_sid,omitempty"`
	// The phone number, SIP address, Client identifier or SIM SID that received this call. Phone numbers are in [E.164](https://www.twilio.com/docs/glossary/what-e164) format (e.g., +16175551212). SIP addresses are formatted as `name@company.com`. Client identifiers are formatted `client:name`. SIM SIDs are formatted as `sim:sid`.
	To *string `json:"to,omitempty"`
	// // The phone number, SIP address or Client identifier that received this call. Formatted for display. Non-North American phone numbers are in [E.164](https://www.twilio.com/docs/glossary/what-e164) format (e.g., +442071838750).
	// ToFormatted *string `json:"to_formatted,omitempty"`
	// // The phone number, SIP address, Client identifier or SIM SID that made this call. Phone numbers are in [E.164](https://www.twilio.com/docs/glossary/what-e164) format (e.g., +16175551212). SIP addresses are formatted as `name@company.com`. Client identifiers are formatted `client:name`. SIM SIDs are formatted as `sim:sid`.
	// From *string `json:"from,omitempty"`
	// // The calling phone number, SIP address, or Client identifier formatted for display. Non-North American phone numbers are in [E.164](https://www.twilio.com/docs/glossary/what-e164) format (e.g., +442071838750).
	// FromFormatted *string `json:"from_formatted,omitempty"`
	// // If the call was inbound, this is the SID of the IncomingPhoneNumber resource that received the call. If the call was outbound, it is the SID of the OutgoingCallerId resource from which the call was placed.
	// PhoneNumberSid *string `json:"phone_number_sid,omitempty"`
	Status *string `json:"status,omitempty"`
	// // The start time of the call, given as UTC in [RFC 2822](https://www.php.net/manual/en/class.datetime.php#datetime.constants.rfc2822) format. Empty if the call has not yet been dialed.
	// StartTime *string `json:"start_time,omitempty"`
	// // The time the call ended, given as UTC in [RFC 2822](https://www.php.net/manual/en/class.datetime.php#datetime.constants.rfc2822) format. Empty if the call did not complete successfully.
	// EndTime *string `json:"end_time,omitempty"`
	// // The length of the call in seconds. This value is empty for busy, failed, unanswered, or ongoing calls.
	// Duration *string `json:"duration,omitempty"`
	// // The charge for this call, in the currency associated with the account. Populated after the call is completed. May not be immediately available. The price associated with a call only reflects the charge for connectivity.  Charges for other call-related features such as Answering Machine Detection, Text-To-Speech, and SIP REFER are not included in this value.
	// Price *string `json:"price,omitempty"`
	// // The currency in which `Price` is measured, in [ISO 4127](https://www.iso.org/iso/home/standards/currency_codes.htm) format (e.g., `USD`, `EUR`, `JPY`). Always capitalized for calls.
	// PriceUnit *string `json:"price_unit,omitempty"`
	// // A string describing the direction of the call. Can be: `inbound` for inbound calls, `outbound-api` for calls initiated via the REST API or `outbound-dial` for calls initiated by a `<Dial>` verb. Using [Elastic SIP Trunking](https://www.twilio.com/docs/sip-trunking), the values can be [`trunking-terminating`](https://www.twilio.com/docs/sip-trunking#termination) for outgoing calls from your communications infrastructure to the PSTN or [`trunking-originating`](https://www.twilio.com/docs/sip-trunking#origination) for incoming calls to your communications infrastructure from the PSTN.
	// Direction *string `json:"direction,omitempty"`
	// // Either `human` or `machine` if this call was initiated with answering machine detection. Empty otherwise.
	// AnsweredBy *string `json:"answered_by,omitempty"`
	// The API version used to create the call.
	ApiVersion *string `json:"api_version,omitempty"`
	// // The forwarding phone number if this call was an incoming call forwarded from another number (depends on carrier supporting forwarding). Otherwise, empty.
	// ForwardedFrom *string `json:"forwarded_from,omitempty"`
	// // The Group SID associated with this call. If no Group is associated with the call, the field is empty.
	// GroupSid *string `json:"group_sid,omitempty"`
	// // The caller's name if this call was an incoming call to a phone number with caller ID Lookup enabled. Otherwise, empty.
	// CallerName *string `json:"caller_name,omitempty"`
	// // The wait time in milliseconds before the call is placed.
	// QueueTime *string `json:"queue_time,omitempty"`
	// // The unique identifier of the trunk resource that was used for this call. The field is empty if the call was not made using a SIP trunk or if the call is not terminated.
	// TrunkSid *string `json:"trunk_sid,omitempty"`
	// // The URI of this resource, relative to `https://api.twilio.com`.
	// Uri *string `json:"uri,omitempty"`
	// // A list of subresources available to this call, identified by their URIs relative to `https://api.twilio.com`.
	// SubresourceUris *map[string]interface{} `json:"subresource_uris,omitempty"`
}

type ApiV2010Message struct {
	// // The text content of the message
	// Body *string `json:"body,omitempty"`
	// // The number of segments that make up the complete message. SMS message bodies that exceed the [character limit](https://www.twilio.com/docs/glossary/what-sms-character-limit) are segmented and charged as multiple messages. Note: For messages sent via a Messaging Service, `num_segments` is initially `0`, since a sender hasn't yet been assigned.
	// NumSegments *string `json:"num_segments,omitempty"`
	// Direction   *string `json:"direction,omitempty"`
	// // The sender's phone number (in [E.164](https://en.wikipedia.org/wiki/E.164) format), [alphanumeric sender ID](https://www.twilio.com/docs/sms/quickstart), [Wireless SIM](https://www.twilio.com/docs/iot/wireless/programmable-wireless-send-machine-machine-sms-commands), [short code](https://www.twilio.com/en-us/messaging/channels/sms/short-codes), or  [channel address](https://www.twilio.com/docs/messaging/channels) (e.g., `whatsapp:+15554449999`). For incoming messages, this is the number or channel address of the sender. For outgoing messages, this value is a Twilio phone number, alphanumeric sender ID, short code, or channel address from which the message is sent.
	// From *string `json:"from,omitempty"`
	// The recipient's phone number (in [E.164](https://en.wikipedia.org/wiki/E.164) format) or [channel address](https://www.twilio.com/docs/messaging/channels) (e.g. `whatsapp:+15552229999`)
	To *string `json:"to,omitempty"`
	// // The [RFC 2822](https://datatracker.ietf.org/doc/html/rfc2822#section-3.3) timestamp (in GMT) of when the Message resource was last updated
	// DateUpdated *string `json:"date_updated,omitempty"`
	// // The amount billed for the message in the currency specified by `price_unit`. The `price` is populated after the message has been sent/received, and may not be immediately availalble. View the [Pricing page](https://www.twilio.com/en-us/pricing) for more details.
	// Price *string `json:"price,omitempty"`
	// // The description of the `error_code` if the Message `status` is `failed` or `undelivered`. If no error was encountered, the value is `null`. The value returned in this field for a specific error cause is subject to change as Twilio improves errors. Users should not use the `error_code` and `error_message` fields programmatically.
	// ErrorMessage *string `json:"error_message,omitempty"`
	// // The URI of the Message resource, relative to `https://api.twilio.com`.
	// Uri *string `json:"uri,omitempty"`
	// The SID of the [Account](https://www.twilio.com/docs/iam/api/account) associated with the Message resource
	AccountSid *string `json:"account_sid,omitempty"`
	// // The number of media files associated with the Message resource.
	// NumMedia *string `json:"num_media,omitempty"`
	// Status   *string `json:"status,omitempty"`
	// // The SID of the [Messaging Service](https://www.twilio.com/docs/messaging/api/service-resource) associated with the Message resource. A unique default value is assigned if a Messaging Service is not used.
	// MessagingServiceSid *string `json:"messaging_service_sid,omitempty"`
	// // The unique, Twilio-provided string that identifies the Message resource.
	// Sid *string `json:"sid,omitempty"`
	// // The [RFC 2822](https://datatracker.ietf.org/doc/html/rfc2822#section-3.3) timestamp (in GMT) of when the Message was sent. For an outgoing message, this is when Twilio sent the message. For an incoming message, this is when Twilio sent the HTTP request to your incoming message webhook URL.
	// DateSent *string `json:"date_sent,omitempty"`
	// // The [RFC 2822](https://datatracker.ietf.org/doc/html/rfc2822#section-3.3) timestamp (in GMT) of when the Message resource was created
	// DateCreated *string `json:"date_created,omitempty"`
	// The message delivery status (e.g., "queued", "sent", "delivered", "failed", "undelivered").
	Status *string `json:"status,omitempty"`

	// The [error code](https://www.twilio.com/docs/api/errors) returned if the Message `status` is `failed` or `undelivered`. If no error was encountered, the value is `null`.
	ErrorCode *int `json:"error_code,omitempty"`
	// // The currency in which `price` is measured, in [ISO 4127](https://www.iso.org/iso/home/standards/currency_codes.htm) format (e.g. `usd`, `eur`, `jpy`).
	// PriceUnit *string `json:"price_unit,omitempty"`
	// The API version used to process the Message
	ApiVersion *string `json:"api_version,omitempty"`
	// // A list of related resources identified by their URIs relative to `https://api.twilio.com`
	// SubresourceUris *map[string]interface{} `json:"subresource_uris,omitempty"`
}
