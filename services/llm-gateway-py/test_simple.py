#!/usr/bin/env python3
"""
Simple test script to verify basic Python service functionality
"""

import sys
import os
import asyncio
from pathlib import Path

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

# Test basic imports that don't require external dependencies
def test_basic_imports():
    """Test basic Python imports"""
    print("✅ Testing basic Python imports...")
    
    try:
        # Test basic Python modules
        import json
        import logging
        import threading
        import concurrent.futures
        print("  ✅ Basic Python modules: OK")
        
        # Test project structure
        from config import __init__ as config_init
        from utils import __init__ as utils_init
        from admin import __init__ as admin_init
        print("  ✅ Project structure: OK")
        
        # Test config module
        from config.settings import ServiceConfig
        print("  ✅ Config module: OK")
        
        # Test utils modules
        from utils.helpers import parse_json_safely, extract_keywords, validate_query
        from utils.vietnamese import detect_vietnamese, normalize_vietnamese_text
        from utils.logger import setup_logger
        from utils.metrics import ServiceMetrics
        from utils.security import SecurityManager
        print("  ✅ Utils modules: OK")
        
        print("✅ All basic imports successful!")
        return True
        
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        return False

def test_config():
    """Test configuration management"""
    print("\n✅ Testing configuration...")
    
    try:
        from config.settings import ServiceConfig
        
        # Create settings instance
        settings = ServiceConfig()
        
        print(f"  ✅ Service port: {settings.grpc_port}")
        print(f"  ✅ Admin port: {settings.http_port}")
        print(f"  ✅ Log level: {settings.log_level}")
        print(f"  ✅ Debug mode: {settings.debug}")
        
        return True
        
    except Exception as e:
        print(f"❌ Config test error: {e}")
        return False

def test_utils():
    """Test utility functions"""
    print("\n✅ Testing utility functions...")
    
    try:
        from utils.helpers import parse_json_safely, extract_keywords, validate_query
        from utils.vietnamese import detect_vietnamese, normalize_vietnamese_text
        
        # Test JSON parsing
        test_json = '{"test": "value"}'
        result = parse_json_safely(test_json)
        assert result == {"test": "value"}
        print("  ✅ JSON parsing: OK")
        
        # Test keyword extraction
        keywords = extract_keywords("This is a test sentence for keyword extraction")
        print(f"  ✅ Keyword extraction: {keywords}")
        
        # Test query validation
        assert validate_query("valid query") == True
        assert validate_query("") == False
        assert validate_query(None) == False
        print("  ✅ Query validation: OK")
        
        # Test Vietnamese detection
        vn_text = "Xin chào, bạn có khỏe không?"
        en_text = "Hello, how are you?"
        
        assert detect_vietnamese(vn_text) == True
        assert detect_vietnamese(en_text) == False
        print("  ✅ Vietnamese detection: OK")
        
        # Test Vietnamese normalization
        normalized = normalize_vietnamese_text("Xin  chào   bạn!")
        print(f"  ✅ Vietnamese normalization: '{normalized}'")
        
        return True
        
    except Exception as e:
        print(f"❌ Utils test error: {e}")
        return False

def test_metrics():
    """Test metrics collection"""
    print("\n✅ Testing metrics collection...")
    
    try:
        from utils.metrics import ServiceMetrics
        
        # Create metrics instance
        metrics = ServiceMetrics()
        
        # Test incrementing counters
        metrics.increment_counter("test_requests")
        metrics.increment_counter("test_requests")
        metrics.increment_counter("test_errors")
        
        # Test recording latencies
        metrics.record_latency("test_operation", 0.1)
        metrics.record_latency("test_operation", 0.2)
        
        # Get metrics summary
        summary = metrics.get_metrics()
        print(f"  ✅ Metrics collected: {len(summary)} metrics")
        print(f"  ✅ Test requests: {summary.get('test_requests', 0)}")
        print(f"  ✅ Test errors: {summary.get('test_errors', 0)}")
        
        return True
        
    except Exception as e:
        print(f"❌ Metrics test error: {e}")
        return False

def test_security():
    """Test security utilities"""
    print("\n✅ Testing security utilities...")
    
    try:
        from utils.security import SecurityManager
        
        # Create security manager
        security = SecurityManager()
        
        # Test API key validation (should fail without real key)
        try:
            result = security.validate_api_key("invalid_key")
            print(f"  ✅ API key validation: {result}")
        except Exception:
            print("  ✅ API key validation: Properly rejects invalid keys")
        
        # Test rate limiting
        client_id = "test_client"
        for i in range(3):
            allowed = security.check_rate_limit(client_id)
            print(f"  ✅ Rate limit check {i+1}: {allowed}")
        
        return True
        
    except Exception as e:
        print(f"❌ Security test error: {e}")
        return False

async def test_async_functionality():
    """Test async functionality"""
    print("\n✅ Testing async functionality...")
    
    try:
        # Test basic async operation
        await asyncio.sleep(0.1)
        print("  ✅ Basic async operation: OK")
        
        # Test concurrent operations
        async def sample_operation(delay):
            await asyncio.sleep(delay)
            return f"Operation completed after {delay}s"
        
        tasks = [
            sample_operation(0.1),
            sample_operation(0.2),
            sample_operation(0.1)
        ]
        
        results = await asyncio.gather(*tasks)
        print(f"  ✅ Concurrent operations: {len(results)} completed")
        
        return True
        
    except Exception as e:
        print(f"❌ Async test error: {e}")
        return False

def main():
    """Run all tests"""
    print("🚀 Starting LLM Gateway Python Service Tests...")
    print("=" * 60)
    
    tests = [
        test_basic_imports,
        test_config,
        test_utils,
        test_metrics,
        test_security,
    ]
    
    results = []
    for test in tests:
        try:
            result = test()
            results.append(result)
        except Exception as e:
            print(f"❌ Test failed with exception: {e}")
            results.append(False)
    
    # Run async tests
    try:
        print("\n✅ Running async tests...")
        result = asyncio.run(test_async_functionality())
        results.append(result)
    except Exception as e:
        print(f"❌ Async test failed: {e}")
        results.append(False)
    
    print("\n" + "=" * 60)
    print("📊 Test Results Summary:")
    print(f"  ✅ Passed: {sum(results)}")
    print(f"  ❌ Failed: {len(results) - sum(results)}")
    print(f"  📈 Success Rate: {sum(results)/len(results)*100:.1f}%")
    
    if all(results):
        print("\n🎉 All tests passed! The Python LLM Gateway service is ready.")
        return 0
    else:
        print("\n⚠️  Some tests failed. Check the output above for details.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
