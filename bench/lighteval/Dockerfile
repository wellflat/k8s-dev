# syntax=docker/dockerfile:1

# ==============================================================================
# Builder Stage
#
# このステージでは、lightevalのソースコードをクローンし、
# 必要なPython依存関係をインストールします。
# ==============================================================================
FROM python:3.12-slim AS builder

# システムの依存関係をインストール (git)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Pythonの仮想環境を作成
ENV VENV_PATH=/opt/venv
RUN python3 -m venv $VENV_PATH
ENV PATH="$VENV_PATH/bin:$PATH"

# pipをアップグレード
RUN pip install --upgrade pip

# lightevalのソースコードをクローン
WORKDIR /app
RUN git clone https://github.com/huggingface/lighteval.git .

# lightevalの依存関係をインストール
# extrasとしてlitellmプロキシに必要なendpointを含める
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install ".[endpoint]" emoji "litellm[caching]" accelerate

# ==============================================================================
# Final Stage
#
# このステージでは、ビルドステージでインストールしたライブラリと
# ソースコードをコピーし、実行可能なイメージを作成します。
# ==============================================================================
FROM python:3.12-slim

# gitをインストール (lightevalの依存ライブラリGitPythonが実行時に必要とするため)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git curl \
    && rm -rf /var/lib/apt/lists/*

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
