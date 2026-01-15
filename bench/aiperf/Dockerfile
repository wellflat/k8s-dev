# build stage
FROM python:3.11-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --user --no-cache-dir aiperf


# package stage
FROM python:3.11-slim

WORKDIR /app

# copy libraries and executables
COPY --from=builder /root/.local /root/.local
COPY . .

ENV PATH=/root/.local/bin:$PATH
CMD ["aiperf", "--help"]
