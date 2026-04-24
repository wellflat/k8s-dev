#!/bin/sh

#ENDPOINT_URL="https://api.groq.com/openai/v1"
ENDPOINT_URL="https://api.openai.com/v1"

run_aiperf() {
    aiperf profile \
        --model "gpt-5.4-mini" \
        --tokenizer "elements-dev/o200k-base-tokenizer" \
        --url ${ENDPOINT_URL} \
        --endpoint-type "chat" \
        --streaming \
        --concurrency 5 \
        --request-count 200 \
        --random-seed 42 \
        --isl 150 \
        --api-key ${OPENAI_API_KEY} \
        --export-http-trace \
        --show-trace-timing \
        --use-server-token-count
        #--osl 200 \
        #--extra-inputs ignore_eos:true \
}

run_aiperf
