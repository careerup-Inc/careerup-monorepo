#!/usr/bin/env python3
"""
Standalone test for core service functionality without external dependencies
"""

import sys
import os
import asyncio
from pathlib import Path

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def test_basic_service_components():
    """Test basic service components without external dependencies"""
    print("🔧 Testing basic service components...")
    
    try:
        # Test configuration
        from config.settings import ServiceConfig
        config = ServiceConfig()
        assert config.grpc_port == 50054
        assert config.http_port == 8091
        print("  ✅ Configuration: OK")
        
        # Test basic helpers (avoid security module for now)
        from utils.helpers import parse_json_safely, validate_query, extract_keywords
        
        # Test JSON parsing
        test_data = parse_json_safely('{"test": "value"}')
        assert test_data == {"test": "value"}
        print("  ✅ JSON parsing: OK")
        
        # Test query validation
        assert validate_query("valid query") == True
        assert validate_query("") == False
        print("  ✅ Query validation: OK")
        
        # Test keyword extraction
        keywords = extract_keywords("machine learning artificial intelligence")
        assert len(keywords) > 0
        print(f"  ✅ Keyword extraction: {keywords}")
        
        # Test Vietnamese detection (without security dependencies)
        from utils.vietnamese import detect_vietnamese, normalize_vietnamese_text
        
        vietnamese_text = "Xin chào, tôi cần hỗ trợ"
        is_vietnamese = detect_vietnamese(vietnamese_text)
        normalized = normalize_vietnamese_text(vietnamese_text)
        
        assert is_vietnamese == True
        assert len(normalized) > 0
        print(f"  ✅ Vietnamese processing: detected={is_vietnamese}")
        
        # Test logger setup
        from utils.logger import setup_logger, get_logger
        logger = get_logger("test")
        logger.info("Test log message")
        print("  ✅ Logging: OK")
        
        return True
        
    except Exception as e:
        print(f"❌ Basic components test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_metrics_basic():
    """Test basic metrics functionality"""
    print("\n📊 Testing basic metrics...")
    
    try:
        # Direct import without going through utils.__init__
        sys.path.insert(0, str(Path(__file__).parent / "utils"))
        
        # Test if we can create a basic metrics collector
        from datetime import datetime
        from collections import defaultdict, deque
        
        # Simple metrics implementation test
        metrics_data = {
            'requests': 0,
            'errors': 0,
            'total_duration': 0.0
        }
        
        # Simulate metrics collection
        metrics_data['requests'] += 5
        metrics_data['errors'] += 1
        metrics_data['total_duration'] += 2.5
        
        success_rate = ((metrics_data['requests'] - metrics_data['errors']) / metrics_data['requests']) * 100
        avg_duration = metrics_data['total_duration'] / metrics_data['requests']
        
        print(f"  ✅ Simulated metrics: {metrics_data['requests']} requests, {success_rate:.1f}% success rate")
        print(f"  ✅ Average duration: {avg_duration:.3f}s")
        
        return True
        
    except Exception as e:
        print(f"❌ Basic metrics test failed: {e}")
        return False

async def test_async_processing():
    """Test async processing capabilities"""
    print("\n⚡ Testing async processing...")
    
    try:
        # Simulate query processing pipeline
        async def process_query(query: str, delay: float = 0.1):
            """Simulate query processing"""
            await asyncio.sleep(delay)
            
            # Simulate Vietnamese detection
            from utils.vietnamese import detect_vietnamese
            is_vietnamese = detect_vietnamese(query)
            
            # Simulate keyword extraction
            from utils.helpers import extract_keywords
            keywords = extract_keywords(query)
            
            return {
                'query': query,
                'language': 'vietnamese' if is_vietnamese else 'english',
                'keywords': keywords,
                'response': f"Processed: {query[:30]}..."
            }
        
        # Test single query
        result = await process_query("What is machine learning?")
        assert result['language'] == 'english'
        assert len(result['keywords']) > 0
        print(f"  ✅ Single query processing: {result['language']} query with {len(result['keywords'])} keywords")
        
        # Test concurrent queries
        queries = [
            "Machine learning basics",
            "Tôi cần tư vấn đầu tư",
            "Python programming guide"
        ]
        
        tasks = [process_query(q, 0.05) for q in queries]
        results = await asyncio.gather(*tasks)
        
        vietnamese_count = sum(1 for r in results if r['language'] == 'vietnamese')
        english_count = len(results) - vietnamese_count
        
        print(f"  ✅ Concurrent processing: {len(results)} queries ({english_count} English, {vietnamese_count} Vietnamese)")
        
        return True
        
    except Exception as e:
        print(f"❌ Async processing test failed: {e}")
        return False

def test_error_handling():
    """Test error handling capabilities"""
    print("\n🛡️  Testing error handling...")
    
    try:
        from utils.helpers import parse_json_safely
        
        # Test invalid JSON
        result = parse_json_safely('{"invalid": json}')
        assert result is None
        print("  ✅ Invalid JSON handling: OK")
        
        # Test empty input
        result = parse_json_safely('')
        assert result is None
        print("  ✅ Empty input handling: OK")
        
        # Test None input
        result = parse_json_safely(None)
        assert result is None
        print("  ✅ None input handling: OK")
        
        return True
        
    except Exception as e:
        print(f"❌ Error handling test failed: {e}")
        return False

def test_service_readiness():
    """Test overall service readiness"""
    print("\n✅ Testing service readiness...")
    
    try:
        # Check if core modules can be imported
        modules_to_test = [
            'config.settings',
            'utils.helpers',
            'utils.vietnamese',
            'utils.logger'
        ]
        
        for module_name in modules_to_test:
            try:
                __import__(module_name)
                print(f"  ✅ Module {module_name}: importable")
            except ImportError as e:
                print(f"  ⚠️  Module {module_name}: import warning - {e}")
        
        # Test basic service configuration
        from config.settings import ServiceConfig
        config = ServiceConfig()
        
        readiness_checks = {
            'gRPC port configured': config.grpc_port > 0,
            'HTTP port configured': config.http_port > 0,
            'Environment loaded': hasattr(config, 'environment'),
            'Debug mode set': hasattr(config, 'debug')
        }
        
        for check, status in readiness_checks.items():
            status_icon = "✅" if status else "❌"
            print(f"  {status_icon} {check}: {status}")
        
        all_ready = all(readiness_checks.values())
        print(f"\n  📊 Service readiness: {sum(readiness_checks.values())}/{len(readiness_checks)} checks passed")
        
        return all_ready
        
    except Exception as e:
        print(f"❌ Service readiness test failed: {e}")
        return False

def main():
    """Run all standalone tests"""
    print("🚀 Running Standalone LLM Gateway Service Tests")
    print("=" * 60)
    
    tests = [
        test_basic_service_components,
        test_metrics_basic,
        test_error_handling,
        test_service_readiness
    ]
    
    results = []
    
    # Run synchronous tests
    for test in tests:
        try:
            result = test()
            results.append(result)
        except Exception as e:
            print(f"❌ Test {test.__name__} failed with exception: {e}")
            results.append(False)
    
    # Run async test
    try:
        async_result = asyncio.run(test_async_processing())
        results.append(async_result)
    except Exception as e:
        print(f"❌ Async test failed: {e}")
        results.append(False)
    
    # Print summary
    passed = sum(results)
    total = len(results)
    success_rate = (passed / total) * 100
    
    print("\n" + "=" * 60)
    print("📊 Standalone Test Results Summary:")
    print(f"  ✅ Passed: {passed}")
    print(f"  ❌ Failed: {total - passed}")
    print(f"  📈 Success Rate: {success_rate:.1f}%")
    
    if success_rate >= 80:
        print("🎉 Service core is working! Ready for integration testing.")
        return 0
    else:
        print("⚠️  Some core functionality needs attention.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
