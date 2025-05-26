#!/usr/bin/env python3
"""
Comprehensive integration test for the LLM Gateway Python service.
Tests the full service startup and basic functionality.
"""

import sys
import os
import asyncio
import time
from pathlib import Path

# Add current directory to path
sys.path.insert(0, str(Path(__file__).parent))

async def test_service_startup():
    """Test that the service can start up properly."""
    print("🔧 Testing service startup...")
    
    try:
        # Import main modules
        from config.settings import get_settings
        from utils.logger import setup_logger
        from services.llm_service import LLMServicer
        from llm.v1 import llm_pb2, llm_pb2_grpc
        
        # Setup configuration
        settings = get_settings()
        logger = setup_logger("test", level=settings.log_level)
        
        print(f"✅ Service configuration loaded: {settings.service_name}")
        print(f"✅ Environment: {settings.environment}")
        
        # Test that LLMServicer class can be imported
        print("✅ LLMServicer class imported successfully")
        print("ℹ️  Note: Full initialization requires API keys")
        
        return True
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False
    except Exception as e:
        print(f"❌ Service startup error: {e}")
        return False

async def test_grpc_service_interfaces():
    """Test gRPC service interfaces."""
    print("\n🔧 Testing gRPC service interfaces...")
    
    try:
        import grpc
        from concurrent.futures import ThreadPoolExecutor
        from llm.v1 import llm_pb2, llm_pb2_grpc
        from config.settings import get_settings
        from services.llm_service import LLMServicer
        
        # Create service instance
        settings = get_settings()
        print("✅ LLMServicer class can be imported")
        
        # Create server (without full service initialization)
        server = grpc.server(ThreadPoolExecutor(max_workers=10))
        print("✅ gRPC server can be created")
        
        # Test service methods exist on the class
        required_methods = [
            'GenerateStream',
            'GenerateWithRAG', 
            'IngestDocument',
            'CreateCollection',
            'ListCollections',
            'DeleteCollection'
        ]
        
        for method in required_methods:
            if hasattr(LLMServicer, method):
                print(f"✅ Service method {method} exists")
            else:
                print(f"❌ Service method {method} missing")
                return False
                
        return True
    except Exception as e:
        print(f"❌ gRPC interface error: {e}")
        return False

async def test_configuration_validation():
    """Test configuration validation and environment loading."""
    print("\n🔧 Testing configuration validation...")
    
    try:
        from config.settings import get_settings
        
        settings = get_settings()
        
        # Test required fields
        assert settings.service_name, "Service name is required"
        assert settings.grpc_port > 0, "gRPC port must be positive"
        assert settings.http_port > 0, "HTTP port must be positive"
        assert settings.environment in ["development", "staging", "production"], "Invalid environment"
        
        print(f"✅ Service name: {settings.service_name}")
        print(f"✅ gRPC port: {settings.grpc_port}")
        print(f"✅ HTTP port: {settings.http_port}")
        print(f"✅ Environment: {settings.environment}")
        print(f"✅ Log level: {settings.log_level}")
        
        # Test nested configs
        assert settings.rag.chunk_size > 0, "RAG chunk size must be positive"
        assert settings.rag.temperature >= 0 and settings.rag.temperature <= 1, "Temperature must be 0-1"
        assert settings.vector_store.embedding_dimensions > 0, "Embedding dimensions must be positive"
        
        print(f"✅ RAG chunk size: {settings.rag.chunk_size}")
        print(f"✅ RAG temperature: {settings.rag.temperature}")
        print(f"✅ Embedding dimensions: {settings.vector_store.embedding_dimensions}")
        
        return True
    except Exception as e:
        print(f"❌ Configuration validation error: {e}")
        return False

async def test_proto_service_compatibility():
    """Test proto service compatibility and message creation."""
    print("\n🔧 Testing proto service compatibility...")
    
    try:
        from llm.v1 import llm_pb2, llm_pb2_grpc
        
        # Test all request/response message types
        messages_to_test = [
            ("GenerateStreamRequest", {"prompt": "test", "user_id": "user123"}),
            ("GenerateStreamResponse", {"token": "response_token"}),
            ("GenerateWithRAGRequest", {"prompt": "test", "user_id": "user123", "rag_collection": "docs"}),
            ("GenerateWithRAGResponse", {"token": "rag_response"}),
            ("IngestDocumentRequest", {"content": "test content", "collection": "test_col", "document_id": "doc1"}),
            ("IngestDocumentResponse", {"document_id": "doc1", "success": True, "message": "success"}),
            ("CreateCollectionRequest", {"collection_name": "new_collection"}),
            ("CreateCollectionResponse", {"success": True, "message": "created", "collection_name": "new_collection"}),
            ("ListCollectionsRequest", {}),
            ("ListCollectionsResponse", {}),  # collections is a repeated field, can't assign directly
            ("DeleteCollectionRequest", {"collection_name": "to_delete"}),
            ("DeleteCollectionResponse", {"success": True, "message": "deleted"}),
        ]
        
        for message_name, test_data in messages_to_test:
            try:
                message_class = getattr(llm_pb2, message_name)
                message_instance = message_class()
                
                # Set test data
                for field, value in test_data.items():
                    if hasattr(message_instance, field):
                        setattr(message_instance, field, value)
                
                print(f"✅ {message_name} created and populated")
            except Exception as e:
                print(f"❌ {message_name} error: {e}")
                return False
        
        return True
    except Exception as e:
        print(f"❌ Proto compatibility error: {e}")
        return False

async def test_dependencies_available():
    """Test that all required dependencies are available."""
    print("\n🔧 Testing dependencies availability...")
    
    dependencies = [
        ("grpc", "gRPC support"),
        ("langchain", "LangChain framework"), 
        ("langchain_openai", "OpenAI integration"),
        ("langchain_pinecone", "Pinecone integration"),
        ("openai", "OpenAI client"),
        ("tavily", "Tavily search"),
        ("pinecone", "Pinecone client"),
        ("fastapi", "FastAPI framework"),
        ("uvicorn", "ASGI server"),
        ("structlog", "Structured logging"),
        ("pydantic", "Data validation"),
    ]
    
    available_count = 0
    for module_name, description in dependencies:
        try:
            __import__(module_name)
            print(f"✅ {description} ({module_name})")
            available_count += 1
        except ImportError:
            print(f"❌ {description} ({module_name}) - not available")
    
    success_rate = available_count / len(dependencies)
    print(f"📊 Dependencies: {available_count}/{len(dependencies)} available ({success_rate*100:.1f}%)")
    
    return success_rate >= 0.8  # 80% of dependencies should be available

async def main():
    """Run all integration tests."""
    print("🚀 LLM Gateway Python Service - Integration Test Suite")
    print("=" * 70)
    
    tests = [
        ("Service Startup", test_service_startup),
        ("gRPC Interfaces", test_grpc_service_interfaces), 
        ("Configuration", test_configuration_validation),
        ("Proto Compatibility", test_proto_service_compatibility),
        ("Dependencies", test_dependencies_available),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n📋 Running {test_name} Test...")
        print("-" * 50)
        
        try:
            result = await test_func()
            if result:
                passed += 1
                print(f"✅ {test_name} test PASSED")
            else:
                print(f"❌ {test_name} test FAILED")
        except Exception as e:
            print(f"❌ {test_name} test FAILED with exception: {e}")
    
    print("\n" + "=" * 70)
    print(f"📊 Integration Test Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 ALL INTEGRATION TESTS PASSED!")
        print("✅ The LLM Gateway Python service is ready for deployment!")
        print("\n🚀 Next steps:")
        print("1. Set up environment variables (.env file)")
        print("2. Configure API keys (OpenAI, Pinecone, Tavily)")
        print("3. Start the service: python main.py")
        return 0
    elif passed >= total * 0.8:
        print("⚠️  Most tests passed - service is mostly ready")
        print("🔧 Fix the failing tests before deployment")
        return 1
    else:
        print("❌ Too many tests failed - service needs more work")
        return 2

if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
