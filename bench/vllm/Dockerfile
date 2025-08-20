FROM vllm/vllm-openai:gptoss

RUN groupadd -g 1000 vllm_user && \
    useradd -m -s /bin/bash -u 1000 -g 1000 vllm_user
WORKDIR /workspace
RUN chown -R 1000:1000 /workspace
USER vllm_user