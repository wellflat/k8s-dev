#!/bin/bash

COMFYUI_PORT=8188
LOWVRAM=1
if [ "$LOWVRAM" -eq 1 ]; then
    python3 ComfyUI/main.py --listen 0.0.0.0 --port $COMFYUI_PORT --lowvram
else
    python3 ComfyUI/main.py --listen 0.0.0.0 --port $COMFYUI_PORT
fi
