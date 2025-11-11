# Twilio Webhook 实际使用示例

## 📋 快速开始 (3 步骤)

### 1️⃣  启动应用

```bash
cd /Users/op/PersonalSpace/0_project/twilio-prometheus-exporter
go run src/*.go
```

**预期输出:**
```
INFO[...] Starting Twilio Prometheus exporter on :8080
INFO[...] /metrics - Prometheus metrics endpoint
INFO[...] /webhooks/message - Twilio message status callbacks
INFO[...] /webhooks/call - Twilio call status callbacks
INFO[...] /health - Health check endpoint
```

### 2️⃣  在另一个终端测试 webhook

```bash
cd /Users/op/PersonalSpace/0_project/twilio-prometheus-exporter
chmod +x test_webhooks.sh
./test_webhooks.sh http://localhost:8080
```

### 3️⃣  查看生成的指标

```bash
curl http://localhost:8080/metrics | grep twilio_
```

---

## 🎯 实际使用场景

### 场景 1: 监控消息传递成功率

**需求**: 追踪有多少消息成功传递，有多少失败

**实现步骤**:

#### 1. 在 Twilio 控制台配置消息回调

1. 登录 https://console.twilio.com
2. 导航到: **Messaging → Services**
3. 点击你的 Service
4. 设置 **Fallback URL** (生产环境):
   ```
   https://your-domain.com/webhooks/message
   ```
   或本地测试用 ngrok:
   ```bash
   # 在新终端运行
   ngrok http 8080
   # 获得类似的 URL: https://xxxx-xx-xxx-xx.ngrok-free.app
   ```
   然后设置为:
   ```
   https://xxxx-xx-xxx-xx.ngrok-free.app/webhooks/message
   ```

#### 2. 发送测试消息

使用你的 Twilio 应用发送消息:

```python
from twilio.rest import Client

account_sid = "your_account_sid"
auth_token = "your_auth_token"
client = Client(account_sid, auth_token)

message = client.messages.create(
    from_="+1234567890",  # 你的 Twilio 号码
    to="+1987654321",      # 目标号码
    body="Hello from Twilio Webhook Test!"
)

print(f"Message SID: {message.sid}")
```

#### 3. 监控指标

```bash
# 实时查看消息传递指标
watch 'curl -s http://localhost:8080/metrics | grep twilio_message'
```

**期望看到的指标**:

```
# HELP twilio_messages_delivered_total Total number of successfully delivered messages via webhook callbacks.
# TYPE twilio_messages_delivered_total counter
twilio_messages_delivered_total 5

# HELP twilio_messages_failed_total Total number of failed/undelivered messages via webhook callbacks.
# TYPE twilio_messages_failed_total counter
twilio_messages_failed_total{error_code="21614"} 1
twilio_messages_failed_total{error_code="21408"} 2
```

#### 4. 在 Prometheus 中配置

编辑 `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'twilio-exporter'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 10s  # 频繁抓取以捕获所有事件
```

#### 5. Grafana 仪表盘查询

创建以下 PromQL 查询:

**成功率** (近 1 小时):
```promql
rate(twilio_messages_delivered_total[1h]) / (rate(twilio_messages_delivered_total[1h]) + rate(twilio_messages_failed_total[1h]))
```

**每分钟失败消息**:
```promql
rate(twilio_messages_failed_total[1m])
```

**按错误代码分类的失败**:
```promql
topk(5, rate(twilio_messages_failed_total[5m]))
```

---

### 场景 2: 追踪实时呼叫状态

**需求**: 监控进行中的呼叫、已完成和已失败的呼叫

#### 1. 配置呼叫回调

在 Twilio 控制台:

1. 导航到: **Voice → Credentials → TwiML Apps**
2. 编辑你的 TwiML App
3. 设置 **Status Callback URL**:
   ```
   https://your-domain.com/webhooks/call
   ```

或使用代码:

```python
from twilio.rest import Client

account_sid = "your_account_sid"
auth_token = "your_auth_token"
client = Client(account_sid, auth_token)

# 使用 status_callback_method='POST' 和 status_callback_url
call = client.calls.create(
    from_="+1234567890",        # 你的 Twilio 号码
    to="+1987654321",            # 目标号码
    url="http://demo.twilio.com/docs/voice.xml",  # TwiML
    status_callback="https://your-domain.com/webhooks/call",
    status_callback_method="POST"
)

print(f"Call SID: {call.sid}")
```

#### 2. 观察指标变化

**呼叫流程中的指标变化**:

```
# 初始化 (initiated)
twilio_call_callback_total{status="initiated"} 1

# 振铃 (ringing)
twilio_call_callback_total{status="ringing"} 1

# 接通 (in-progress)
twilio_call_callback_total{status="in-progress"} 1

# 完成 (completed)
twilio_call_callback_total{status="completed"} 1
```

#### 3. Prometheus 告警规则

创建 `twilio-alerts.yml`:

```yaml
groups:
  - name: twilio
    interval: 30s
    rules:
      # 告警: 高呼叫失败率
      - alert: HighCallFailureRate
        expr: |
          (rate(twilio_calls_failed_total[5m]) / 
           (rate(twilio_calls_failed_total[5m]) + rate(twilio_calls_completed_total[5m]))) > 0.2
        for: 5m
        annotations:
          summary: "High call failure rate detected"
          description: "More than 20% of calls are failing"

      # 告警: API 错误过多
      - alert: TwilioAPIErrors
        expr: rate(twilio_api_errors_total[5m]) > 1
        for: 2m
        annotations:
          summary: "Twilio API errors detected"
          description: "API error rate is {{ $value }} per second"
```

---

## 🧪 完整测试示例

### 测试脚本 (test_webhooks.sh)

```bash
#!/bin/bash

# 测试消息 webhook
echo "=== 测试消息 Webhook ==="

# 测试 1: 消息已传递
curl -X POST "http://localhost:8080/webhooks/message" \
  -d "MessageSid=SM1234567890" \
  -d "AccountSid=AC1234567890" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "MessageStatus=delivered" \
  -d "ApiVersion=2010-04-01"
echo ""

# 测试 2: 消息失败
curl -X POST "http://localhost:8080/webhooks/message" \
  -d "MessageSid=SM0987654321" \
  -d "AccountSid=AC1234567890" \
  -d "From=%2B1234567890" \
  -d "To=%2Binvalid" \
  -d "MessageStatus=failed" \
  -d "ErrorCode=21614" \
  -d "ApiVersion=2010-04-01"
echo ""

# 测试呼叫 webhook
echo "=== 测试呼叫 Webhook ==="

# 测试 3: 呼叫已启动
curl -X POST "http://localhost:8080/webhooks/call" \
  -d "CallSid=CA1234567890" \
  -d "AccountSid=AC1234567890" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=initiated" \
  -d "ApiVersion=2010-04-01"
echo ""

# 测试 4: 呼叫已完成
curl -X POST "http://localhost:8080/webhooks/call" \
  -d "CallSid=CA1234567890" \
  -d "AccountSid=AC1234567890" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=completed" \
  -d "Duration=120" \
  -d "ApiVersion=2010-04-01"
echo ""

# 查看生成的指标
echo "=== 生成的指标 ==="
curl -s "http://localhost:8080/metrics" | grep twilio_ | head -20
```

### 运行测试

```bash
# 启动应用
go run src/*.go &
sleep 2

# 运行测试脚本
chmod +x test_webhooks.sh
./test_webhooks.sh

# 检查指标
curl http://localhost:8080/metrics | grep -E "twilio_(message|call)_callback"
```

**预期输出**:

```
# HELP twilio_messages_delivered_total Total number of successfully delivered messages via webhook callbacks.
# TYPE twilio_messages_delivered_total counter
twilio_messages_delivered_total 1

# HELP twilio_messages_failed_total Total number of failed/undelivered messages via webhook callbacks.
# TYPE twilio_messages_failed_total counter
twilio_messages_failed_total{error_code="21614"} 1

# HELP twilio_calls_completed_total Total number of completed calls via webhook callbacks.
# TYPE twilio_calls_completed_total counter
twilio_calls_completed_total{call_status="completed"} 1

# HELP twilio_calls_failed_total Total number of failed calls via webhook callbacks.
# TYPE twilio_calls_failed_total counter
twilio_calls_failed_total{disconnect_reason=""} 0
```

---

## 📊 Grafana 仪表盘配置

### Panel 1: 消息交付趋势

```json
{
  "targets": [
    {
      "expr": "rate(twilio_messages_delivered_total[1m])",
      "legendFormat": "Delivered/min"
    },
    {
      "expr": "rate(twilio_messages_failed_total[1m])",
      "legendFormat": "Failed/min"
    }
  ],
  "type": "graph"
}
```

### Panel 2: 错误代码分布

```json
{
  "targets": [
    {
      "expr": "topk(10, sum by (error_code) (twilio_messages_failed_total))",
      "legendFormat": "{{error_code}}"
    }
  ],
  "type": "piechart"
}
```

### Panel 3: 呼叫状态分布

```json
{
  "targets": [
    {
      "expr": "sum by (call_status) (twilio_calls_completed_total)",
      "legendFormat": "{{call_status}}"
    }
  ],
  "type": "stat"
}
```

---

## 🔒 安全考虑

### Webhook 签名验证

Twilio 的每个 webhook 请求都包含 `X-Twilio-Signature` 头。代码中已实现验证:

```go
// 在 webhooks.go 中
ValidateTwilioWebhookSignature(r, authToken)
```

使用时需要将 `authToken` 从配置传入:

```go
// 在 main.go 中修改
mux.HandleFunc("/webhooks/message", func(w http.ResponseWriter, r *http.Request) {
    if !ValidateTwilioWebhookSignature(r, config.TwilioAuthToken) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    webhookMetrics.HandleMessageStatusCallback(w, r)
})
```

### HTTPS 要求

- **本地测试**: 使用 ngrok (`ngrok http 8080`)
- **生产环境**: 必须使用 HTTPS
- **签名验证**: 始终启用 Twilio 签名验证

---

## 📈 关键指标解释

| 指标 | 类型 | 用途 | 告警阈值 |
|------|------|------|---------|
| `twilio_messages_delivered_total` | Counter | 消息交付成功 | N/A |
| `twilio_messages_failed_total` | Counter | 消息交付失败 | > 10/min |
| `twilio_calls_completed_total` | Counter | 已完成呼叫 | N/A |
| `twilio_calls_failed_total` | Counter | 失败呼叫 | > 5/min |
| `twilio_api_errors_total` | Counter | API 错误 | > 1/min |

---

## 🚀 性能优化

### 1. 并发处理

Webhook 处理器使用 goroutine，自动并发:

```go
// http.ListenAndServe 为每个请求创建新的 goroutine
http.ListenAndServe(":8080", mux)
```

### 2. 内存效率

使用同步锁确保线程安全:

```go
wm.mu.Lock()
defer wm.mu.Unlock()
// 更新指标
```

### 3. 指标保留

Prometheus 指标默认保留 15 天。调整:

```yaml
# prometheus.yml
global:
  retention: 30d
```

---

## 🐛 故障排查

### 问题 1: Webhook 未被调用

**检查清单**:
- [ ] 应用在 `:8080` 运行
- [ ] Twilio 控制台中的 webhook URL 正确
- [ ] 本地测试时使用了 ngrok
- [ ] 网络防火墙允许入站连接

**调试**:
```bash
# 查看日志
go run src/*.go 2>&1 | grep -i webhook
```

### 问题 2: 签名验证失败

**原因**: `authToken` 不匹配或请求被篡改

**解决方案**:
```bash
# 验证环境变量
echo $TWILIO_AUTH_TOKEN

# 确保在 main.go 中正确传递
ValidateTwilioWebhookSignature(r, config.TwilioAuthToken)
```

### 问题 3: 指标不显示

**检查**:
```bash
# 验证指标端点
curl http://localhost:8080/metrics | grep twilio

# 测试 webhook
curl -X POST http://localhost:8080/webhooks/message \
  -d "MessageSid=test" \
  -d "MessageStatus=delivered"

# 再次检查指标
curl http://localhost:8080/metrics | grep twilio_message
```

---

## 📚 相关文档

- [Twilio Webhook 文档](https://www.twilio.com/docs/usage/webhooks)
- [Prometheus 指标类型](https://prometheus.io/docs/concepts/metric_types/)
- [Grafana 仪表盘设置](https://grafana.com/docs/grafana/latest/dashboards/)
