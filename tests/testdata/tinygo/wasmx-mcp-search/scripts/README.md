# MCP Search - Embedding Generation Scripts

This directory contains Python scripts for generating text embeddings used by the mcp-search contract.

## Setup

1. Install dependencies (sentence-transformers will download the model on first use):
```bash
cd /Users/user/dev/blockchain/wasmx/tests/testdata/tinygo/wasmx-mcp-search/scripts
pip install -r requirements.txt
```

2. Make sure Python 3 is available in your PATH:
```bash
which python3
```

**No API key required!** The local model runs entirely on your machine.

## Usage

### generate_embedding_local.py (Recommended - No API Key)

Generate embeddings using a local sentence-transformer model.

**From command line:**
```bash
python generate_embedding_local.py "blockchain technology"
```

**From stdin:**
```bash
echo "smart contracts" | python generate_embedding_local.py
```

**With custom model:**
```bash
# Use a different sentence-transformer model
EMBEDDING_MODEL=all-mpnet-base-v2 python generate_embedding_local.py "DeFi"
```

**Available models:**
- `all-MiniLM-L6-v2` (default): 384 dimensions, 80MB, fast and high quality
- `all-mpnet-base-v2`: 768 dimensions, 420MB, highest quality
- `paraphrase-MiniLM-L3-v2`: 384 dimensions, 60MB, fastest
- See [sentence-transformers models](https://www.sbert.net/docs/pretrained_models.html) for more options

### generate_embedding.py (OpenAI API - Requires API Key)

Generate embeddings using OpenAI's API (requires OPENAI_API_KEY).

**Output format:**
The script outputs a JSON array of float values:
```json
[0.123, -0.456, 0.789, ...]
```

**Error handling:**
If an error occurs, the output will be a JSON object with an error field:
```json
{"error": "error message"}
```

## Integration with wasmx-mcp-search

The mcp-search contract calls this script directly using the wasmx-env-core ExecuteCliCommand API to generate embeddings dynamically when users perform searches.

The flow is:
1. User calls `search_knowledge` tool with text query
2. mcp-search contract executes `generate_embedding.py` via ExecuteCliCommand
3. Script generates embedding using OpenAI API
4. mcp-search receives embedding and performs vector search in PostgreSQL
5. Results returned to user

## Models

- **text-embedding-3-large** (default): 3072 dimensions, highest quality
- **text-embedding-3-small**: 1536 dimensions, faster and cheaper

Configure via `EMBEDDING_MODEL` environment variable.

## Important Notes

1. **No API key required** - the contract now uses local sentence-transformer models by default
2. The Python script path is currently hardcoded in tools.go - should be configurable in production
3. The embedding dimension in InitGenesis should match the model:
   - **Local models (sentence-transformers):**
     - all-MiniLM-L6-v2: 384 dimensions (default)
     - all-mpnet-base-v2: 768 dimensions
     - paraphrase-MiniLM-L3-v2: 384 dimensions
   - **OpenAI models (if using generate_embedding.py):**
     - text-embedding-3-large: 3072 dimensions
     - text-embedding-3-small: 1536 dimensions
     - text-embedding-ada-002: 1536 dimensions
4. The model is downloaded on first use (~80MB for all-MiniLM-L6-v2) and cached locally
5. Dependencies are auto-installed when the contract runs (via `pip install -q -r requirements.txt`)
