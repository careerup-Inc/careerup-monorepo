"""
LLM Service Implementation for Python-based LLM Gateway

This service provides adaptive RAG-augmented LLM responses with Vietnamese language support,
web search integration, document grading, hallucination detection, and multi-representation indexing.
Based on LangGraph adaptive RAG patterns.
"""

import asyncio
import json
import logging
import re
import uuid
from typing import List, Optional, Dict, Any, AsyncGenerator, Literal
from dataclasses import dataclass
from enum import Enum

import grpc
from pydantic import BaseModel, Field
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_community.document_loaders import PyPDFLoader
from langchain_core.documents import Document
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.messages import SystemMessage, HumanMessage
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
    """Query routing options for adaptive RAG."""
    VECTORSTORE = "vectorstore"
    WEB_SEARCH = "web_search"
    DIRECT_LLM = "direct_llm"

@dataclass
class RAGState:
    """State management for adaptive RAG operations."""
    question: str
    documents: List[Document]
    generation: str = ""
    route: QueryRoute = QueryRoute.VECTORSTORE
    iteration: int = 0
    max_retries: int = 3
    relevance_scores: List[str] = None
    hallucination_score: str = ""

# Pydantic models for structured LLM outputs
class GradeDocuments(BaseModel):
    """Binary score for relevance check on retrieved documents."""
    binary_score: str = Field(
        description="Documents are relevant to the question, 'yes' or 'no'"
    )

class GradeHallucinations(BaseModel):
    """Binary score for hallucination present in generation answer."""
    binary_score: str = Field(
        description="Answer is grounded in the facts, 'yes' or 'no'"
    )

class RouteQuery(BaseModel):
    """Route a user query to the most relevant datasource."""
    datasource: Literal["vectorstore", "web_search"] = Field(
        ...,
        description="Given a user question choose to route it to web search or a vectorstore.",
    )

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
        
        # Initialize embeddings based on the configured model
        embedding_model = self.config.vector_store.embedding_model
        
        if embedding_model == "llama" or embedding_model.startswith("sentence-transformers"):
            # Use HuggingFace sentence-transformers for llama or similar models
            from langchain_huggingface import HuggingFaceEmbeddings
            
            # Map llama to a specific sentence-transformers model
            if embedding_model == "llama":
                model_name = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
            else:
                model_name = embedding_model
                
            self.embeddings = HuggingFaceEmbeddings(
                model_name=model_name,
                model_kwargs={'device': 'cpu'},  # Use CPU for compatibility
                encode_kwargs={'normalize_embeddings': True}
            )
            logger.info(f"Initialized HuggingFace embeddings with model: {model_name}")
        else:
            # Default to OpenAI embeddings
            self.embeddings = OpenAIEmbeddings(
                model=embedding_model,
                openai_api_key=self.config.openai_api_key
            )
            logger.info(f"Initialized OpenAI embeddings with model: {embedding_model}")
        
        # Initialize Pinecone
        if hasattr(self.config, 'pinecone_api_key') and self.config.pinecone_api_key:
            self.pinecone = Pinecone(api_key=self.config.pinecone_api_key)
            self._initialize_vector_store()
            self._initialize_vietnamese_vector_store()
        else:
            logger.warning("Pinecone API key not provided, vector search disabled")
            self.vector_store = None
            self.vietnamese_vector_store = None
        
        # Initialize web search
        if self.config.rag.web_search_enabled and self.config.rag.web_search_api_key:
            self.web_search = TavilySearchResults(
                api_key=self.config.rag.web_search_api_key,
                max_results=5
            )
        else:
            logger.warning("Web search disabled or API key not provided")
            self.web_search = None
        
        # Initialize text splitter
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=self.config.rag.chunk_size,
            chunk_overlap=self.config.rag.chunk_overlap
        )
        
        # Initialize adaptive RAG components
        self._initialize_adaptive_rag_components()
    
    def _initialize_vector_store(self):
        """Initialize vector store connection."""
        try:
            # Check if index exists
            index_name = self.config.vector_store.default_index

            if not index_name:
                raise ValueError("No index name specified in configuration")

            # Connect to existing index
            index = self.pinecone.Index(name=index_name)
            
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

    def _initialize_vietnamese_vector_store(self):
        """Initialize Vietnamese vector store with Llama embeddings and correct index."""
        try:
            # Use the Vietnamese index name from config
            # vietnamese_index_name = getattr(self.config.vector_store, 'vietnamese_index', 'vietnamese-university-data')

            # For now, use the same index as the main one since we only have university-scores
            # TODO: Create separate Vietnamese index when needed
            index_name = self.config.vector_store.default_index
            
            if not index_name:
                raise ValueError("No Vietnamese index name specified")
        
            # Connect to Vietnamese index with explicit name
            vietnamese_index = self.pinecone.Index(name=index_name)
            
            # Create LangChain Pinecone wrapper with Llama embeddings (384 dimensions)
            self.vietnamese_vector_store = PineconeVectorStore(
                index=vietnamese_index,
                embedding=self.embeddings,  # Use Llama embeddings
                text_key="text"
            )
            
            logger.info(f"Connected to Vietnamese Pinecone index: {vietnamese_index}")
            
        except Exception as e:
            logger.error(f"Failed to initialize Vietnamese vector store: {e}")
            self.vietnamese_vector_store = None
    
    def _initialize_adaptive_rag_components(self):
        """Initialize adaptive RAG components including structured LLM graders."""
        # Initialize structured LLM graders for document relevance
        self.grade_documents_llm = self.llm.with_structured_output(GradeDocuments)
        self.grade_documents_system_prompt = """You are a grader assessing relevance of a retrieved document to a user question about Vietnamese university admissions, scores, and education. 
If the document contains keyword(s) or semantic meaning related to the user question, grade it as relevant. 
For Vietnamese university queries, prioritize documents containing university names, admission scores, program names, or educational information.
It does not need to be a stringent test. The goal is to filter out erroneous retrievals. 
Give a binary score 'yes' or 'no' score to indicate whether the document is relevant to the question."""
        
        # Initialize hallucination grader
        self.grade_hallucinations_llm = self.llm.with_structured_output(GradeHallucinations)
        self.grade_hallucinations_system_prompt = """You are a grader assessing whether an LLM generation is grounded in / supported by a set of retrieved facts about Vietnamese universities and education.
Give a binary score 'yes' or 'no'. 'Yes' means that the answer is grounded in / supported by the set of facts.
For Vietnamese university information, ensure the response uses actual data from the provided documents."""
        
        # Initialize query router
        self.router_llm = self.llm.with_structured_output(RouteQuery)
        self.router_system_prompt = """You are an expert at routing a user question to a vectorstore or web search.
The vectorstore contains documents related to Vietnamese university admissions, admission scores (điểm chuẩn), programs, and educational information.
Use the vectorstore for questions about:
- Vietnamese universities and colleges (đại học, cao đẳng)
- Admission scores and requirements (điểm chuẩn, yêu cầu tuyển sinh)
- Academic programs and majors (ngành học, chương trình đào tạo)
- Educational institutions in Vietnam
- Career guidance related to Vietnamese education
Otherwise, use web-search for general questions, current events, or non-educational topics."""
        
        # Vietnamese-specific RAG prompts
        self.vietnamese_rag_prompt = """Bạn là một trợ lý tư vấn giáo dục chuyên về hệ thống đại học Việt Nam.

Sử dụng thông tin từ các tài liệu được cung cấp để trả lời câu hỏi của người dùng.

Nếu bạn không biết câu trả lời dựa trên tài liệu, hãy nói rõ là bạn không có thông tin đủ.

Đối với thông tin về điểm chuẩn và tuyển sinh, hãy cung cấp:
- Tên trường và ngành học chính xác
- Điểm chuẩn cụ thể
- Tổ hợp môn thi
- Phương thức xét tuyển
- Năm áp dụng

Câu hỏi: {question}

Thông tin tham khảo: {context}

Trả lời:"""
        
        self.english_rag_prompt = """You are an assistant for question-answering tasks about Vietnamese university education system.

Use the following pieces of retrieved context to answer the question about Vietnamese universities, admission scores, and educational programs.

If you don't know the answer based on the provided context, say that you don't have sufficient information.

For admission score and university information, provide:
- Exact university and program names
- Specific admission scores
- Subject combinations required
- Admission methods
- Applicable year

Question: {question}

Context: {context}

Answer:"""
    
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
        """Grade documents for relevance using structured LLM grader."""
        if not documents:
            return documents
            
        logger.info(f"---CHECK DOCUMENT RELEVANCE TO QUESTION---")
        filtered_docs = []
        
        for doc in documents:
            try:
                grade_prompt = f"Here is the retrieved document: \n\n {doc.page_content} \n\n Here is the user question: \n\n {query}"
                
                messages = [
                    SystemMessage(content=self.grade_documents_system_prompt),
                    HumanMessage(content=grade_prompt)
                ]
                
                score = self.grade_documents_llm.invoke(messages)
                grade = score.binary_score
                
                if grade == "yes":
                    logger.info("---GRADE: DOCUMENT RELEVANT---")
                    filtered_docs.append(doc)
                else:
                    logger.info("---GRADE: DOCUMENT NOT RELEVANT---")
                    continue
                    
            except Exception as e:
                logger.error(f"Error grading document: {e}")
                # On error, include the document to be safe
                filtered_docs.append(doc)
        
        logger.info(f"Filtered {len(filtered_docs)} relevant documents from {len(documents)}")
        return filtered_docs
    
    def _grade_hallucinations(self, generation: str, documents: List[Document]) -> bool:
        """Grade generation for hallucinations using structured LLM grader."""
        if not documents:
            return True  # If no documents, can't check hallucinations
            
        logger.info("---CHECK HALLUCINATIONS---")
        
        try:
            formatted_docs = "\n\n".join(doc.page_content for doc in documents)
            grade_prompt = f"Set of facts: \n\n {formatted_docs} \n\n LLM generation: {generation}"
            
            messages = [
                SystemMessage(content=self.grade_hallucinations_system_prompt),
                HumanMessage(content=grade_prompt)
            ]
            
            score = self.grade_hallucinations_llm.invoke(messages)
            grade = score.binary_score
            
            if grade == "yes":
                logger.info("---DECISION: GENERATION IS GROUNDED IN DOCUMENTS---")
                return True
            else:
                logger.info("---DECISION: GENERATION IS NOT GROUNDED IN DOCUMENTS---")
                return False
                
        except Exception as e:
            logger.error(f"Error grading hallucinations: {e}")
            return True  # On error, assume no hallucination to be safe
    
    def _route_query_with_llm(self, query: str) -> QueryRoute:
        """Route query using structured LLM router."""
        try:
            logger.info("---ROUTE QUESTION---")
            
            messages = [
                SystemMessage(content=self.router_system_prompt),
                HumanMessage(content=query)
            ]
            
            source = self.router_llm.invoke(messages)
            
            if source.datasource == "web_search":
                logger.info("---ROUTE QUESTION TO WEB SEARCH---")
                return QueryRoute.WEB_SEARCH
            elif source.datasource == "vectorstore":
                logger.info("---ROUTE QUESTION TO RAG---")
                return QueryRoute.VECTORSTORE
            else:
                return QueryRoute.VECTORSTORE  # Default fallback
                
        except Exception as e:
            logger.error(f"Error routing query: {e}")
            return self._route_query(query)  # Fallback to keyword-based routing
    
    def _build_vietnamese_rag_prompt(self, query: str, documents: List[Document]) -> str:
        """Build Vietnamese-specific RAG prompt using enhanced template."""
        if not documents:
            return f"""Bạn là một trợ lý AI chuyên về hướng nghiệp và giáo dục tại Việt Nam.
Tôi không có thông tin cụ thể cho câu hỏi này, vì vậy tôi sẽ cung cấp câu trả lời chung dựa trên kiến thức của mình.

Câu hỏi: {query}

Trả lời:"""
        
        context = ""
        for i, doc in enumerate(documents, 1):
            source = doc.metadata.get("source", "cơ sở dữ liệu")
            context += f"\n[Nguồn {i} - {source}]:\n{doc.page_content}\n"
        
        return self.vietnamese_rag_prompt.format(question=query, context=context)
    
    def _build_english_rag_prompt(self, query: str, documents: List[Document]) -> str:
        """Build English RAG prompt using enhanced template."""
        if not documents:
            return f"""You are an AI assistant specializing in career guidance and education in Vietnam.
I don't have specific information for this question, so I'll provide a general answer based on my knowledge.

Question: {query}

Answer:"""
        
        context = ""
        for i, doc in enumerate(documents, 1):
            source = doc.metadata.get("source", "database")
            context += f"\n[Source {i} - {source}]:\n{doc.page_content}\n"
        
        return self.english_rag_prompt.format(question=query, context=context)
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
        """Handle adaptive RAG-augmented streaming generation requests."""
        logger.info(f"GenerateWithRAG request: user_id={request.user_id}, collection={request.rag_collection}, adaptive={request.adaptive}")
        
        try:
            # Initialize RAG state
            state = RAGState(
                question=request.prompt,
                documents=[],
                max_retries=self.config.rag.max_retries
            )
            
            # Use adaptive routing if enabled
            if request.adaptive:
                route = self._route_query_with_llm(request.prompt)
            else:
                route = self._route_query(request.prompt)
            state.route = route
            
            # Retrieve documents based on route
            if route == QueryRoute.VECTORSTORE:
                docs = await self._retrieve_documents(request.prompt)
                # Grade documents for relevance using LLM grader
                relevant_docs = self._grade_documents(docs, request.prompt)
                state.documents = relevant_docs
                
                # Fallback to web search if no relevant documents and adaptive mode
                if not relevant_docs and self.web_search and request.adaptive:
                    logger.info("No relevant documents found, falling back to web search")
                    web_docs = await self._web_search_documents(request.prompt)
                    state.documents = web_docs
                    state.route = QueryRoute.WEB_SEARCH
                    
            elif route == QueryRoute.WEB_SEARCH:
                docs = await self._web_search_documents(request.prompt)
                state.documents = docs
            
            # Generate response with retry logic for hallucination checking
            for attempt in range(state.max_retries):
                state.iteration = attempt + 1
                
                # Build prompt based on language and documents
                is_vietnamese = self._is_vietnamese_text(request.prompt)
                
                if is_vietnamese:
                    prompt = self._build_vietnamese_rag_prompt(request.prompt, state.documents)
                else:
                    prompt = self._build_english_rag_prompt(request.prompt, state.documents)
                
                logger.info(f"Generating {'Vietnamese' if is_vietnamese else 'English'} RAG response (attempt {attempt + 1}) with {len(state.documents)} documents")
                
                # Generate response
                full_response = ""
                async for chunk in self.llm.astream(prompt):
                    if hasattr(chunk, 'content'):
                        token = chunk.content
                        if token:
                            full_response += token
                            # Stream tokens in real-time only on final attempt or if not checking hallucinations
                            if not request.adaptive or attempt == state.max_retries - 1:
                                yield llm_pb2.GenerateWithRAGResponse(token=token)
                
                state.generation = full_response
                
                # Check for hallucinations if adaptive mode and we have documents
                if request.adaptive and state.documents and attempt < state.max_retries - 1:
                    is_grounded = self._grade_hallucinations(full_response, state.documents)
                    if is_grounded:
                        logger.info("Generation is grounded, streaming response")
                        # If we didn't stream earlier, stream now
                        if attempt < state.max_retries - 1:
                            for token in full_response:
                                yield llm_pb2.GenerateWithRAGResponse(token=token)
                        break
                    else:
                        logger.info(f"Generation not grounded, retrying (attempt {attempt + 1})")
                        continue
                else:
                    # No hallucination checking or final attempt
                    break
                        
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
    
    async def clear_collection(self, collection_name: str) -> bool:
        """Clear all data from a specific Pinecone collection/index.
        
        Args:
            collection_name: Name of the collection/index to clear
            
        Returns:
            bool: True if successful, False otherwise
        """
        try:
            if not self.pinecone:
                logger.error("Pinecone client not initialized")
                return False
            
            # Get the index
            index = self.pinecone.Index(collection_name)
            
            # Delete all vectors from the index by querying and deleting all namespaces
            # First, try to delete all vectors in the default namespace
            try:
                # Get index stats to see if it has vectors
                stats = index.describe_index_stats()
                total_vector_count = stats.get('total_vector_count', 0)
                
                if total_vector_count > 0:
                    # Delete all vectors in the default namespace
                    index.delete(delete_all=True)
                    logger.info(f"Cleared {total_vector_count} vectors from index '{collection_name}'")
                else:
                    logger.info(f"Index '{collection_name}' is already empty")
                    
                return True
                
            except Exception as delete_error:
                logger.error(f"Error deleting vectors from index '{collection_name}': {delete_error}")
                return False
                
        except Exception as e:
            logger.error(f"Error clearing collection '{collection_name}': {e}")
            return False
    
    async def list_collections(self) -> List[Dict[str, Any]]:
        """List all available Pinecone collections/indexes.
        
        Returns:
            List[Dict]: List of collection information
        """
        try:
            if not self.pinecone:
                logger.error("Pinecone client not initialized")
                return []
            
            # List all indexes
            indexes = self.pinecone.list_indexes()
            collections = []
            
            for index_info in indexes:
                # Get index stats
                try:
                    index_name = index_info.name
                    index = self.pinecone.Index(index_name)
                    stats = index.describe_index_stats()
                    
                    collection = {
                        "name": index_name,
                        "dimension": int(index_info.dimension),
                        "metric": str(index_info.metric),
                        "document_count": int(stats.get('total_vector_count', 0)),
                        "host": str(index_info.host),
                        "status": str(index_info.status.state) if index_info.status else "unknown",
                        "created_at": "unknown",  # Pinecone doesn't provide creation time via API
                        "metadata": {
                            "index_fullness": float(stats.get('index_fullness', 0.0)),
                            "namespace_count": len(stats.get('namespaces', {}))
                        }
                    }
                    collections.append(collection)
                    
                except Exception as index_error:
                    logger.error(f"Error getting stats for index '{index_info.name}': {index_error}")
                    # Add basic info even if stats fail
                    collections.append({
                        "name": str(index_info.name),
                        "dimension": int(index_info.dimension),
                        "metric": str(index_info.metric),
                        "document_count": 0,
                        "host": str(index_info.host),
                        "status": "error",
                        "created_at": "unknown",
                        "metadata": {"error": str(index_error)}
                    })
            
            logger.info(f"Found {len(collections)} Pinecone collections")
            return collections
            
        except Exception as e:
            logger.error(f"Error listing collections: {e}")
            return []
    
    async def ingest_vietnamese_university_data(
        self, 
        file_path: str, 
        file_type: str = "auto", 
        collection_name: str = "vietnamese-university-data"
    ) -> Dict[str, Any]:
        """
        Ingest Vietnamese university data from JSON or PDF files.
        
        Args:
            file_path: Path to the data file
            file_type: Type of file ('json', 'pdf', or 'auto')
            collection_name: Target collection name
            
        Returns:
            Dictionary containing ingestion results
        """
        try:
            logger.info(f"Starting Vietnamese university data ingestion from {file_path}")
            
            # Auto-detect file type if needed
            if file_type == "auto":
                file_type = "json" if file_path.lower().endswith('.json') else "pdf"
            
            documents_processed = 0
            summaries_created = 0
            total_chunks = 0
            
            if file_type == "json":
                # Handle JSON file containing university admission data
                with open(file_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                
                documents = []
                if isinstance(data, list):
                    # Process each university record as a separate document
                    for i, record in enumerate(data):
                        
                        # Check if this is the enhanced format (simple content-only structure)
                        if isinstance(record, dict) and len(record) == 1 and 'content' in record:
                            # Enhanced format: just content field
                            content = record.get('content', '').strip()
                            
                            if not content:
                                logger.warning(f"Record {i} has empty content, skipping...")
                                continue
                            
                            # Create metadata from the content by parsing key information
                            metadata = {
                                "source": file_path,
                                "document_type": "vietnamese_university_enhanced",
                                "record_index": i,
                                "format": "enhanced"
                            }
                            
                            # Extract structured information from content for better search
                            # Parse university name
                            if " tại " in content:
                                university_part = content.split(" tại ")[1].split(".")[0]
                                metadata["university"] = university_part.strip()
                            
                            # Parse program name (first part before "tại")
                            if " tại " in content:
                                program_part = content.split(" tại ")[0]
                                if program_part.startswith("Ngành "):
                                    metadata["major"] = program_part[6:].strip()  # Remove "Ngành "
                            
                            # Parse admission score
                            if "Điểm chuẩn năm 2024: " in content:
                                score_part = content.split("Điểm chuẩn năm 2024: ")[1].split(" điểm")[0]
                                try:
                                    metadata["admission_score"] = float(score_part)
                                except:
                                    metadata["admission_score"] = score_part
                            
                            # Parse subject combinations
                            if "Tổ hợp môn xét tuyển: " in content:
                                subjects_part = content.split("Tổ hợp môn xét tuyển: ")[1].split(".")[0]
                                metadata["subject_combinations"] = subjects_part.strip()
                            
                            # Parse competition level
                            if "Mức độ cạnh tranh: " in content:
                                competition_part = content.split("Mức độ cạnh tranh: ")[1].split(".")[0]
                                metadata["competition_level"] = competition_part.strip()
                            
                            # Parse admission method
                            if "Phương thức tuyển sinh: " in content:
                                method_part = content.split("Phương thức tuyển sinh: ")[1].split(".")[0]
                                metadata["admission_method"] = method_part.strip()
                            
                            logger.info(f"Enhanced format record {i}: {content[:100]}...")
                        
                        elif isinstance(record, dict) and 'metadata' in record and 'content' in record:
                            # Original enhanced format with metadata and content
                            content = record.get('content', '').strip()
                            metadata_obj = record.get('metadata', {})
                            
                            if not content:
                                logger.warning(f"Record {i} has empty content, skipping...")
                                continue
                            
                            # Build metadata from the metadata object
                            metadata = {
                                "source": file_path,
                                "document_type": "vietnamese_university_original",
                                "record_index": i,
                                "format": "original_enhanced"
                            }
                            
                            # Add all fields from the metadata object
                            for key, value in metadata_obj.items():
                                if value is not None:
                                    if isinstance(value, list):
                                        metadata[key] = ", ".join(str(v) for v in value)
                                    else:
                                        metadata[key] = str(value)
                            
                            logger.info(f"Original enhanced format record {i}: {content[:100]}...")
                        
                        else:
                            # Fallback for other formats
                            logger.warning(f"Unknown record format at index {i}, attempting to process...")
                            content = str(record)
                            metadata = {
                                "source": file_path,
                                "document_type": "vietnamese_university_unknown",
                                "record_index": i,
                                "format": "unknown"
                            }
                        
                        # Create document
                        doc = Document(
                            page_content=content,
                            metadata=metadata
                        )
                        documents.append(doc)
                        documents_processed += 1
                        
                        # Log progress for first few and every 100 documents
                        if i < 5 or i % 100 == 0:
                            logger.info(f"Processed document {i}: {len(content)} chars, university: {metadata.get('university', 'Unknown')}")
                
                logger.info(f"Processed {documents_processed} university records from enhanced JSON")
                
            elif file_type == "pdf":
                # Handle PDF file
                from langchain_community.document_loaders import PyPDFLoader
                
                loader = PyPDFLoader(file_path)
                documents = loader.load()
                
                # Enhance metadata for PDF documents
                for i, doc in enumerate(documents):
                    doc.metadata.update({
                        "source": file_path,
                        "document_type": "university_guidelines",
                        "page_number": i + 1,
                        "format": "pdf"
                    })
                
                documents_processed = len(documents)
                logger.info(f"Loaded {documents_processed} pages from PDF")
            
            else:
                raise ValueError(f"Unsupported file type: {file_type}")
            
            # Split documents into chunks for better retrieval
            if documents:
                logger.info(f"Processing {len(documents)} documents for vector store ingestion")
                
                # For the enhanced JSON format, documents are already well-sized
                # Only split if individual documents are extremely large
                oversized_docs = [doc for doc in documents if len(doc.page_content.encode('utf-8')) > 2000000]  # 2MB limit
                
                if oversized_docs:
                    logger.warning(f"Found {len(oversized_docs)} oversized documents, splitting them")
                    all_chunks = []
                    for doc in documents:
                        if len(doc.page_content.encode('utf-8')) > 2000000:
                            chunks = self.text_splitter.split_documents([doc])
                            all_chunks.extend(chunks)
                        else:
                            all_chunks.append(doc)
                    documents = all_chunks
                    logger.info(f"After splitting: {len(documents)} total chunks")
                
                total_chunks = len(documents)
                
                # Use Vietnamese vector store if available, otherwise fall back to regular vector store
                target_vector_store = self.vietnamese_vector_store or self.vector_store
                
                if not target_vector_store:
                    raise ValueError("No vector store available for ingestion")
                
                # Process documents in batches to avoid overwhelming Pinecone
                batch_size = 100  # Increase batch size since enhanced documents are more manageable
                total_docs = len(documents)
                successful_ingestions = 0
                
                for i in range(0, total_docs, batch_size):
                    batch = documents[i:i + batch_size]
                    batch_num = i // batch_size + 1
                    total_batches = (total_docs + batch_size - 1) // batch_size
                    
                    logger.info(f"Processing batch {batch_num}/{total_batches} ({len(batch)} documents)")
                    
                    # Log sample content from batch for debugging
                    if batch and logger.isEnabledFor(logging.INFO):
                        sample_doc = batch[0]
                        logger.info(f"Sample document content: {sample_doc.page_content[:200]}...")
                        logger.info(f"Sample document metadata: {sample_doc.metadata}")
                    
                    try:
                        # Add documents to vector store
                        await asyncio.get_event_loop().run_in_executor(
                            None,
                            lambda: target_vector_store.add_documents(batch)
                        )
                        successful_ingestions += len(batch)
                        logger.info(f"Successfully processed batch {batch_num}/{total_batches}")
                    except Exception as e:
                        logger.error(f"Error processing batch {batch_num}: {e}")
                        # Continue with next batch instead of failing completely
                        continue
                
                logger.info(f"Successfully ingested {successful_ingestions}/{total_docs} documents into vector store")
                
                # Set summaries created equal to successful ingestions for JSON data
                if file_type == "json":
                    summaries_created = successful_ingestions
                    total_chunks = successful_ingestions
            
            return {
                "success": True,
                "message": f"Successfully ingested Vietnamese university data from {file_type.upper()} file",
                "documents_processed": documents_processed,
                "summaries_created": summaries_created,
                "total_chunks": total_chunks
            }
            
        except Exception as e:
            logger.error(f"Error ingesting Vietnamese university data: {e}")
            return {
                "success": False,
                "message": f"Error: {str(e)}",
                "documents_processed": 0,
                "summaries_created": 0,
                "total_chunks": 0
            }
