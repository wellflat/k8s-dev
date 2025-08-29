#!/bin/bash

MODEL_ID=openai/gpt-oss-20b
SERVER_URL=http://192.168.49.2:30201

run_perf_synthetic() {
    genai-perf profile -m ${MODEL_ID} \
           --service-kind openai \
           --endpoint-type chat \
           --synthetic-input-tokens-mean 150 \
           --synthetic-input-tokens-stddev 30 \
           --output-tokens-mean 100 \
           --output-tokens-stddev 30 \
           --num-prompts 200 \
           --concurrency 1 \
           --random-seed 42 \
           --streaming \
           --request-count 100 \
           --warmup-request-count 10 \
           --url ${SERVER_URL} \
           --artifact-dir /workspace/profiles/${MODEL_ID} \
           --generate-plots
}

run_perf_dataset() {
    genai-perf profile -m ${MODEL_ID} \
           --input-file /workspace/profiles/prompts.jsonl \
           --service-kind openai \
           --endpoint-type chat \
           --output-tokens-mean 100 \
           --output-tokens-stddev 30 \
           --num-prompts 30 \
           --concurrency 1 \
           --random-seed 42 \
           --streaming \
           --request-count 100 \
           --warmup-request-count 10 \
           --url ${SERVER_URL} \
           --artifact-dir /workspace/profiles/${MODEL_ID} \
           --generate-plots
}

run_perf_dataset