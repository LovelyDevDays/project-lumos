#!/usr/bin/env python3
"""
BM42 Search for Qdrant
Uses fastembed library with BM42 model for sparse vector search
"""

import json
from typing import List, Dict, Any
from fastembed import SparseTextEmbedding
from qdrant_client import QdrantClient
from qdrant_client.models import SparseVector, NamedSparseVector, Query
import argparse


class BM42Searcher:
    def __init__(self, qdrant_host: str = "localhost", qdrant_port: int = 6333):
        """Initialize BM42 searcher with Qdrant client and embedding model"""
        self.client = QdrantClient(host=qdrant_host, port=qdrant_port)
        self.model = SparseTextEmbedding(
            model_name="Qdrant/bm42-all-minilm-l6-v2-attentions"
        )

    def search(self, collection_name: str, query_text: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Search using BM42 sparse embeddings"""
        # Generate query embedding
        query_embedding = list(self.model.embed([query_text]))[0]

        # Create sparse vector for query
        query_vector = SparseVector(
            indices=query_embedding.indices.tolist(),
            values=query_embedding.values.tolist()
        )

        # Search in Qdrant using sparse vector
        from qdrant_client.models import SearchRequest, NamedSparseVector

        results = self.client.search(
            collection_name=collection_name,
            query_vector=NamedSparseVector(
                name="bm42",
                vector=query_vector
            ),
            limit=limit,
            with_payload=True
        )

        # Format results
        formatted_results = []
        for point in results:
            result = {
                "id": point.id,
                "score": point.score,
                "key": point.payload.get("key", ""),
                "title": point.payload.get("title", ""),
                "content": point.payload.get("content", "")
            }
            formatted_results.append(result)

        return formatted_results


def main():
    parser = argparse.ArgumentParser(description="Search using BM42")
    parser.add_argument("--query", "-q", required=True, help="Search query")
    parser.add_argument("--collection", "-c", default="jira_bm42", help="Collection name")
    parser.add_argument("--limit", "-l", type=int, default=10, help="Number of results")
    parser.add_argument("--output", "-o", help="Output JSON file")
    parser.add_argument("--host", default="localhost", help="Qdrant host")
    parser.add_argument("--port", type=int, default=6333, help="Qdrant port")

    args = parser.parse_args()

    # Initialize searcher
    searcher = BM42Searcher(args.host, args.port)

    # Perform search
    results = searcher.search(args.collection, args.query, args.limit)

    # Output results
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            json.dump({
                "query": args.query,
                "count": len(results),
                "results": results
            }, f, ensure_ascii=False, indent=2)
        print(f"Results saved to {args.output}")

    # Display results
    print(f"\nFound {len(results)} results for query: '{args.query}'")
    print("-" * 80)

    for i, result in enumerate(results, 1):
        print(f"{i}. [{result['key']}] {result['title']} (Score: {result['score']:.3f})")
        if result['content']:
            print(f"   {result['content'][:100]}...")
        print()


if __name__ == "__main__":
    main()
