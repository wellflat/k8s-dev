#!/bin/sh

run_aiperf() {
    aiperf profile \
        --model "openai/gpt-oss-20b" \
        --url "http://localhost:30201" \
        --endpoint-type "chat" \
        --streaming \
        --concurrency 2 \
        --request-count 100 \
        --output-tokens-mean 200 \
        --random-seed 42 \
        --api-key "test key" \
        --gpu-telemetry 
}

run_aiperf
