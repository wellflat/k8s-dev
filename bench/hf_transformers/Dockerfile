FROM nvidia/cuda:12.6.0-devel-ubuntu22.04

# cu126 or cu128
ARG TORCH_CUDA_VERSION=cu126

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    build-essential \
    python3 \
    python3-pip \
    git \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --pre torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/${TORCH_CUDA_VERSION}
RUN pip install --no-cache-dir -r requirements.txt --timeout 300 --retries 5

WORKDIR /workspace
COPY src/app.py /workspace/app.py

CMD [ "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000", "--reload" ]
