#!/bin/sh

STACK_NAME=BenchmarkETLStack
PROFILE=wellflat

# 1. アプリケーションをビルドします
sam build

# 2. ビルドしたアプリケーションをデプロイします
sam deploy \
    --stack-name ${STACK_NAME} \
    --profile ${PROFILE} \
    --capabilities CAPABILITY_IAM \
    --resolve-s3