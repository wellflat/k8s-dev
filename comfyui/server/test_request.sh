#!/bin/sh

curl -X POST \
  http://localhost:3000/prompt \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "A girl at fantasy landscape with mountains and a river"}'