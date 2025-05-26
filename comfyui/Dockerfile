# Multi-stage build, build stage
# APIサーバなどを含むことを想定してbuild stageをセット
# 現状は特に処理なし
FROM python:3.12-slim AS builder

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /workspace

# Multi-stage build, package stage
FROM pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime

ENV DEBIAN_FRONTEND=noninteractive \
    PIP_PREFER_BINARY=1 \
    PYTHONUNBUFFERED=1 \
    NVIDIA_VISIBLE_DEVICES=all \
    NVIDIA_DRIVER_CAPABILITIES=all

WORKDIR /workspace

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    apt-get update &&  \
    apt-get install -y git libgl1-mesa-dev libglib2.0-0 && \
    rm -rf /var/lib/apt/lists/*

COPY --chmod=775 ComfyUI/ ComfyUI/
COPY --chmod=775 ComfyUI-Manager/ ComfyUI/custom_nodes/ComfyUI-Manager/
COPY --chmod=775 ./entrypoint.sh /workspace/entrypoint.sh

RUN --mount=type=cache,target=/root/.cache/pip,sharing=locked \
    pip3 install -r ComfyUI/requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip,sharing=locked \
    pip3 install -r ComfyUI/custom_nodes/ComfyUI-Manager/requirements.txt

ENV COMFYUI_PATH=/workspace/ComfyUI \
    COMFYUI_MODEL_DIR=/workspace/ComfyUI/models

CMD ["/bin/bash", "entrypoint.sh"]