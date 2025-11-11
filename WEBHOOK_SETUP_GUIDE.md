# Twilio Webhook 集成指南

## 概述

exporter 提供两个 webhook 端点来接收 Twilio 的实时事件：
- `/webhooks/message` - 消息状态回调 (Message Status Callbacks)
- `/webhooks/call` - 呼叫状态事件 (Call Status Events)

这些端点能实时捕获消息传递和呼叫状态变化，生成 Prometheus 指标。

---

## 1. 消息状态回调 (Message Status Callback)

### 什么是消息状态回调？
当 SMS 消息发送时，Twilio 会在状态改变时向你的 webhook 发送回调，告诉你消息的传递状态。

### 可能的消息状态
- `queued` - 消息已排队，等待发送
- `sent` - 消息已发送到运营商
- `delivered` - 消息已送达用户手机
- `failed` - 消息发送失败
- `undelivered` - 消息无法送达（如无效号码）

### 配置步骤

#### Step 1: 确保 exporter 可公网访问

如果在本地开发，需要使用 ngrok 或类似工具暴露端口：

```bash
# 终端 1: 运行 exporter
cd /Users/op/PersonalSpace/0_project/twilio-prometheus-exporter
go run ./src

# 终端 2: 使用 ngrok 暴露本地端口
ngrok http 8080
# 输出示例:
# Forwarding                    https://1234-56-78-90-123.ngrok.io -> http://localhost:8080
```

你的 webhook URL 是：`https://1234-56-78-90-123.ngrok.io/webhooks/message`

#### Step 2: 在 Twilio Console 配置回调 URL

1. 登录 [Twilio Console](https://www.twilio.com/console)
2. 进入 **Messaging** → **Services**
3. 选择你的服务（或创建新的）
4. 找到 **Webhook & Callbacks** 部分
5. 设置 **Webhook Fallback URL** 为你的 ngrok URL：
   ```
   https://1234-56-78-90-123.ngrok.io/webhooks/message
   ```
6. 确保选择 **HTTP POST** 方法
7. 保存更改

#### Step 3: 发送测试消息

使用 Twilio CLI 或 Python SDK 发送 SMS：

```bash
# 使用 Twilio CLI
twilio api:core:messages:create \
  --from=+1234567890 \
  --to=+9876543210 \
  --body="Hello World"
```

#### Step 4: 查看 Prometheus 指标

访问 `http://localhost:8080/metrics`，你将看到：

```
# 消息回调计数器
# HELP twilio_message_callback_total Total message callbacks received from Twilio
# TYPE twilio_message_callback_total counter
twilio_message_callback_total{status="delivered"} 1
twilio_message_callback_total{status="sent"} 1
twilio_message_callback_total{status="failed"} 0

# 消息状态分布
# HELP twilio_message_status_gauge Current message delivery status
# TYPE twilio_message_status_gauge gauge
twilio_message_status_gauge{status="delivered"} 1
twilio_message_status_gauge{status="sent"} 1
twilio_message_status_gauge{status="failed"} 0
```

### Webhook Payload 示例

当消息状态改变时，Twilio 会发送 POST 请求到你的 webhook，包含以下数据：

```
POST https://your-webhook-url/webhooks/message

Form Data:
  MessageSid=SMxxxx
  AccountSid=ACxxxx
  From=+1234567890
  To=+9876543210
  MessageStatus=delivered
  ErrorCode=  (empty if success)
  ApiVersion=2010-04-01
```

### 示例 1: Python 发送 SMS 并接收回调

```python
from twilio.rest import Client

# 初始化 Twilio 客户端
account_sid = "your_account_sid"
auth_token = "your_auth_token"
client = Client(account_sid, auth_token)

# 发送消息
message = client.messages.create(
    body="Hello from Python!",
    from_="+1234567890",  # 你的 Twilio 号码
    to="+9876543210"      # 接收者号码
)

print(f"Message SID: {message.sid}")
print(f"Message Status: {message.status}")
```

当消息状态改变时（sent → delivered），Twilio 会自动向 `/webhooks/message` 发送回调。

---

## 2. 呼叫状态事件 (Call Status Callback)

### 什么是呼叫状态回调？
当发起呼叫时，Twilio 会在呼叫状态改变时发送回调。

### 可能的呼叫状态
- `queued` - 呼叫已排队
- `ringing` - 对方手机正在振铃
- `in-progress` - 呼叫已接通
- `completed` - 呼叫已完成
- `failed` - 呼叫失败
- `busy` - 对方忙线
- `no-answer` - 对方未接听

### 配置步骤

#### Step 1: 在 TwiML Bin 中创建呼叫处理逻辑

1. 进入 [Twilio Console](https://www.twilio.com/console)
2. 进入 **Dev Tools** → **TwiML Bins**
3. 创建新的 TwiML Bin，例如：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say>Welcome to my Twilio app!</Say>
    <Gather numDigits="1" action="https://your-webhook-url/call-handler">
        <Say>Press 1 to continue</Say>
    </Gather>
</Response>
```

#### Step 2: 配置状态回调 URL

在发起呼叫时指定回调 URL（使用 Twilio SDK）：

```python
from twilio.rest import Client

client = Client(account_sid, auth_token)

call = client.calls.create(
    to="+9876543210",           # 接收者号码
    from_="+1234567890",        # 你的 Twilio 号码
    url="https://your-twiml-url",  # TwiML Bin URL
    status_callback="https://1234-56-78-90-123.ngrok.io/webhooks/call",  # 状态回调
    status_callback_method="POST",
    status_callback_events=["initiated", "ringing", "answered", "completed"]
)

print(f"Call SID: {call.sid}")
```

#### Step 3: 查看 Prometheus 指标

访问 `http://localhost:8080/metrics`：

```
# 呼叫回调计数器
# HELP twilio_call_callback_total Total call callbacks received from Twilio
# TYPE twilio_call_callback_total counter
twilio_call_callback_total{status="initiated"} 1
twilio_call_callback_total{status="ringing"} 1
twilio_call_callback_total{status="answered"} 1
twilio_call_callback_total{status="completed"} 1
twilio_call_callback_total{status="failed"} 0

# 呼叫状态分布
# HELP twilio_call_status_gauge Current call status
# TYPE twilio_call_status_gauge gauge
twilio_call_status_gauge{status="completed"} 1
twilio_call_status_gauge{status="failed"} 0
```

### Webhook Payload 示例

```
POST https://your-webhook-url/webhooks/call

Form Data:
  CallSid=CAxxxx
  AccountSid=ACxxxx
  From=+1234567890
  To=+9876543210
  CallStatus=completed
  Duration=45
  Timestamp=Mon, 11 Nov 2025 15:30:00 +0000
  CallbackSource=call-status-events
```

---

## 3. 完整示例：发送消息 + 监控传递

### Python 脚本示例

```python
#!/usr/bin/env python3
import os
import time
from twilio.rest import Client

# 从环境变量读取凭证
account_sid = os.getenv("TWILIO_ACCOUNT_SID")
auth_token = os.getenv("TWILIO_AUTH_TOKEN")

client = Client(account_sid, auth_token)

# 发送 5 条测试消息
for i in range(5):
    message = client.messages.create(
        body=f"Test message #{i+1} at {time.strftime('%H:%M:%S')}",
        from_="+1234567890",
        to="+9876543210"
    )
    print(f"[{i+1}] Sent message {message.sid} with status: {message.status}")
    time.sleep(2)

print("\n✅ All messages sent!")
print("Monitor delivery status at: http://localhost:8080/metrics")
print("\nMetric to check:")
print("  twilio_message_callback_total{status=\"delivered\"}")
print("  twilio_message_status_gauge{status=\"delivered\"}")
```

运行脚本：

```bash
# 设置凭证
export TWILIO_ACCOUNT_SID="your_account_sid"
export TWILIO_AUTH_TOKEN="your_auth_token"

python3 send_messages.py
```

---

## 4. Prometheus 告警示例

### 告警 1: 消息传递率过低

```yaml
groups:
  - name: twilio_alerts
    rules:
      - alert: LowMessageDeliveryRate
        expr: |
          (twilio_message_callback_total{status="delivered"} / 
           (twilio_message_callback_total{status="delivered"} + 
            twilio_message_callback_total{status="failed"})) < 0.95
        for: 5m
        annotations:
          summary: "消息传递率低于 95%"
          description: "过去 5 分钟内消息传递率为 {{ $value | humanizePercentage }}"
```

### 告警 2: 消息发送失败过多

```yaml
- alert: MessageSendingFailures
  expr: rate(twilio_message_callback_total{status="failed"}[5m]) > 1
  for: 1m
  annotations:
    summary: "消息发送失败率过高"
    description: "过去 1 分钟每秒失败消息数: {{ $value }}"
```

### 告警 3: 呼叫失败率高

```yaml
- alert: HighCallFailureRate
  expr: |
    (twilio_call_callback_total{status="failed"} / 
     twilio_call_callback_total) > 0.1
  for: 5m
  annotations:
    summary: "呼叫失败率高于 10%"
    description: "{{ $value | humanizePercentage }} 的呼叫失败"
```

---

## 5. 故障排除

### 问题 1: Webhook 未收到回调

**检查清单**:
1. ✅ 确认 exporter 正在运行：`curl http://localhost:8080/health`
2. ✅ 确认 ngrok 仍在运行（有时会超时）
3. ✅ 验证 Twilio 中配置的 URL 是否正确
4. ✅ 检查 exporter 日志中是否有错误

### 问题 2: 指标数据没有更新

**可能原因**:
- 消息/呼叫尚未完成（等待状态变化）
- Webhook URL 配置错误
- 防火墙阻止回调

**解决方案**:
```bash
# 查看 exporter 日志
tail -f exporter.log | grep webhook

# 手动测试 webhook（模拟 Twilio 回调）
curl -X POST http://localhost:8080/webhooks/message \
  -d "MessageSid=SMtest123" \
  -d "MessageStatus=delivered"
```

### 问题 3: ngrok URL 不稳定

**长期解决方案** - 使用公网服务器：
1. 在云服务器上运行 exporter（如 AWS EC2）
2. 配置 Twilio 指向服务器的真实 IP/域名
3. 设置 SSL 证书（Twilio 建议 HTTPS）

---

## 6. 最佳实践

### 1️⃣ 验证 Webhook 来源
添加 Twilio 的 auth token 验证，确保请求来自 Twilio：

```python
from twilio.request_validator import RequestValidator

def verify_twilio_request(auth_token, url, post_data, signature):
    validator = RequestValidator(auth_token)
    return validator.validate(url, post_data, signature)
```

### 2️⃣ 添加请求超时
```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

### 3️⃣ 监控 Webhook 处理延迟
```go
// 在 webhook 处理前记录开始时间
start := time.Now()
// ... 处理逻辑
duration := time.Since(start)
logrus.Infof("Webhook processed in %v", duration)
```

### 4️⃣ 实现重试机制
如果 Twilio 未收到 2xx 响应，会重试回调。确保你的端点：
- ✅ 立即返回 200 OK
- ✅ 异步处理复杂逻辑
- ✅ 幂等（能安全处理重复）

---

## 7. 监控仪表板示例

### Grafana 查询示例

**消息传递成功率**:
```
sum(twilio_message_callback_total{status="delivered"}) / 
sum(twilio_message_callback_total)
```

**呼叫完成率**:
```
sum(twilio_call_callback_total{status="completed"}) / 
sum(twilio_call_callback_total)
```

**每分钟消息数**:
```
rate(twilio_message_callback_total[1m])
```

---

## 8. 总结

| 功能 | 端点 | 用途 |
|------|------|------|
| 消息状态回调 | `/webhooks/message` | 追踪 SMS 送达状态 |
| 呼叫状态回调 | `/webhooks/call` | 追踪呼叫完成/失败 |
| 健康检查 | `/health` | 验证 exporter 运行状态 |
| Prometheus 指标 | `/metrics` | Prometheus scrape endpoint |

**下一步**:
1. ✅ 使用 ngrok 暴露本地服务
2. ✅ 在 Twilio Console 配置 webhook URL
3. ✅ 发送测试消息/呼叫
4. ✅ 验证指标收集
5. ✅ 配置 Prometheus 告警
