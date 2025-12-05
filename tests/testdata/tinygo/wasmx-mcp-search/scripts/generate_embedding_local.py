#!/usr/bin/env python3
"""
Generate text embeddings using a local sentence-transformer model (no API key required).
Reads text from stdin or command line argument and outputs embedding as JSON array.
"""

import sys
import json
import os

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print(json.dumps({"error": "sentence-transformers package not installed. Run: pip install sentence-transformers"}))
    sys.exit(1)


# Global model instance (loaded once and cached)
_model = None


def get_model(model_name: str = None):
    """
    Get or initialize the embedding model.
    Model is loaded once and cached for subsequent calls.

    Args:
        model_name: Name of the sentence-transformers model (default: all-MiniLM-L6-v2)

    Returns:
        SentenceTransformer model instance
    """
    global _model

    if _model is None:
        if model_name is None:
            # Default to all-mpnet-base-v2: 768 dimensions, 420MB, better quality
            # Alternative: all-MiniLM-L6-v2 (384 dimensions, 80MB, faster)
            model_name = os.environ.get("EMBEDDING_MODEL", "sentence-transformers/all-mpnet-base-v2")

        # Don't print to stderr in production - it may cause issues with wasmx
        # print(f"Loading model: {model_name}...", file=sys.stderr)
        _model = SentenceTransformer(model_name)
        # print(f"Model loaded successfully", file=sys.stderr)

    return _model


def generate_embedding(text: str, model_name: str = None) -> list[float]:
    """
    Generate embedding for the given text using sentence-transformers.

    Args:
        text: Input text to embed
        model_name: Optional model name to use

    Returns:
        List of float values representing the embedding vector
    """
    model = get_model(model_name)

    # Generate embedding
    embedding = model.encode(text, convert_to_numpy=True)

    # Convert numpy array to list
    return embedding.tolist()


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
        model_name = os.environ.get("EMBEDDING_MODEL", "sentence-transformers/all-mpnet-base-v2")

        # Generate embedding
        embedding = generate_embedding(text, model_name)

        # Output as JSON array
        print(json.dumps(embedding))

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
