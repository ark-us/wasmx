#!/usr/bin/env python3
"""
Generate text embeddings using OpenAI's text-embedding-3-large model.
Reads text from stdin or command line argument and outputs embedding as JSON array.
"""

import sys
import json
import os

try:
    from openai import OpenAI
except ImportError:
    print(json.dumps({"error": "openai package not installed. Run: pip install openai"}))
    sys.exit(1)


def generate_embedding(text: str, model: str = "text-embedding-3-large") -> list[float]:
    """
    Generate embedding for the given text using OpenAI API.

    Args:
        text: Input text to embed
        model: OpenAI embedding model name (default: text-embedding-3-large)

    Returns:
        List of float values representing the embedding vector
    """
    api_key = os.environ.get("OPENAI_API_KEY")
    if not api_key:
        raise ValueError("OPENAI_API_KEY environment variable not set")

    client = OpenAI(api_key=api_key)

    response = client.embeddings.create(
        model=model,
        input=text,
        encoding_format="float"
    )

    return response.data[0].embedding


def main():
    # Check if text provided as argument or stdin
    if len(sys.argv) > 1:
        text = " ".join(sys.argv[1:])
    else:
        # Read from stdin
        text = sys.stdin.read().strip()

    if not text:
        print(json.dumps({"error": "No input text provided"}))
        sys.exit(1)

    try:
        # Get model from environment or use default
        model = os.environ.get("EMBEDDING_MODEL", "text-embedding-3-large")

        # Generate embedding
        embedding = generate_embedding(text, model)

        # Output as JSON array
        print(json.dumps(embedding))

    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()
