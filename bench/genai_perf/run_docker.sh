#!/bin/sh

IMAGE=nvcr.io/nvidia/tritonserver:25.07-py3-sdk
VOLUME=./profiles:/workspace/profiles
#docker run --rm -it --net=host -v ${VOLUME} --gpus=all ${IMAGE} /bin/bash
docker run --rm -it --net=host -v ${VOLUME} ${IMAGE} /bin/bash
