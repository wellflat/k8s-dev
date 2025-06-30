#!/bin/sh

IMAGE=nvcr.io/nvidia/tritonserver:25.01-py3-sdk
docker run --rm -it --net=host --gpu=all ${IMAGE} /bin/bash