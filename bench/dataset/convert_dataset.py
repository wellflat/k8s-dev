#!/usr/bin/env python

import argparse
import json
import logging
import os

from datasets import Dataset, load_dataset

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="download dataset from Hugging Face Hub")
    parser.add_argument("dataset", type=str, help="dataset name")
    parser.add_argument("--subset", type=str, default="", help="dataset subset name")
    parser.add_argument("--split", type=str, default="train", help="dataset split name (default: train)")
    parser.add_argument("--output_dir", type=str, default=None, help="dataset cache directory")
    parser.add_argument("--columns", nargs="+", required=True, help="column names to include in the output text")
    parser.add_argument("--output_file", type=str, required=True, help="path to the output JSONL file")
    return parser.parse_args()


def download_huggingface_dataset(
    dataset_name: str,
    subset_name: str | None = None,
    split_name: str | None = "train",
    output_dir: str | None = None
) -> Dataset | None:
    try:
        logging.info(f"'{dataset_name}' start downloading...")
        dataset = load_dataset(dataset_name, name=subset_name, split=split_name, cache_dir=output_dir)
        logging.info(f"'{dataset_name}' download complete.")
        logging.info("dataset info:")
        print(dataset)

        if output_dir:
            logging.info(f"dataset saved to'{os.path.abspath(output_dir)}'")
        else:
            logging.info("dataset saved to default cache directory")
        
        return dataset

    except Exception as e:
        logging.error(f"dataset download error: {e}")

def convert_jsonl(dataset: Dataset, columns: list[str], output_file: str):
    """
    Converts a Hugging Face Dataset to a JSONL file.

    Args:
        dataset: The Hugging Face Dataset object.
        columns: A list of column names to concatenate into the 'text' field.
        output_file: The path to save the output JSONL file.
    """
    logging.info(f"Converting dataset to JSONL format, saving to '{output_file}'...")
    try:
        with open(output_file, "w", encoding="utf-8") as f:
            for item in dataset:
                # 指定されたカラムの値を連結
                text_content = " ".join(str(item[col]) for col in columns if col in item and item[col] is not None)
                json_line = json.dumps({"text": text_content}, ensure_ascii=False)
                f.write(json_line + "\n")
        logging.info(f"Successfully converted and saved to '{os.path.abspath(output_file)}'")
    except Exception as e:
        logging.error(f"Failed to convert dataset to JSONL: {e}")


def main():
    args = parse_arguments()
    dataset = download_huggingface_dataset(args.dataset, args.subset, args.split, args.output_dir)
    if dataset:
        convert_jsonl(dataset, args.columns, args.output_file)

if __name__ == "__main__":
    main()
