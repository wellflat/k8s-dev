#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import os
import sys

from openai import OpenAI


DEFAULT_BASE_URL = "http://127.0.0.1:8080/v1"
DEFAULT_MODEL = "nemotron-3-nano-30b-a3b-bf16"
DEFAULT_PROMPT = "Kubernetes と llm-d の疎通確認です。短く応答してください。"


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Send a test request to an OpenAI-compatible completion API."
    )
    parser.add_argument(
        "--base-url",
        default=os.getenv("LLM_D_BASE_URL")
        or os.getenv("OPENAI_BASE_URL")
        or DEFAULT_BASE_URL,
        help="OpenAI-compatible base URL including /v1.",
    )
    parser.add_argument(
        "--api-key",
        default=os.getenv("LLM_D_API_KEY")
        or os.getenv("OPENAI_API_KEY")
        or "dummy",
        help="API key for the gateway. Defaults to a dummy value.",
    )
    parser.add_argument(
        "--model",
        default=os.getenv("LLM_D_MODEL") or DEFAULT_MODEL,
        help="Model name exposed by llm-d.",
    )
    parser.add_argument(
        "--prompt",
        default=DEFAULT_PROMPT,
        help="Prompt text sent to the completion API.",
    )
    parser.add_argument(
        "--max-tokens",
        type=int,
        default=128,
        help="Maximum number of tokens to generate.",
    )
    parser.add_argument(
        "--temperature",
        type=float,
        default=0.2,
        help="Sampling temperature.",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=120.0,
        help="Request timeout in seconds.",
    )
    parser.add_argument(
        "--raw",
        action="store_true",
        help="Print the full JSON response instead of only the completion text.",
    )
    return parser


def main() -> int:
    args = build_parser().parse_args()
    client = OpenAI(
        api_key=args.api_key,
        base_url=args.base_url,
        timeout=args.timeout,
    )

    try:
        response = client.completions.create(
            model=args.model,
            prompt=args.prompt,
            max_tokens=args.max_tokens,
            temperature=args.temperature,
        )
    except Exception as exc:
        print(f"request failed: {exc}", file=sys.stderr)
        return 1

    payload = response.model_dump()
    if args.raw:
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 0

    choices = payload.get("choices") or []
    if not choices:
        print("no choices in response", file=sys.stderr)
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 1

    text = choices[0].get("text", "")
    print(text.strip())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
