FROM vllm/vllm-openai:gptoss

RUN groupadd -g 1001 vllm_user && \
    useradd -m -s /bin/bash -u 1001 -g 1001 vllm_user
WORKDIR /workspace
RUN chown -R 1001:1001 /workspace
USER vllm_user