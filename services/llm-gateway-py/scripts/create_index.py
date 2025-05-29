import os
from pinecone import Pinecone, ServerlessSpec

def create_pinecone_index():
    """Create a new Pinecone index with correct dimensions for OpenAI embeddings."""
    
    # Initialize Pinecone
    pc = Pinecone(api_key=os.getenv("PINECONE_API_KEY"))
    
    index_name = "vietnamese-university-rag-1536"  # New index name
    
    # Delete existing index if it exists
    if index_name in pc.list_indexes().names():
        print(f"Deleting existing index: {index_name}")
        pc.delete_index(index_name)
    
    # Create new index with 1536 dimensions (for OpenAI text-embedding-3-small)
    print(f"Creating new index: {index_name}")
    pc.create_index(
        name=index_name,
        dimension=1536,  # Match OpenAI embedding dimension
        metric="cosine",
        spec=ServerlessSpec(
            cloud="aws",
            region="us-east-1"
        )
    )
    
    print(f"✅ Index {index_name} created successfully!")
    print(f"Update your .env file with: PINECONE_INDEX_NAME={index_name}")

if __name__ == "__main__":
    create_pinecone_index()