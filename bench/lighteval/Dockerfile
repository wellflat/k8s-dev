# syntax=docker/dockerfile:1

# ==============================================================================
# Builder Stage
#
# このステージでは、lightevalのソースコードをクローンし、
# 必要なPython依存関係をインストールします。
# ==============================================================================
FROM nvidia/cuda:12.6.0-devel-ubuntu22.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

# システムの依存関係をインストール (git, python)
RUN apt-get update && \
    apt-get install -y --no-install-recommends software-properties-common && \
    add-apt-repository ppa:deadsnakes/ppa && \
    apt-get update && \
    apt-get install -y --no-install-recommends git python3.12 python3.12-venv python3-pip \
    && rm -rf /var/lib/apt/lists/*

# python3とpipのシンボリックリンクを設定
RUN update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1 && \
    update-alternatives --install /usr/bin/python python /usr/bin/python3.12 1


# Pythonの仮想環境を作成
ENV VENV_PATH=/opt/venv
RUN python3 -m venv $VENV_PATH
ENV PATH="$VENV_PATH/bin:$PATH"

# pipをアップグレード
RUN pip install --upgrade pip

# lightevalのソースコードをクローン
WORKDIR /app
RUN git clone https://github.com/wellflat/lighteval.git .

# PyTorch (CUDA対応版) をインストール
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install torch --index-url https://download.pytorch.org/whl/cu121

# lightevalの依存関係をインストール (PyTorchはインストール済みのためスキップされる)
# extrasとしてlitellmプロキシに必要なendpointを含める
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install ".[endpoint]" emoji "litellm[caching]" accelerate

# ==============================================================================
# Final Stage
#
# このステージでは、ビルドステージでインストールしたライブラリと
# ソースコードをコピーし、実行可能なイメージを作成します。
# ==============================================================================
FROM nvidia/cuda:12.6.0-runtime-ubuntu22.04

ENV DEBIAN_FRONTEND=noninteractive

# 実行に必要な依存関係をインストール (git, curl, python)
RUN apt-get update && \
    apt-get install -y --no-install-recommends software-properties-common && \
    add-apt-repository ppa:deadsnakes/ppa && \
    apt-get update && \
    apt-get install -y --no-install-recommends git curl python3.12 \
    && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1 && \
    update-alternatives --install /usr/bin/python python /usr/bin/python3.12 1

# Builderステージから仮想環境をコピー
# これにより、最終イメージにビルドツールやヘッダーファイルが含まれるのを防ぎます
ENV VENV_PATH=/opt/venv
COPY --from=builder $VENV_PATH $VENV_PATH
ENV PATH="$VENV_PATH/bin:$PATH"
ENV OPENAI_API_KEY="dummy"

# Builderステージからソースコードをコピー
WORKDIR /app
COPY --from=builder /app .
COPY config.yaml run_eval.sh .

# コンテナ起動時のデフォルトコマンドとしてbashを起動
# これにより、コンテナ内でインタラクティブにlightevalコマンドを実行できます
CMD ["bash"]
