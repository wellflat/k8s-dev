#/bin/sh

HUGGINGFACE_HUB_TOKEN=token_here

docker run --gpus all \
    -v ~/.cache/huggingface:/root/.cache/huggingface \
    --env "HUGGING_FACE_HUB_TOKEN=${HUGGINGFACE_HUB_TOKEN}" \
    -p 8000:8000 \
    --ipc=host \
    vllm/vllm-openai:latest \
    --model elyza/Llama-3-ELYZA-JP-8B \
    --dtype=auto \
    --gpu-memory-utilization 0.5 \
    #--model meta-llama/Llama-3.1-8B

# INFO 06-26 00:52:05 [launcher.py:29] Available routes are:
# INFO 06-26 00:52:05 [launcher.py:37] Route: /openapi.json, Methods: HEAD, GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /docs, Methods: HEAD, GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /docs/oauth2-redirect, Methods: HEAD, GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /redoc, Methods: HEAD, GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /health, Methods: GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /load, Methods: GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /ping, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /ping, Methods: GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /tokenize, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /detokenize, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/models, Methods: GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /version, Methods: GET
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/chat/completions, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/completions, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/embeddings, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /pooling, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /classify, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /score, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/score, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/audio/transcriptions, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /rerank, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v1/rerank, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /v2/rerank, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /invocations, Methods: POST
# INFO 06-26 00:52:05 [launcher.py:37] Route: /metrics, Methods: GET
