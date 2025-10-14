#!/bin/sh

curl -d \
'{
  "dataset-id": "HuggingFaceH4/OpenR1-Math-220k-default-verified",
  "model-id": "openai/gpt-oss-20b",
  "request_count": 100
}' \
-H "Content-Type: application/json" -X POST http://localhost:8080/performance
