#!/bin/sh

BASE_URL=http://0.0.0.0:4000
#BASE_URL=http://localhost:30201

curl $BASE_URL/v1/chat/completions \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer dummy" \
    -d '{
            "model": "gpt-oss-20b",
            "messages": [
                {
                "role": "user",
                "content": "おはようございます"
                }
            ]
        }'
