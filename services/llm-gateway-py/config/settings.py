# LLM Gateway Python Configuration

import os
from dataclasses import dataclass, field
from typing import Optional

@dataclass
class RAGConfig:
    """Configuration for RAG operations."""
    chunk_size: int = 1000
    chunk_overlap: int = 200
    retrieval_top_k: int = 5
    temperature: float = 0.7
    max_tokens: int = 1000
    max_retries: int = 3
    web_search_enabled: bool = True
    web_search_api_key: Optional[str] = None
    web_search_base_url: str = "https://api.tavily.com/search"

@dataclass
class VectorStoreConfig:
    def __init__(self):
        """Initialize with default values."""
        self.pinecone_api_key: Optional[str] = None
        self.pinecone_environment= os.getenv("PINECONE_ENVIRONMENT", "us-east-1")
        self.default_index = os.getenv("PINECONE_INDEX_NAME", "vietnamese-university-rag")
        self.embedding_model = os.getenv("EMBEDDING_MODEL", "text-embedding-3-small")  # 1536 dims
        # dynamically set dimensions based on model
        if self.embedding_model == "llama":
            self.embedding_dimensions = 384
        else:
            self.embedding_dimensions = 1536

@dataclass
class ServiceConfig:
    """Main service configuration."""
    # Service identity
    service_name: str = "llm-gateway-py"
    environment: str = "development"
    version: str = "1.0.0"
    
    # Server configuration
    grpc_port: int = 50054
    http_port: int = 8091
    max_workers: int = 10
    
    # Logging
    log_level: str = "INFO"
    debug: bool = False
    
    # Admin API
    enable_admin_api: bool = True
    admin_api_key: str = "admin-secret-key-change-me"
    
    # External API keys
    openai_api_key: Optional[str] = None
    pinecone_api_key: Optional[str] = None
    tavily_api_key: Optional[str] = None
    
    # RAG and Vector Store configs
    rag: RAGConfig = field(default_factory=RAGConfig)
    vector_store: VectorStoreConfig = field(default_factory=VectorStoreConfig)
    
    def __post_init__(self):
        """Load configuration from environment variables."""
        # Service configuration
        self.service_name = os.getenv("SERVICE_NAME", self.service_name)
        self.environment = os.getenv("ENVIRONMENT", self.environment)
        
        # Server configuration
        self.grpc_port = int(os.getenv("GRPC_PORT", str(self.grpc_port)))
        self.http_port = int(os.getenv("HTTP_PORT", str(self.http_port)))
        self.max_workers = int(os.getenv("MAX_WORKERS", str(self.max_workers)))
        
        # Logging
        self.log_level = os.getenv("LOG_LEVEL", self.log_level)
        self.debug = os.getenv("DEBUG", "false").lower() == "true"
        
        # Admin API
        self.enable_admin_api = os.getenv("ENABLE_ADMIN_API", "true").lower() == "true"
        self.admin_api_key = os.getenv("ADMIN_API_KEY", self.admin_api_key)
        
        # External API keys
        self.openai_api_key = os.getenv("OPENAI_API_KEY")
        self.pinecone_api_key = os.getenv("PINECONE_API_KEY")
        self.tavily_api_key = os.getenv("TAVILY_API_KEY")
        
        # Update nested configurations
        self.rag.web_search_api_key = self.tavily_api_key
        self.rag.web_search_enabled = os.getenv("WEB_SEARCH_ENABLED", "true").lower() == "true"
        
        self.vector_store.pinecone_api_key = self.pinecone_api_key
        self.vector_store.pinecone_environment = os.getenv("PINECONE_ENVIRONMENT", self.vector_store.pinecone_environment)
        self.vector_store.default_index = os.getenv("PINECONE_INDEX", self.vector_store.default_index)
        self.vector_store.embedding_model = os.getenv("EMBEDDING_MODEL", self.vector_store.embedding_model)
        if self.vector_store.embedding_model == "llama":
            self.vector_store.embedding_dimensions = int(os.getenv("EMBEDDING_DIMENSIONS", "384"))
        else:
            self.vector_store.embedding_dimensions = int(os.getenv("EMBEDDING_DIMENSIONS", "1536"))
        self.http_port = int(os.getenv("HTTP_PORT", "8091"))
        self.log_level = os.getenv("LOG_LEVEL", "INFO")
        self.debug = os.getenv("DEBUG", "false").lower() == "true"
        
        # RAG parameters
        self.rag.chunk_size = int(os.getenv("RAG_CHUNK_SIZE", "1000"))
        self.rag.chunk_overlap = int(os.getenv("RAG_CHUNK_OVERLAP", "200"))
        self.rag.retrieval_top_k = int(os.getenv("RAG_TOP_K", "5"))
        self.rag.temperature = float(os.getenv("RAG_TEMPERATURE", "0.7"))
        self.rag.max_tokens = int(os.getenv("RAG_MAX_TOKENS", "1000"))
        self.rag.max_retries = int(os.getenv("RAG_MAX_RETRIES", "3"))

def get_config() -> ServiceConfig:
    """Get the service configuration."""
    return ServiceConfig()

def get_settings() -> ServiceConfig:
    """Get the service configuration (alias for get_config)."""
    return get_config()
