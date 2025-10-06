#!/bin/sh

#https://github.com/huggingface/lighteval/blob/b1d45e36a6255ebefbfc2aa8999b7b25a007d619/docs/source/use-litellm-as-backend.mdx#
lighteval endpoint litellm config.yaml "lighteval|aime25@k=1|0" --output-dir result
