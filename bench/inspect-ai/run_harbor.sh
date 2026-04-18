#!/bin/sh

#TASK=terminal_bench_2_0
TASK=compilebench_1_0

inspect eval inspect_harbor/${TASK} \
    --sandbox docker \
    --model openai/gpt-5.4-mini