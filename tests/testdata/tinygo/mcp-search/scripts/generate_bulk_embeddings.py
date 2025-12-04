#!/usr/bin/env python3
"""
Generate embeddings for multiple items in bulk and insert directly into PostgreSQL.
Reads JSON array of items from stdin or uses mock data.
Reads database connection info from stdin as first line.

Input format:
First line (connection info):
{"connection_string": "postgresql://...", "database_name": "mcp_search_test", "dimension": 384}

Following lines (items):
[
    {"key": "item1", "value": "text to embed"},
    {"key": "item2", "value": "more text"}
]

Output format:
{"success": true, "items_stored": 5}
"""

import sys
import json
import os

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print(json.dumps({"error": "sentence-transformers package not installed. Run: pip install sentence-transformers"}), file=sys.stderr)
    sys.exit(1)

try:
    import psycopg2
    from pgvector.psycopg2 import register_vector
except ImportError:
    print(json.dumps({"error": "psycopg2 and pgvector packages not installed. Run: pip install psycopg2-binary pgvector"}), file=sys.stderr)
    sys.exit(1)


# Mock data - in production this would come from PostgreSQL
MOCK_DATA = [
    {
        "key": "blockchain_intro",
        "value": "Blockchain is a distributed ledger technology that enables secure and transparent record-keeping."
    },
    {
        "key": "smart_contracts",
        "value": "Smart contracts are self-executing contracts with the terms directly written into code."
    },
    {
        "key": "consensus_mechanisms",
        "value": "Consensus mechanisms like Proof of Work and Proof of Stake ensure agreement in distributed networks."
    },
    {
        "key": "cryptography",
        "value": "Cryptographic techniques secure blockchain transactions and ensure data integrity."
    },
    {
        "key": "defi",
        "value": "Decentralized Finance (DeFi) enables financial services without traditional intermediaries."
    }
]


def insert_embeddings_to_db(items: list[dict], connection_string: str, database_name: str, dimension: int, model_name: str = None) -> dict:
    """
    Generate embeddings and insert directly into PostgreSQL.

    Args:
        items: List of dicts with 'key' and 'value' fields
        connection_string: PostgreSQL connection string
        database_name: Database name
        dimension: Embedding dimension
        model_name: Optional model name (default: all-MiniLM-L6-v2)

    Returns:
        Dict with success status and items_stored count
    """
    if model_name is None:
        model_name = os.environ.get("EMBEDDING_MODEL", "all-MiniLM-L6-v2")

    # Load model
    model = SentenceTransformer(model_name)

    # Extract texts to embed
    texts = [item["value"] for item in items]

    # Generate embeddings in batch (more efficient)
    embeddings = model.encode(texts, convert_to_numpy=True, show_progress_bar=False)

    # First, connect to postgres database to create target database if needed
    conn = psycopg2.connect(connection_string)
    conn.autocommit = True  # Set autocommit before any queries
    cursor = conn.cursor()

    # Create database if it doesn't exist
    cursor.execute(f"SELECT 1 FROM pg_database WHERE datname = '{database_name}'")
    if not cursor.fetchone():
        cursor.execute(f"CREATE DATABASE {database_name}")

    # Close connection to postgres
    cursor.close()
    conn.close()

    # Now connect to the target database
    # Parse connection string and replace database name
    conn = psycopg2.connect(connection_string.rsplit('/', 1)[0] + '/' + database_name)
    cursor = conn.cursor()

    # Enable pgvector extension first
    cursor.execute("CREATE EXTENSION IF NOT EXISTS vector")
    conn.commit()

    # Now register vector type after extension is created
    register_vector(conn)

    # Create table if it doesn't exist
    cursor.execute(f"""
        CREATE TABLE IF NOT EXISTS state_embeddings (
            key TEXT PRIMARY KEY,
            value TEXT,
            embedding vector({dimension})
        )
    """)

    # Insert embeddings
    items_stored = 0
    for item, embedding in zip(items, embeddings):
        try:
            cursor.execute(
                """
                INSERT INTO state_embeddings (key, value, embedding)
                VALUES (%s, %s, %s)
                ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, embedding = EXCLUDED.embedding
                """,
                (item["key"], item["value"], embedding.tolist())
            )
            items_stored += 1
        except Exception as e:
            print(f"Error inserting {item['key']}: {e}", file=sys.stderr)

    conn.commit()
    cursor.close()
    conn.close()

    return {
        "success": items_stored > 0,
        "items_stored": items_stored
    }


def main():
    # Read stdin
    stdin_lines = sys.stdin.read().strip().split('\n', 1)

    # Parse connection info from first line
    try:
        config = json.loads(stdin_lines[0])
        connection_string = config.get("connection_string", "postgresql://localhost:5432/postgres")
        database_name = config.get("database_name", "mcp_search_test")
        dimension = config.get("dimension", 384)
    except (json.JSONDecodeError, IndexError):
        # No config provided, use defaults
        connection_string = os.environ.get("DATABASE_URL", "postgresql://localhost:5432/postgres")
        database_name = "mcp_search_test"
        dimension = 384

    # Parse items from second part or use mock data
    items = MOCK_DATA
    if len(stdin_lines) > 1 and stdin_lines[1]:
        try:
            items = json.loads(stdin_lines[1])
        except json.JSONDecodeError:
            pass  # Use mock data

    # Validate items
    if not isinstance(items, list):
        print(json.dumps({"error": "Items must be a JSON array"}))
        sys.exit(1)

    for item in items:
        if not isinstance(item, dict) or "key" not in item or "value" not in item:
            print(json.dumps({"error": "Each item must have 'key' and 'value' fields"}))
            sys.exit(1)

    try:
        # Get model from environment or use default
        model_name = os.environ.get("EMBEDDING_MODEL", "all-MiniLM-L6-v2")

        # Generate embeddings and insert into database
        result = insert_embeddings_to_db(items, connection_string, database_name, dimension, model_name)

        # Output result as JSON
        print(json.dumps(result))

    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()
