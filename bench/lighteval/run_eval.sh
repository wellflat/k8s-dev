#!/bin/sh

#https://github.com/huggingface/lighteval/blob/b1d45e36a6255ebefbfc2aa8999b7b25a007d619/docs/source/use-litellm-as-backend.mdx#
TASK="lighteval|aime25@k=1|0"
TASK="leaderboard|mmlu|0"
lighteval endpoint litellm config.yaml $TASK --output-dir result
