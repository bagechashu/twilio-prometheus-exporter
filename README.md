# twilio-prometheus-exporter

This application serves as a metrics exporter for Twilio, a cloud communications platform. It retrieves various metrics such as account balance, usage records, calls, and messages from Twilio's API and exposes them in a format suitable for monitoring using Prometheus.

## Configuration

This application uses environment variables for configuration. You can create a .env file in the project root directory to specify these variables. An example .env file may look like this:
```sh
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
LOG_LEVEL=info
RECORD_LIMIT=200
START_DATE=-10m
DURATION=5m
```
Make sure to replace your_twilio_account_sid and your_twilio_auth_token with your actual Twilio account credentials.

## Usage
To run the application, execute the built binary:
``` bash
./twilio-metrics-exporter
```
This will start the exporter server, which listens for HTTP requests on port 8080 by default.
You can access the metrics endpoint at `http://localhost:8080/metrics`.

## Prometheus Configuration
To scrape metrics from this exporter, add the following job configuration to your Prometheus prometheus.yml file:
```yaml
scrape_configs:
- job_name: 'twilio_metrics'
  scrape_interval: 5m
  scrape_align_interval: 5m
  static_configs:
  - targets: ['localhost:8080']
  metrics_path: /metrics
```
Make sure Prometheus can reach the exporter at the specified target address.

## Metrics
The following metrics are exported:
 - twilio_account_balance: Current balance of the Twilio account.
 - twilio_usage: The amount used to bill usage, measured in usage units.
 - twilio_calls: Number of calls made or received.
 - twilio_messages: Number of messages sent or received.
