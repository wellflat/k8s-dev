#!/bin/sh

MODEL_ID=meta-llama/Llama-3.1-8B
ENDPOINT=$(minikube service vllm-service --url)

genai-perf profile -m ${MODEL_ID} \
            --endpoint-type chat \
            --synthetic-input-tokens-mean 150 \
            --synthetic-input-tokens-stddev 30 \
            --output-tokens-mean 300 \
            --output-tokens-stddev 30 \
            --num-prompts 300  \
            --concurrency 1 \
            --measurement-interval 180000 \
            --random-seed 42 \
            --url ${ENDPOINT} \
            --artifact-dir ~/.cache/huggingface/hub/benchmark_results/genai-perf/${MODEL_ID}/vllm \
            --generate-plots