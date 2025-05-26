"""FastAPI admin endpoints for HTTP management of the LLM Gateway service."""

from fastapi import FastAPI, HTTPException, Depends, status, Request
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, PlainTextResponse
from pydantic import BaseModel, Field
from typing import Dict, Any, Optional, List
import asyncio
from datetime import datetime

from config.settings import get_settings
from utils.metrics import get_metrics_collector
from utils.security import validate_api_key, SecurityHeaders
from utils.logger import get_logger
from utils.helpers import sanitize_text, get_timestamp
from services.llm_service import LLMServicer

# Initialize logger
logger = get_logger("admin_api")

# Security
security = HTTPBearer(auto_error=False)

# Pydantic models
class HealthResponse(BaseModel):
    status: str
    timestamp: str
    version: str
    uptime_seconds: float
    
class MetricsResponse(BaseModel):
    current_stats: Dict[str, Any]
    error_summary: Dict[str, Any]
    
class TestQueryRequest(BaseModel):
    query: str = Field(..., min_length=1, max_length=1000)
    context: Optional[str] = Field(None, max_length=5000)
    use_rag: bool = Field(default=True)
    language: Optional[str] = Field(None)
    
class TestQueryResponse(BaseModel):
    request_id: str
    response: str
    query_type: str
    language: str
    duration: float
    tokens_used: Optional[int]
    sources: List[Dict[str, Any]]
    
class ConfigResponse(BaseModel):
    service_name: str
    environment: str
    debug: bool
    grpc_port: int
    http_port: int
    version: str


def create_admin_app() -> FastAPI:
    """Create FastAPI admin application.
    
    Returns:
        FastAPI application instance
    """
    settings = get_settings()
    
    app = FastAPI(
        title="LLM Gateway Admin API",
        description="Administrative endpoints for the LLM Gateway service",
        version="1.0.0",
        docs_url="/admin/docs",
        redoc_url="/admin/redoc",
        openapi_url="/admin/openapi.json"
    )
    
    # Add CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # Configure appropriately for production
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Add security headers middleware
    @app.middleware("http")
    async def add_security_headers(request: Request, call_next):
        response = await call_next(request)
        for header, value in SecurityHeaders.get_default_headers().items():
            response.headers[header] = value
        return response
    
    # Startup time for uptime calculation
    startup_time = datetime.utcnow()
    
    # Dependency for API key validation
    async def verify_api_key(credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)):
        if not credentials:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="API key required"
            )
        
        if not validate_api_key(credentials.credentials, {settings.admin_api_key}):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid API key"
            )
        
        return credentials.credentials
    
    @app.get("/health", response_model=HealthResponse, tags=["Health"])
    async def health_check():
        """Get service health status."""
        uptime = (datetime.utcnow() - startup_time).total_seconds()
        
        return HealthResponse(
            status="healthy",
            timestamp=get_timestamp(),
            version="1.0.0",
            uptime_seconds=uptime
        )
    
    @app.get("/admin/config", response_model=ConfigResponse, tags=["Admin"])
    async def get_config(api_key: str = Depends(verify_api_key)):
        """Get current service configuration."""
        return ConfigResponse(
            service_name=settings.service_name,
            environment=settings.environment,
            debug=settings.debug,
            grpc_port=settings.grpc_port,
            http_port=settings.http_port,
            version="1.0.0"
        )
    
    @app.get("/admin/metrics", response_model=MetricsResponse, tags=["Admin"])
    async def get_metrics(api_key: str = Depends(verify_api_key)):
        """Get service metrics and statistics."""
        metrics_collector = get_metrics_collector()
        
        return MetricsResponse(
            current_stats=metrics_collector.get_current_stats(),
            error_summary=metrics_collector.get_error_summary()
        )
    
    @app.get("/admin/metrics/export", tags=["Admin"])
    async def export_metrics(
        format_type: str = "json",
        api_key: str = Depends(verify_api_key)
    ):
        """Export metrics in specified format (json or prometheus)."""
        metrics_collector = get_metrics_collector()
        
        try:
            exported_data = metrics_collector.export_metrics(format_type)
            
            if format_type == "prometheus":
                return PlainTextResponse(
                    content=exported_data,
                    media_type="text/plain"
                )
            else:
                return JSONResponse(
                    content=exported_data,
                    media_type="application/json"
                )
        except ValueError as e:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=str(e)
            )
    
    @app.post("/admin/test", response_model=TestQueryResponse, tags=["Admin"])
    async def test_query(
        request: TestQueryRequest,
        api_key: str = Depends(verify_api_key)
    ):
        """Test the LLM service with a query."""
        try:
            # Sanitize input
            sanitized_query = sanitize_text(request.query)
            sanitized_context = sanitize_text(request.context) if request.context else ""
            
            # Create LLM service instance
            llm_service = LLMServicer()
            
            # Prepare request
            from proto.llm.v1 import llm_pb2
            
            grpc_request = llm_pb2.PromptRequest(
                query=sanitized_query,
                context=sanitized_context,
                use_rag=request.use_rag,
                language=request.language or "auto"
            )
            
            # Execute query
            start_time = datetime.utcnow()
            response = await llm_service.ProcessPrompt(grpc_request, None)
            duration = (datetime.utcnow() - start_time).total_seconds()
            
            return TestQueryResponse(
                request_id=response.request_id,
                response=response.response,
                query_type=response.query_type,
                language=response.language,
                duration=duration,
                tokens_used=response.tokens_used if response.tokens_used > 0 else None,
                sources=[
                    {
                        "content": source.content,
                        "url": source.url,
                        "title": source.title,
                        "score": source.score
                    }
                    for source in response.sources
                ]
            )
            
        except Exception as e:
            logger.error(f"Test query failed: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Query execution failed: {str(e)}"
            )
    
    @app.post("/admin/ingest", tags=["Admin"])
    async def ingest_documents(
        documents: List[Dict[str, Any]],
        api_key: str = Depends(verify_api_key)
    ):
        """Ingest documents into the vector store."""
        try:
            # Create LLM service instance
            llm_service = LLMServicer()
            
            # Prepare request
            from proto.llm.v1 import llm_pb2
            
            doc_requests = []
            for doc in documents:
                doc_request = llm_pb2.Document(
                    id=doc.get("id", ""),
                    content=sanitize_text(doc.get("content", "")),
                    url=doc.get("url", ""),
                    title=sanitize_text(doc.get("title", "")),
                    metadata=doc.get("metadata", {})
                )
                doc_requests.append(doc_request)
            
            grpc_request = llm_pb2.IngestRequest(documents=doc_requests)
            
            # Execute ingestion
            response = await llm_service.IngestDocuments(grpc_request, None)
            
            return {
                "message": response.message,
                "processed_count": response.processed_count,
                "failed_count": response.failed_count,
                "errors": list(response.errors)
            }
            
        except Exception as e:
            logger.error(f"Document ingestion failed: {str(e)}", exc_info=True)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Document ingestion failed: {str(e)}"
            )
    
    @app.delete("/admin/metrics", tags=["Admin"])
    async def reset_metrics(api_key: str = Depends(verify_api_key)):
        """Reset all service metrics (use with caution)."""
        metrics_collector = get_metrics_collector()
        metrics_collector.reset_metrics()
        
        logger.warning("Service metrics were reset by admin API")
        
        return {"message": "Metrics reset successfully"}
    
    @app.get("/admin/logs", tags=["Admin"])
    async def get_recent_logs(
        lines: int = 100,
        level: str = "INFO",
        api_key: str = Depends(verify_api_key)
    ):
        """Get recent log entries (if available)."""
        # This is a placeholder - actual implementation would depend on log storage
        return {
            "message": "Log retrieval not implemented",
            "note": "Configure log file path and implement file reading for this endpoint"
        }
    
    @app.post("/admin/cache/clear", tags=["Admin"])
    async def clear_cache(api_key: str = Depends(verify_api_key)):
        """Clear application caches."""
        # This is a placeholder for cache clearing logic
        return {"message": "Cache clearing not implemented"}
    
    @app.get("/admin/status", tags=["Admin"])
    async def get_detailed_status(api_key: str = Depends(verify_api_key)):
        """Get detailed service status including dependencies."""
        settings = get_settings()
        metrics = get_metrics_collector().get_current_stats()
        
        # Check external service connectivity
        status_checks = {
            "openai": "unknown",
            "pinecone": "unknown", 
            "tavily": "unknown"
        }
        
        try:
            # These would be actual connectivity checks
            # For now, just check if API keys are configured
            if settings.openai_api_key:
                status_checks["openai"] = "configured"
            if settings.pinecone_api_key:
                status_checks["pinecone"] = "configured"
            if settings.tavily_api_key:
                status_checks["tavily"] = "configured"
        except Exception as e:
            logger.error(f"Status check failed: {str(e)}")
        
        return {
            "service": {
                "name": settings.service_name,
                "version": "1.0.0",
                "environment": settings.environment,
                "uptime_seconds": (datetime.utcnow() - startup_time).total_seconds()
            },
            "metrics": metrics,
            "dependencies": status_checks,
            "timestamp": get_timestamp()
        }
    
    return app


# Global admin app instance
admin_app = None


def get_admin_app() -> FastAPI:
    """Get the admin FastAPI application instance.
    
    Returns:
        FastAPI admin application
    """
    global admin_app
    if admin_app is None:
        admin_app = create_admin_app()
    return admin_app
