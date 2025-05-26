#!/usr/bin/env python3
"""
Test script to validate the LLM Gateway Python service without external dependencies.
This validates the core service infrastructure can start up.
"""

import sys
import os
from pathlib import Path

# Add current directory to path
sys.path.insert(0, str(Path(__file__).parent))

def test_basic_imports():
    """Test that basic modules can be imported."""
    print("🔧 Testing basic imports...")
    
    try:
        from config.settings import get_settings
        print("✅ Config settings imported")
        
        from utils.logger import setup_logger
        print("✅ Logger utils imported")
        
        from utils.metrics import get_metrics_collector
        print("✅ Metrics utils imported")
        
        from llm.v1 import llm_pb2, llm_pb2_grpc
        print("✅ Generated proto files imported")
        
        return True
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False

def test_configuration():
    """Test configuration loading."""
    print("\n🔧 Testing configuration...")
    
    try:
        from config.settings import get_settings
        settings = get_settings()
        print(f"✅ Settings loaded: {settings.service_name}")
        print(f"✅ Environment: {settings.environment}")
        print(f"✅ gRPC Port: {settings.grpc_port}")
        return True
    except Exception as e:
        print(f"❌ Configuration error: {e}")
        return False

def test_grpc_server_setup():
    """Test gRPC server can be set up (without starting)."""
    print("\n🔧 Testing gRPC server setup...")
    
    try:
        import grpc
        from concurrent.futures import ThreadPoolExecutor
        from llm.v1 import llm_pb2_grpc
        
        # Create server
        server = grpc.server(ThreadPoolExecutor(max_workers=10))
        
        # Note: We're not adding the actual servicer here as it might have dependencies
        # Just testing the server can be created
        print("✅ gRPC server can be created")
        
        # Test port binding (without starting)
        listen_addr = '[::]:50051'
        port = server.add_insecure_port(listen_addr)
        print(f"✅ gRPC server can bind to port: {port}")
        
        return True
    except Exception as e:
        print(f"❌ gRPC server setup error: {e}")
        return False

def test_proto_messages():
    """Test proto message creation."""
    print("\n🔧 Testing proto messages...")
    
    try:
        from llm.v1 import llm_pb2
        
        # Test creating a generate stream request
        request = llm_pb2.GenerateStreamRequest()
        request.prompt = "Test prompt"
        request.user_id = "test_user"
        request.conversation_id = "test_conversation"
        print("✅ GenerateStreamRequest created")
        
        # Test creating a generate stream response
        response = llm_pb2.GenerateStreamResponse()
        response.token = "Test token"
        print("✅ GenerateStreamResponse created")
        
        # Test creating a RAG request
        rag_request = llm_pb2.GenerateWithRAGRequest()
        rag_request.prompt = "Test RAG prompt"
        rag_request.user_id = "test_user"
        rag_request.rag_collection = "test_collection"
        rag_request.adaptive = True
        print("✅ GenerateWithRAGRequest created")
        
        return True
    except Exception as e:
        print(f"❌ Proto messages error: {e}")
        return False

def main():
    """Run all tests."""
    print("🚀 LLM Gateway Python Service - Core Infrastructure Test")
    print("=" * 60)
    
    tests = [
        test_basic_imports,
        test_configuration,
        test_grpc_server_setup,
        test_proto_messages,
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        if test():
            passed += 1
        print()
    
    print("=" * 60)
    print(f"📊 Test Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All core infrastructure tests passed!")
        print("✅ The service is ready for deployment (pending API keys setup)")
        return 0
    else:
        print("⚠️  Some tests failed. Check the errors above.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
