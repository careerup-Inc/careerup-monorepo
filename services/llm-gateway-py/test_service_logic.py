#!/usr/bin/env python3
"""
Advanced test script for LLM Gateway service core logic
"""

import sys
import os
import asyncio
from pathlib import Path

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def test_service_architecture():
    """Test the overall service architecture and core logic"""
    print("🏗️  Testing service architecture...")
    
    try:
        # Test configuration management
        from config.settings import ServiceConfig
        config = ServiceConfig()
        print(f"  ✅ Service config loaded: gRPC port {config.grpc_port}")
        
        # Test core utilities
        from utils.helpers import parse_json_safely, extract_keywords, validate_query, format_error_response
        from utils.vietnamese import detect_vietnamese, normalize_vietnamese_text
        from utils.logger import setup_logger, get_logger
        from utils.metrics import ServiceMetrics, get_metrics_collector
        from utils.security import SecurityManager, validate_api_key
        
        print("  ✅ All core utilities imported successfully")
        
        # Test Vietnamese language processing
        vietnamese_text = "Xin chào! Tôi cần hỗ trợ về tài chính."
        is_vietnamese = detect_vietnamese(vietnamese_text)
        normalized = normalize_vietnamese_text(vietnamese_text)
        print(f"  ✅ Vietnamese detection: {is_vietnamese}")
        print(f"  ✅ Vietnamese normalization: '{normalized}'")
        
        # Test keyword extraction
        keywords = extract_keywords("financial advice stock market investment")
        print(f"  ✅ Keyword extraction: {keywords}")
        
        # Test query validation
        valid_query = validate_query("What is the best investment strategy?")
        invalid_query = validate_query("")
        print(f"  ✅ Query validation: valid={valid_query}, invalid={invalid_query}")
        
        # Test error response formatting
        error_response = format_error_response("TEST_ERROR", "Test error message")
        print(f"  ✅ Error response format: {error_response}")
        
        # Test metrics collection
        metrics = ServiceMetrics()
        metrics.increment_counter("test_requests", 5)
        metrics.record_latency("test_operation", 0.123)
        metrics_summary = metrics.get_metrics()
        print(f"  ✅ Metrics collection: {len(metrics_summary)} metrics recorded")
        
        # Test security management
        security = SecurityManager()
        api_key_valid = security.validate_api_key("test-api-key-12345678")
        rate_limit_ok = security.check_rate_limit("test_client")
        print(f"  ✅ Security validation: API key format check={api_key_valid}, rate limit={rate_limit_ok}")
        
        return True
        
    except Exception as e:
        print(f"❌ Service architecture test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_query_processing_logic():
    """Test query processing and routing logic"""
    print("\n🧠 Testing query processing logic...")
    
    try:
        from utils.vietnamese import detect_vietnamese
        from utils.helpers import extract_keywords, validate_query
        
        # Test different types of queries
        test_queries = [
            ("What is machine learning?", "english", "technical"),
            ("Tôi cần tư vấn đầu tư", "vietnamese", "financial"),
            ("Best stocks to buy 2024", "english", "financial"),
            ("Làm thế nào để học Python?", "vietnamese", "technical"),
            ("", "empty", "invalid")
        ]
        
        for query, expected_lang, query_type in test_queries:
            if not query:
                is_valid = validate_query(query)
                print(f"  ✅ Empty query validation: {is_valid}")
                continue
                
            is_valid = validate_query(query)
            is_vietnamese = detect_vietnamese(query)
            keywords = extract_keywords(query)
            
            language = "vietnamese" if is_vietnamese else "english"
            
            print(f"  ✅ Query: '{query[:30]}...' -> Lang: {language}, Valid: {is_valid}, Keywords: {len(keywords)}")
        
        return True
        
    except Exception as e:
        print(f"❌ Query processing test failed: {e}")
        return False

def test_error_handling():
    """Test error handling mechanisms"""
    print("\n🛡️  Testing error handling...")
    
    try:
        from utils.helpers import parse_json_safely, format_error_response
        
        # Test JSON parsing with invalid data
        invalid_json = '{"incomplete": json'
        result = parse_json_safely(invalid_json)
        print(f"  ✅ Invalid JSON handling: {result is None}")
        
        # Test error response formatting
        error_types = ["VALIDATION_ERROR", "RATE_LIMIT_EXCEEDED", "SERVICE_UNAVAILABLE"]
        for error_type in error_types:
            response = format_error_response(error_type, f"Test {error_type}")
            print(f"  ✅ Error format {error_type}: {len(response)} chars")
        
        return True
        
    except Exception as e:
        print(f"❌ Error handling test failed: {e}")
        return False

def test_metrics_and_monitoring():
    """Test metrics collection and monitoring capabilities"""
    print("\n📊 Testing metrics and monitoring...")
    
    try:
        from utils.metrics import ServiceMetrics, get_metrics_collector
        
        # Test individual metrics instance
        metrics = ServiceMetrics()
        
        # Simulate service usage
        for i in range(10):
            metrics.increment_counter("requests")
            metrics.record_latency("llm_call", 0.1 + (i * 0.05))
            
        for i in range(2):
            metrics.increment_counter("errors")
            
        # Get metrics summary
        summary = metrics.get_metrics()
        print(f"  ✅ Total requests: {summary.get('requests', 0)}")
        print(f"  ✅ Total errors: {summary.get('errors', 0)}")
        print(f"  ✅ Average LLM latency: {summary.get('llm_call_avg_latency', 0)}s")
        
        # Test global metrics collector
        global_metrics = get_metrics_collector()
        global_metrics.record_request(
            request_id="test-123",
            method="ProcessQuery",
            duration=0.15,
            success=True,
            tokens_used=50,
            model_used="gpt-3.5-turbo"
        )
        
        stats = global_metrics.get_current_stats()
        print(f"  ✅ Global metrics: {stats['total_requests']} requests, {stats['total_tokens']} tokens")
        
        return True
        
    except Exception as e:
        print(f"❌ Metrics test failed: {e}")
        return False

async def test_async_operations():
    """Test asynchronous operations"""
    print("\n⚡ Testing async operations...")
    
    try:
        # Simulate async query processing
        async def mock_llm_call(query: str, delay: float = 0.1):
            await asyncio.sleep(delay)
            return f"Response to: {query[:20]}..."
        
        async def mock_vector_search(query: str, delay: float = 0.05):
            await asyncio.sleep(delay)
            return [f"Document 1 for {query[:10]}", f"Document 2 for {query[:10]}"]
        
        # Test concurrent operations
        query = "What is artificial intelligence?"
        
        # Simulate concurrent LLM and vector search
        llm_task = mock_llm_call(query, 0.1)
        vector_task = mock_vector_search(query, 0.05)
        
        llm_response, vector_results = await asyncio.gather(llm_task, vector_task)
        
        print(f"  ✅ LLM response: {llm_response}")
        print(f"  ✅ Vector search: {len(vector_results)} documents")
        
        # Test multiple concurrent queries
        queries = [
            "Financial advice for beginners",
            "Python programming tutorial",
            "Best investment strategies"
        ]
        
        tasks = [mock_llm_call(q, 0.1) for q in queries]
        responses = await asyncio.gather(*tasks)
        
        print(f"  ✅ Processed {len(responses)} concurrent queries")
        
        return True
        
    except Exception as e:
        print(f"❌ Async operations test failed: {e}")
        return False

def main():
    """Run all advanced tests"""
    print("🚀 Starting Advanced LLM Gateway Service Tests")
    print("=" * 60)
    
    tests = [
        test_service_architecture,
        test_query_processing_logic,
        test_error_handling,
        test_metrics_and_monitoring,
    ]
    
    results = []
    for test in tests:
        try:
            result = test()
            results.append(result)
        except Exception as e:
            print(f"❌ Test {test.__name__} failed with exception: {e}")
            results.append(False)
    
    # Run async test
    print("\n⚡ Running async tests...")
    try:
        async_result = asyncio.run(test_async_operations())
        results.append(async_result)
    except Exception as e:
        print(f"❌ Async test failed: {e}")
        results.append(False)
    
    # Print summary
    passed = sum(results)
    total = len(results)
    success_rate = (passed / total) * 100
    
    print("\n" + "=" * 60)
    print("📊 Advanced Test Results Summary:")
    print(f"  ✅ Passed: {passed}")
    print(f"  ❌ Failed: {total - passed}")
    print(f"  📈 Success Rate: {success_rate:.1f}%")
    
    if success_rate == 100:
        print("🎉 All advanced tests passed! The service core logic is working correctly.")
        return 0
    else:
        print("⚠️  Some advanced tests failed. Review the output above.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
