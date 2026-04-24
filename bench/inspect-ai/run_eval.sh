#!/bin/sh

TASK=inspect_evals/personality_TRAIT
#MODEL=openai/google/gemma-4-31b-it
MODEL=openai/llm-jp-4-32b-a3b-thinking
#BASE_URL=https://openrouter.ai/api/v1
BASE_URL=http://192.168.250.103:8000/v1
LOG_DIR=./logs
inspect eval ${TASK} \
    --model ${MODEL} \
    --model-base-url ${BASE_URL} \
    --log-dir ${LOG_DIR}

