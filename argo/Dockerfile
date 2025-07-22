FROM python:3.10-slim

ENV HUGGINGFACE_HUB_CACHE="/root/.cache/huggingface"
ENV PIP_NO_CACHE_DIR=yes

# パッケージマネージャを更新し、gitをインストール
# huggingface-cliがリポジトリのクローンなどでgitを利用するため
RUN apt-get update && \
    apt-get install -y --no-install-recommends git && \
    rm -rf /var/lib/apt/lists/*

RUN pip install --upgrade pip && \
    pip install "huggingface-hub[cli]"

WORKDIR /app

ENTRYPOINT ["huggingface-cli"]

# ENTRYPOINTに引数が渡されなかった場合のデフォルト引数
CMD ["--help"]
