"""
LLM Service Implementation for Python-based LLM Gateway

This service provides RAG-augmented LLM responses with Vietnamese language support,
web search integration, and adaptive query routing.
"""

import asyncio
import logging
import re
from typing import List, Optional, Dict, Any, AsyncGenerator
from dataclasses import dataclass
from enum import Enum

import grpc
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_core.documents import Document
from langchain_core.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_pinecone import PineconeVectorStore
from pinecone import Pinecone
import openai

# Import proto files
import sys
from pathlib import Path
project_root = Path(__file__).parent.parent.parent.parent
proto_path = project_root / "proto"
sys.path.insert(0, str(proto_path))

from llm.v1 import llm_pb2, llm_pb2_grpc
from config import get_config

logger = logging.getLogger(__name__)

class QueryRoute(Enum):
    """Query routing options."""
    VECTORSTORE = "vectorstore"
    WEB_SEARCH = "web_search"
    DIRECT_LLM = "direct_llm"

@dataclass
class RAGState:
    """State management for RAG operations."""
    question: str
    documents: List[Document]
    generation: str = ""
    route: QueryRoute = QueryRoute.VECTORSTORE
    iteration: int = 0
    max_retries: int = 3

class LLMServicer(llm_pb2_grpc.LLMServiceServicer):
    """Python implementation of the LLM service."""
    
    def __init__(self):
        """Initialize the LLM service with all necessary components."""
        self.config = get_config()
        self._initialize_components()
        logger.info("LLM Service initialized successfully")
    
    def _initialize_components(self):
        """Initialize LLM, embeddings, and vector store components."""
        # Initialize OpenAI LLM
        if not self.config.openai_api_key:
            raise ValueError("OPENAI_API_KEY environment variable not set")
        
        self.llm = ChatOpenAI(
            model="gpt-4o",
            temperature=self.config.rag.temperature,
            max_tokens=self.config.rag.max_tokens,
            openai_api_key=self.config.openai_api_key
        )
        
        # Initialize embeddings
        self.embeddings = OpenAIEmbeddings(
            model=self.config.vector_store.embedding_model,
            openai_api_key=self.config.openai_api_key
        )
        
        # Initialize Pinecone
        if self.config.vector_store.pinecone_api_key:
            self.pinecone = Pinecone(api_key=self.config.vector_store.pinecone_api_key)
            self._initialize_vector_store()
        else:
            logger.warning("Pinecone API key not provided, vector search disabled")
            self.vector_store = None
        
        # Initialize web search
        if self.config.rag.web_search_enabled and self.config.rag.web_search_api_key:
            self.web_search = TavilySearchResults(
                api_key=self.config.rag.web_search_api_key,
                max_results=3
            )
        else:
            logger.warning("Web search disabled or API key not provided")
            self.web_search = None
        
        # Initialize text splitter
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=self.config.rag.chunk_size,
            chunk_overlap=self.config.rag.chunk_overlap
        )
    
    def _initialize_vector_store(self):
        """Initialize vector store connection."""
        try:
            # Check if index exists
            index_name = self.config.vector_store.default_index
            
            # Connect to existing index
            index = self.pinecone.Index(index_name)
            
            # Create LangChain Pinecone wrapper
            self.vector_store = PineconeVectorStore(
                index=index,
                embedding=self.embeddings,
                text_key="text"
            )
            
            logger.info(f"Connected to Pinecone index: {index_name}")
            
        except Exception as e:
            logger.error(f"Failed to initialize vector store: {e}")
            self.vector_store = None
    
    def _is_vietnamese_text(self, text: str) -> bool:
        """Check if text contains Vietnamese characters."""
        vietnamese_pattern = r'[àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ]'
        return bool(re.search(vietnamese_pattern, text.lower()))
    
    def _route_query(self, query: str) -> QueryRoute:
        """Determine the best data source for the query."""
        query_lower = query.lower()
        
        # Vietnamese university-specific keywords
        vietnamese_uni_keywords = [
            "đại học", "điểm chuẩn", "tuyển sinh", "ngành học",
            "trường", "khoa", "bách khoa", "kinh tế", "luật", "y khoa",
            "xét tuyển", "học phí", "đề án", "chỉ tiêu"
        ]
        
        # Check for Vietnamese university terms
        for keyword in vietnamese_uni_keywords:
            if keyword in query_lower:
                logger.info(f"Query contains Vietnamese university keyword '{keyword}', routing to vectorstore")
                return QueryRoute.VECTORSTORE
        
        # General education/career keywords
        education_keywords = [
            "university", "college", "admission", "degree", "course",
            "career", "job", "education", "study", "learn"
        ]
        
        for keyword in education_keywords:
            if keyword in query_lower:
                return QueryRoute.VECTORSTORE
        
        # Default to web search for other queries
        return QueryRoute.WEB_SEARCH
    
    async def _retrieve_documents(self, query: str, top_k: int = None) -> List[Document]:
        """Retrieve documents from vector store."""
        if not self.vector_store:
            return []
        
        try:
            top_k = top_k or self.config.rag.retrieval_top_k
            docs = await asyncio.get_event_loop().run_in_executor(
                None, 
                lambda: self.vector_store.similarity_search(query, k=top_k)
            )
            logger.info(f"Retrieved {len(docs)} documents for query")
            return docs
        except Exception as e:
            logger.error(f"Error retrieving documents: {e}")
            return []
    
    async def _web_search_documents(self, query: str) -> List[Document]:
        """Perform web search and return results as documents."""
        if not self.web_search:
            return []
        
        try:
            results = await asyncio.get_event_loop().run_in_executor(
                None,
                lambda: self.web_search.run(query)
            )
            
            documents = []
            for result in results:
                doc = Document(
                    page_content=result.get("content", ""),
                    metadata={
                        "source": result.get("url", "web_search"),
                        "title": result.get("title", ""),
                        "type": "web_search"
                    }
                )
                documents.append(doc)
            
            logger.info(f"Web search returned {len(documents)} documents")
            return documents
            
        except Exception as e:
            logger.error(f"Error in web search: {e}")
            return []
    
    def _grade_documents(self, documents: List[Document], query: str) -> List[Document]:
        """Grade documents for relevance to the query."""
        if not documents:
            return documents
        
        # Simple relevance check based on keyword matching
        relevant_docs = []
        query_words = set(query.lower().split())
        
        for doc in documents:
            content_words = set(doc.page_content.lower().split())
            # Calculate overlap
            overlap = len(query_words.intersection(content_words))
            relevance_score = overlap / len(query_words) if query_words else 0
            
            # Keep documents with at least 10% relevance
            if relevance_score >= 0.1:
                relevant_docs.append(doc)
        
        logger.info(f"Filtered {len(relevant_docs)} relevant documents from {len(documents)}")
        return relevant_docs
    
    def _build_vietnamese_rag_prompt(self, query: str, documents: List[Document]) -> str:
        """Build Vietnamese-specific RAG prompt."""
        if not documents:
            return f"""Bạn là một trợ lý AI chuyên về hướng nghiệp và giáo dục tại Việt Nam.
Tôi không có thông tin cụ thể cho câu hỏi này, vì vậy tôi sẽ cung cấp câu trả lời chung dựa trên kiến thức của mình.

Câu hỏi: {query}

Trả lời:"""
        
        context = ""
        for i, doc in enumerate(documents, 1):
            source = doc.metadata.get("source", "cơ sở dữ liệu")
            context += f"\n[Nguồn {i} - {source}]:\n{doc.page_content}\n"
        
        return f"""Bạn là một trợ lý AI chuyên về hướng nghiệp và giáo dục tại Việt Nam.
Sử dụng thông tin từ các nguồn được cung cấp để trả lời câu hỏi một cách chính xác và hữu ích.
Nếu thông tin không đủ, hãy nói rõ và cung cấp câu trả lời tổng quát.
Giữ câu trả lời súc tích nhưng đầy đủ thông tin.

Thông tin tham khảo:{context}

Câu hỏi: {query}

Trả lời:"""
    
    def _build_english_rag_prompt(self, query: str, documents: List[Document]) -> str:
        """Build English RAG prompt."""
        if not documents:
            return f"""You are an AI assistant helping with career guidance and educational content.
I don't have specific context available for this question, so I'll provide a general response based on my knowledge.

Question: {query}

Answer:"""
        
        context = ""
        for i, doc in enumerate(documents, 1):
            source = doc.metadata.get("source", "knowledge base")
            context += f"\n[Source {i} - {source}]:\n{doc.page_content}\n"
        
        return f"""You are an AI assistant helping with career guidance and educational content.
Use the following retrieved context to answer the question accurately and helpfully.
If the context doesn't contain enough information, say so clearly.
Keep your answer concise but comprehensive.

Context from retrieved sources:{context}

Question: {query}

Answer:"""
    
    async def GenerateStream(self, request, context):
        """Handle basic streaming generation requests."""
        logger.info(f"GenerateStream request: user_id={request.user_id}, prompt='{request.prompt[:100]}...'")
        
        try:
            # Stream response from LLM
            async for chunk in self.llm.astream(request.prompt):
                if hasattr(chunk, 'content'):
                    token = chunk.content
                    if token:
                        yield llm_pb2.GenerateStreamResponse(token=token)
                        
        except Exception as e:
            logger.error(f"Error in GenerateStream: {e}")
            yield llm_pb2.GenerateStreamResponse(token=f"Error: {str(e)}")
    
    async def GenerateWithRAG(self, request, context):
        """Handle RAG-augmented streaming generation requests."""
        logger.info(f"GenerateWithRAG request: user_id={request.user_id}, collection={request.rag_collection}, adaptive={request.adaptive}")
        
        try:
            # Initialize RAG state
            state = RAGState(
                question=request.prompt,
                documents=[],
                max_retries=self.config.rag.max_retries
            )
            
            # Route query
            route = self._route_query(request.prompt)
            state.route = route
            
            # Retrieve documents based on route
            if route == QueryRoute.VECTORSTORE:
                docs = await self._retrieve_documents(request.prompt)
                # Grade documents for relevance
                relevant_docs = self._grade_documents(docs, request.prompt)
                state.documents = relevant_docs
                
                # Fallback to web search if no relevant documents
                if not relevant_docs and self.web_search:
                    logger.info("No relevant documents found, falling back to web search")
                    web_docs = await self._web_search_documents(request.prompt)
                    state.documents = web_docs
                    state.route = QueryRoute.WEB_SEARCH
                    
            elif route == QueryRoute.WEB_SEARCH:
                docs = await self._web_search_documents(request.prompt)
                state.documents = docs
            
            # Build prompt based on language and documents
            is_vietnamese = self._is_vietnamese_text(request.prompt)
            
            if is_vietnamese:
                prompt = self._build_vietnamese_rag_prompt(request.prompt, state.documents)
            else:
                prompt = self._build_english_rag_prompt(request.prompt, state.documents)
            
            logger.info(f"Generated {'Vietnamese' if is_vietnamese else 'English'} RAG prompt with {len(state.documents)} documents")
            
            # Stream response
            async for chunk in self.llm.astream(prompt):
                if hasattr(chunk, 'content'):
                    token = chunk.content
                    if token:
                        yield llm_pb2.GenerateWithRAGResponse(token=token)
                        
        except Exception as e:
            logger.error(f"Error in GenerateWithRAG: {e}")
            yield llm_pb2.GenerateWithRAGResponse(token=f"Error: {str(e)}")
    
    async def IngestDocument(self, request, context):
        """Ingest a document into the vector store."""
        try:
            if not self.vector_store:
                return llm_pb2.IngestDocumentResponse(
                    success=False,
                    message="Vector store not available"
                )
            
            # Split document into chunks
            doc = Document(
                page_content=request.content,
                metadata=dict(request.metadata) if request.metadata else {}
            )
            
            chunks = self.text_splitter.split_documents([doc])
            
            # Add to vector store
            await asyncio.get_event_loop().run_in_executor(
                None,
                lambda: self.vector_store.add_documents(chunks)
            )
            
            logger.info(f"Ingested document with {len(chunks)} chunks")
            
            return llm_pb2.IngestDocumentResponse(
                document_id=request.document_id or "auto_generated",
                success=True,
                message=f"Successfully ingested document with {len(chunks)} chunks",
                chunks_created=len(chunks)
            )
            
        except Exception as e:
            logger.error(f"Error ingesting document: {e}")
            return llm_pb2.IngestDocumentResponse(
                success=False,
                message=f"Error: {str(e)}"
            )
    
    async def CreateCollection(self, request, context):
        """Create a new collection (index)."""
        # For now, return not implemented
        return llm_pb2.CreateCollectionResponse(
            success=False,
            message="Collection creation not implemented in Python version"
        )
    
    async def ListCollections(self, request, context):
        """List available collections."""
        # For now, return the default collection
        collections = []
        if self.vector_store:
            collections.append(llm_pb2.CollectionInfo(
                name=self.config.vector_store.default_index,
                document_count=0,  # Would need to query Pinecone for actual count
                created_at="unknown",
                metadata={}
            ))
        
        return llm_pb2.ListCollectionsResponse(collections=collections)
    
    async def DeleteCollection(self, request, context):
        """Delete a collection."""
        return llm_pb2.DeleteCollectionResponse(
            success=False,
            message="Collection deletion not implemented in Python version"
        )
