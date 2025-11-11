#!/bin/bash

# Twilio Webhook 测试脚本
# 用于模拟 Twilio 发送的回调请求

WEBHOOK_URL="${1:-http://localhost:8080}"

echo "=== Twilio Webhook 测试 ==="
echo "Webhook 基础 URL: $WEBHOOK_URL"
echo ""

# 测试 1: 消息传递成功
echo "测试 1: 模拟消息已传递 (delivered)"
curl -X POST "$WEBHOOK_URL/webhooks/message" \
  -d "MessageSid=SMtest001" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "MessageStatus=delivered" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 2: 消息已发送
echo "测试 2: 模拟消息已发送 (sent)"
curl -X POST "$WEBHOOK_URL/webhooks/message" \
  -d "MessageSid=SMtest002" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543211" \
  -d "MessageStatus=sent" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 3: 消息发送失败
echo "测试 3: 模拟消息发送失败 (failed)"
curl -X POST "$WEBHOOK_URL/webhooks/message" \
  -d "MessageSid=SMtest003" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2Binvalid" \
  -d "MessageStatus=failed" \
  -d "ErrorCode=21614" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 4: 呼叫已启动
echo "测试 4: 模拟呼叫已启动 (initiated)"
curl -X POST "$WEBHOOK_URL/webhooks/call" \
  -d "CallSid=CAtest001" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=initiated" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 5: 呼叫振铃中
echo "测试 5: 模拟呼叫振铃中 (ringing)"
curl -X POST "$WEBHOOK_URL/webhooks/call" \
  -d "CallSid=CAtest001" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=ringing" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 6: 呼叫已接通
echo "测试 6: 模拟呼叫已接通 (in-progress)"
curl -X POST "$WEBHOOK_URL/webhooks/call" \
  -d "CallSid=CAtest001" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=in-progress" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 7: 呼叫已完成
echo "测试 7: 模拟呼叫已完成 (completed)"
curl -X POST "$WEBHOOK_URL/webhooks/call" \
  -d "CallSid=CAtest001" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543210" \
  -d "CallStatus=completed" \
  -d "Duration=120" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 等待 1 秒
sleep 1

# 测试 8: 呼叫失败
echo "测试 8: 模拟呼叫失败 (failed)"
curl -X POST "$WEBHOOK_URL/webhooks/call" \
  -d "CallSid=CAtest002" \
  -d "AccountSid=ACtest001" \
  -d "From=%2B1234567890" \
  -d "To=%2B9876543212" \
  -d "CallStatus=failed" \
  -d "ApiVersion=2010-04-01"
echo -e "\n✅ 请求已发送\n"

# 现在查看指标
echo "=== 检查生成的指标 ==="
echo ""
echo "获取指标..."
curl -s "$WEBHOOK_URL/metrics" | grep -E "twilio_(message|call)_(callback_total|status_gauge)"
echo ""
echo "=== 测试完成 ==="
