#!/bin/sh

TASK=inspect_evals/personality_TRAIT
MODEL=openai/google/gemma-4-31b-it
BASE_URL=https://openrouter.ai/api/v1
LOG_DIR=./logs
inspect eval ${TASK} \
    --model ${MODEL} \
    --model-base-url ${BASE_URL} \
    --log-dir ${LOG_DIR}

#https://api.openai.com/v1
