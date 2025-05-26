#!/bin/bash

if [ "${LOWVRAM}" -eq 1 ]; then
    python3 ComfyUI/main.py --listen 0.0.0.0 --port ${COMFYUI_PORT} --lowvram
else
    python3 ComfyUI/main.py --listen 0.0.0.0 --port ${COMFYUI_PORT}
fi
