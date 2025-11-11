# Twilio Prometheus Exporter - Metrics Analysis

## Summary
Based on analysis of Twilio APIs and current exporter implementation, here's an evaluation of monitored metrics.

---

## Current Exporter Metrics

### 1. `twilio_account_balance` ✅ **KEEP & IMPROVE**
**Purpose**: Track current account balance (numeric value) with currency label.

**Use Cases**:
- **Credit depletion alerting**: Set alert when balance drops below threshold (e.g., < $10)
- **Cost trend analysis**: Track balance over time to understand burn rate
- **Budget management**: Ensure account doesn't run out of credits (prevents service interruption)

**API Used**: `FetchBalance()` → `/2010-04-01/Accounts/{AccountSid}/Balance.json`

**Practical Value**: ⭐⭐⭐⭐⭐ **VERY HIGH**
- Simple, direct, actionable
- Real business value: prevents service disruption
- Lightweight API call (minimal latency/cost)

**Recommendation**: Keep. This is essential for any Twilio integration.

---

### 2. `twilio_usage_amount` ✅ **KEEP & IMPROVE**
**Purpose**: Track usage totals by category (SMS, calls, etc.) in the requested time window.

**API Used**: `FetchUsageRecordsToday()` → `/2010-04-01/Accounts/{AccountSid}/Usage/Records/Today.json`

**Data Provided**:
- **Category**: Feature used (e.g., `sms-outbound`, `calls-inbound-tollfree`, `wireless-usage-data`)
- **Usage**: Numeric amount consumed (e.g., 41 SMS segments sent today)
- **UsageUnit**: Unit of measurement (e.g., "segments", "minutes", "recognitions")

**Use Cases**:
1. **Billing forecast**: Aggregate daily usage × unit price to predict monthly bill
   - Example: 41 SMS outbound segments today → 41 × $0.0075 = $0.3075 daily cost
   - Extrapolate to month: $0.3075 × 30 ≈ $9.23
   
2. **Usage anomaly detection**: Alert if daily SMS volume suddenly spikes (compromised credentials?)
   
3. **Service capacity planning**: Track which features are heavily used to plan scaling
   
4. **SLA monitoring**: Ensure usage is within contract limits (e.g., "max 1M API calls/month")

**Practical Value**: ⭐⭐⭐⭐⭐ **VERY HIGH**
- Directly tied to billing
- Enables proactive cost management
- Detects unusual activity (security/fraud)
- Aggregated data (low cardinality): only meaningful categories appear

**Current Implementation**: Good (aggregated by category + usage_unit)

**Recommendation**: Keep and enhance:
- Add optional `USAGE_START_DATE` config for looking back N days (instead of hardcoded today)
- Document mapping of categories to costs (pricing varies by region)
- Add alerts for usage spikes (e.g., 2x baseline)

---

### 3. `twilio_calls_window_count` ❌ **REMOVE or RETHINK**
**Purpose**: Count of calls in requested time window, grouped by status.

**API Used**: `FetchCalls()` → `/2010-04-01/Accounts/{AccountSid}/Calls?StartTime=...&EndTime=...`

**Data Returned**:
- **Status**: Call state (e.g., `queued`, `ringing`, `in-progress`, `completed`, `failed`, `busy`, `no-answer`)
- **To/From**: Phone numbers (but typically not useful for aggregation, high cardinality!)
- **ApiVersion**: Minimal value for metrics

**Critical Problems**:

1. **High Cardinality Labels** ⚠️⚠️⚠️
   - Including `to` and `from` phone numbers would explode cardinality
   - Current code doesn't include them (good), but the API returns them
   - With Prometheus relabeling, users might be tempted to include these → metric explosion

2. **Limited Actionability** 
   - "20 calls in-progress" — so what? Need context:
     - Is 20 normal or abnormal?
     - What's the baseline?
     - This is a snapshot, not a trend
   
3. **Better Data Exists**: Twilio Events API or Webhooks
   - Webhooks give real-time call status events
   - Better for alerting on specific error patterns (e.g., repeated `failed` status)
   - Less polling overhead than scraping `/Calls` every 5 minutes

4. **Time Window Sensitivity**
   - Default window: `-10m` to `-5m` (last 10 min ago for 5 min duration)
   - If no calls in window, metric shows 0 → node reboot? metric bug? unclear
   - With configurable window, users can query yesterday's calls → stale data

5. **Use Case Mismatch**
   - **Real need**: SLA monitoring (e.g., "< 5% call failure rate")
     - Requires aggregated failure count over long period (1 hour, 1 day)
     - Twilio Bill provides this via `/Usage/Records` → use that instead
   - **Real need**: Active calls monitoring
     - Use `/Accounts/{AccountSid}/Calls?Status=in-progress` (current calls only)
     - Or better: webhook callbacks

**Practical Value**: ⭐ **LOW** (mostly noise, fragile)

**Recommendation**: **REMOVE** this metric. 

**If users need call metrics, guide them to:**
1. Use Twilio Webhooks for event-driven alerting (real-time, no polling)
2. Query `/Usage/Records` for SLA tracking (aggregated, billing-aware)
3. Use Twilio Monitor API for real-time call events with structured logging

---

### 4. `twilio_messages_window_count` ❌ **REMOVE or RETHINK**
**Purpose**: Count of messages in requested time window, grouped by status.

**API Used**: `FetchMessages()` → `/2010-04-01/Accounts/{AccountSid}/Messages?DateSentAfter=...&DateSentBefore=...`

**Data Returned**:
- **Status**: Message state (e.g., `queued`, `sent`, `delivered`, `failed`, `undelivered`)
- **ErrorCode**: Optional error if failed (e.g., `30003` = Unreachable destination handset)
- **To/From**: Phone numbers (high cardinality! ⚠️)

**Critical Problems**:

1. **High Cardinality Labels** ⚠️⚠️⚠️⚠️⚠️ **Even Worse Than Calls**
   - Messages often used for marketing/notifications → thousands of unique recipients
   - Including `to` label would crash monitoring (Prometheus cardinality limit)
   - Current code excludes it, but it's a footgun

2. **Limited Actionability** 
   - "50 messages delivered" — insufficient context:
     - Without timestamp granularity, hard to detect rate changes
     - Example: 50 delivered today could be normal or catastrophic (depends on baseline)
   - Better: track delivery rate (%) not just count

3. **Better Data Exists**: Twilio Webhooks + Callbacks
   - Message Status Callbacks provide real-time delivery reports
   - Enable alerting on actual business impact (e.g., "delivery rate < 95%")
   - No polling required

4. **Error Code Blind Spot**
   - Only recent Twilio Go SDK versions expose `Status` field
   - ErrorCode is returned but hard to aggregate (code `30003` vs `21614` → different root causes)
   - Would need mapping table to be useful

5. **Time Window Problem** (same as calls)
   - Stale data risk with configurable start/end dates
   - Snapshot metric doesn't show trend

6. **Use Case Mismatch**
   - **Real need**: Billing/quota tracking
     - Use `/Usage/Records` instead (e.g., `sms-outbound: 1000 segments today`)
     - This includes ALL messages (not just sampled in scrape window)
   - **Real need**: Delivery SLA monitoring
     - Track delivery rate (failed/total) over 1-day or 1-hour windows
     - Requires stateful aggregation or Prometheus recording rules

**Practical Value**: ⭐ **LOW** (unreliable, high cardinality risk)

**Recommendation**: **REMOVE** this metric.

**If users need message metrics, guide them to:**
1. Use Twilio Message Status Callbacks for real-time delivery tracking
2. Query `/Usage/Records` for billing-accurate message counts
3. Set up Prometheus alerts on callback webhook delivery rates (e.g., via custom app)

---

## Recommendations Summary

### Keep
| Metric | Reason |
|--------|--------|
| `twilio_account_balance` | Essential for credit alerts; direct business value |
| `twilio_usage_amount` | Billing-accurate; enables cost forecasting & anomaly detection |
| `twilio_api_errors_total` | Good practice; tracks exporter health |

### Remove
| Metric | Reason |
|--------|--------|
| `twilio_calls_window_count` | Low cardinality at status level only; better via webhooks; snapshot data |
| `twilio_messages_window_count` | Massive cardinality risk (To/From labels); better via webhooks; snapshot data |

### Refactor
| Metric | Action |
|--------|--------|
| `twilio_usage_amount` | Add configurable date range; improve docs on pricing; add example Prometheus rules |

---

## Architecture Decision

### Current Problem
Exporter polls Twilio API every scrape interval (default 5min). Calls/Messages endpoints are:
- **Slow**: Paginate through potentially 1000s of records
- **Stale**: Snapshot at scrape time, not real-time
- **Unreliable**: High cardinality if additional labels exposed

### Better Approach: Hybrid Model

1. **Keep: Usage Records Scraping** (batch/aggregated)
   - `FetchBalance()` - lightweight, essential
   - `FetchUsageRecordsToday()` - already aggregated by Twilio, billing-accurate

2. **Add: Webhook Ingestion** (real-time events)
   - Message Status Callbacks → HTTP endpoint in exporter
   - Call Status Events → HTTP endpoint in exporter
   - Push data to Prometheus (no scraping needed)
   - Real-time alerting, lower latency, accurate delivery/failure tracking

3. **Remove: Raw Call/Message Polling**
   - Causes high cardinality risk
   - Doesn't reflect real business metrics
   - Webhooks are superior

---

## Real-World Monitoring Use Cases

### ✅ Supported Well by Current Exporter

**Use Case 1: Budget Alert**
```
ALERT LowTwilioBalance
  IF twilio_account_balance{currency="USD"} < 10
  FOR 5m
  THEN send alert
```

**Use Case 2: Cost Forecast**
```
daily_sms_cost = twilio_usage_amount{category="sms-outbound"} * 0.0075
predicted_monthly_cost = daily_sms_cost * 30
ALERT if predicted_monthly_cost > $500
```

**Use Case 3: Anomaly Detection**
```
sms_spike = rate(twilio_usage_amount{category="sms-outbound"}[5m]) > 5
ALERT if sms_spike
```

### ❌ NOT Well-Supported

**Use Case 4: "Alert if call success rate < 95%"**
- ❌ Current `twilio_calls_window_count` snapshots don't aggregate properly
- ✅ Better: Webhook callback aggregation → Prometheus gauge per minute

**Use Case 5: "Alert if 10+ message deliveries failed in 5 min"**
- ❌ Current `twilio_messages_window_count` snapshots unreliable
- ✅ Better: Webhook callback aggregation → Prometheus counter

---

## Conclusion

| Metric | Status | Confidence |
|--------|--------|-----------|
| `twilio_account_balance` | ✅ Keep | 99% |
| `twilio_usage_amount` | ✅ Keep & Document | 98% |
| `twilio_calls_window_count` | ❌ Remove | 85% |
| `twilio_messages_window_count` | ❌ Remove | 90% |

**Reasoning**: Focus on aggregated, low-cardinality, billing-accurate metrics. Delegate real-time event tracking to webhooks. This improves reliability, scalability, and actionability.
