#!/bin/bash

SERVER_URL=http://192.168.49.2:30201
curl ${SERVER_URL}/v1/chat/completions -H "Content-Type: application/json" -d '{
  "model": "openai/gpt-oss-20b",
  "messages": [
    {
      "role": "user",
      "content": "こんにちは、はじめまして"
    }
  ]
}' 