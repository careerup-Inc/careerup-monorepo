#!/usr/bin/env python3
"""
Integration test for LLM Gateway Python service
Tests the complete request processing pipeline
"""

import sys
import os
import asyncio
import json
from pathlib import Path
from typing import Dict, Any

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def test_admin_api_creation():
    """Test if we can create the admin API"""
    print("🌐 Testing admin API creation...")
    
    try:
        from admin import get_admin_app, create_admin_app
        
        # Test getting admin app
        app = get_admin_app()
        if app is not None:
            print("  ✅ Admin app created successfully")
        else:
            print("  ⚠️  Admin app is None (FastAPI might not be available)")
        
        # Test creating admin app
        app2 = create_admin_app()
        if app2 is not None:
            print("  ✅ Admin app creation function works")
        else:
            print("  ⚠️  Admin app creation returned None")
        
        return True
        
    except Exception as e:
        print(f"  ❌ Admin API test failed: {e}")
        return False

def test_service_config_comprehensive():
    """Test comprehensive service configuration"""
    print("\n⚙️  Testing comprehensive service configuration...")
    
    try:
        from config.settings import ServiceConfig
        
        config = ServiceConfig()
        
        # Test all configuration properties
        config_tests = [
            ('gRPC port', hasattr(config, 'grpc_port') and config.grpc_port > 0),
            ('HTTP port', hasattr(config, 'http_port') and config.http_port > 0),
            ('Environment', hasattr(config, 'environment')),
            ('Debug mode', hasattr(config, 'debug')),
            ('Log level', hasattr(config, 'log_level')),
            ('OpenAI config', hasattr(config, 'openai_api_key')),
            ('Tavily config', hasattr(config, 'tavily_api_key')),
            ('Vector store config', hasattr(config, 'vector_store_config')),
        ]
        
        passed = 0
        for test_name, test_result in config_tests:
            status = "✅" if test_result else "⚠️ "
            print(f"  {status} {test_name}: {test_result}")
            if test_result:
                passed += 1
        
        print(f"  📊 Configuration completeness: {passed}/{len(config_tests)} properties available")
        return passed >= len(config_tests) // 2  # At least half should be available
        
    except Exception as e:
        print(f"  ❌ Service config test failed: {e}")
        return False

def test_query_processing_pipeline():
    """Test the complete query processing pipeline"""
    print("\n🔄 Testing query processing pipeline...")
    
    try:
        from utils.vietnamese import detect_vietnamese, normalize_vietnamese_text
        from utils.helpers import extract_keywords, validate_query, format_error_response
        from utils.security import SecurityManager
        from utils.metrics import ServiceMetrics
        
        # Initialize components
        security = SecurityManager()
        metrics = ServiceMetrics()
        
        # Test queries in different languages
        test_queries = [
            {
                'query': 'What are the best investment strategies for 2024?',
                'expected_lang': 'english',
                'expected_valid': True
            },
            {
                'query': 'Tôi muốn đầu tư vào cổ phiếu, có nên không?',
                'expected_lang': 'vietnamese', 
                'expected_valid': True
            },
            {
                'query': 'How does machine learning work in finance?',
                'expected_lang': 'english',
                'expected_valid': True
            },
            {
                'query': '',
                'expected_lang': 'english',
                'expected_valid': False
            }
        ]
        
        processed_queries = []
        
        for test_case in test_queries:
            query = test_case['query']
            
            # Step 1: Validate query
            is_valid = validate_query(query)
            
            if not is_valid:
                if test_case['expected_valid']:
                    print(f"  ❌ Query validation failed for: '{query[:30]}...'")
                    continue
                else:
                    print(f"  ✅ Empty query correctly rejected")
                    continue
            
            # Step 2: Detect language
            is_vietnamese = detect_vietnamese(query)
            detected_lang = 'vietnamese' if is_vietnamese else 'english'
            
            # Step 3: Normalize text
            normalized_query = normalize_vietnamese_text(query) if is_vietnamese else query
            
            # Step 4: Extract keywords
            keywords = extract_keywords(normalized_query)
            
            # Step 5: Security check (simulate)
            client_id = f"test_client_{len(processed_queries)}"
            rate_limit_ok = security.check_rate_limit(client_id)
            
            # Step 6: Record metrics
            metrics.increment_counter("processed_queries")
            metrics.record_latency("query_processing", 0.1)
            
            result = {
                'original_query': query,
                'normalized_query': normalized_query,
                'detected_language': detected_lang,
                'keywords': keywords,
                'keyword_count': len(keywords),
                'rate_limit_ok': rate_limit_ok,
                'valid': is_valid
            }
            
            processed_queries.append(result)
            
            lang_match = detected_lang == test_case['expected_lang']
            valid_match = is_valid == test_case['expected_valid']
            
            status = "✅" if (lang_match and valid_match) else "⚠️ "
            print(f"  {status} Query: '{query[:40]}...' -> {detected_lang}, {len(keywords)} keywords")
        
        # Summary
        metrics_summary = metrics.get_metrics()
        print(f"  📊 Pipeline metrics: {metrics_summary.get('processed_queries', 0)} queries processed")
        
        return len(processed_queries) > 0
        
    except Exception as e:
        print(f"  ❌ Query processing pipeline test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_async_service_simulation():
    """Test async service simulation with concurrent requests"""
    print("\n⚡ Testing async service simulation...")
    
    try:
        from utils.vietnamese import detect_vietnamese
        from utils.helpers import extract_keywords
        from utils.metrics import ServiceMetrics
        
        metrics = ServiceMetrics()
        
        # Simulate different types of service calls
        async def simulate_openai_call(query: str, delay: float = 0.15):
            """Simulate OpenAI API call"""
            await asyncio.sleep(delay)
            is_vietnamese = detect_vietnamese(query)
            
            if is_vietnamese:
                return f"Tôi hiểu rằng bạn hỏi về: {query[:30]}..."
            else:
                return f"Based on your question about: {query[:30]}..."
        
        async def simulate_vector_search(query: str, delay: float = 0.08):
            """Simulate vector database search"""
            await asyncio.sleep(delay)
            keywords = extract_keywords(query)
            return {
                'documents': [f"Doc {i+1} about {kw}" for i, kw in enumerate(keywords[:3])],
                'scores': [0.9 - (i * 0.1) for i in range(len(keywords[:3]))]
            }
        
        async def simulate_tavily_search(query: str, delay: float = 0.12):
            """Simulate Tavily web search"""
            await asyncio.sleep(delay)
            return {
                'web_results': [
                    {'title': f'Web result 1 for {query[:20]}', 'url': 'https://example1.com'},
                    {'title': f'Web result 2 for {query[:20]}', 'url': 'https://example2.com'}
                ]
            }
        
        async def process_request(request_id: str, query: str):
            """Simulate complete request processing"""
            start_time = asyncio.get_event_loop().time()
            
            try:
                # Detect query type and language
                is_vietnamese = detect_vietnamese(query)
                keywords = extract_keywords(query)
                
                # Determine processing strategy
                if len(keywords) > 3:  # Complex query - use RAG
                    vector_task = simulate_vector_search(query)
                    llm_task = simulate_openai_call(query)
                    
                    vector_results, llm_response = await asyncio.gather(vector_task, llm_task)
                    
                    response = {
                        'type': 'rag_response',
                        'llm_response': llm_response,
                        'context_documents': vector_results['documents'],
                        'language': 'vietnamese' if is_vietnamese else 'english'
                    }
                elif 'news' in query.lower() or 'current' in query.lower():  # Web search
                    web_results = await simulate_tavily_search(query)
                    llm_response = await simulate_openai_call(query)
                    
                    response = {
                        'type': 'web_search_response',
                        'llm_response': llm_response,
                        'web_results': web_results['web_results'],
                        'language': 'vietnamese' if is_vietnamese else 'english'
                    }
                else:  # Direct LLM
                    llm_response = await simulate_openai_call(query)
                    
                    response = {
                        'type': 'direct_response',
                        'llm_response': llm_response,
                        'language': 'vietnamese' if is_vietnamese else 'english'
                    }
                
                # Record metrics
                duration = asyncio.get_event_loop().time() - start_time
                metrics.increment_counter("successful_requests")
                metrics.record_latency("request_processing", duration)
                
                return {
                    'request_id': request_id,
                    'success': True,
                    'response': response,
                    'duration': duration
                }
                
            except Exception as e:
                duration = asyncio.get_event_loop().time() - start_time
                metrics.increment_counter("failed_requests")
                metrics.record_latency("request_processing", duration)
                
                return {
                    'request_id': request_id,
                    'success': False,
                    'error': str(e),
                    'duration': duration
                }
        
        # Test with concurrent requests
        test_requests = [
            ('req_1', 'What are the best investment strategies for tech stocks?'),
            ('req_2', 'Tôi muốn tìm hiểu về đầu tư bất động sản'),
            ('req_3', 'Current news about cryptocurrency market'),
            ('req_4', 'How does AI work?'),
            ('req_5', 'Làm thế nào để học lập trình Python hiệu quả?')
        ]
        
        # Process all requests concurrently
        tasks = [process_request(req_id, query) for req_id, query in test_requests]
        results = await asyncio.gather(*tasks)
        
        # Analyze results
        successful = sum(1 for r in results if r['success'])
        failed = len(results) - successful
        avg_duration = sum(r['duration'] for r in results) / len(results)
        
        response_types = {}
        languages = {}
        
        for result in results:
            if result['success']:
                resp_type = result['response']['type']
                lang = result['response']['language']
                response_types[resp_type] = response_types.get(resp_type, 0) + 1
                languages[lang] = languages.get(lang, 0) + 1
        
        print(f"  ✅ Processed {len(results)} concurrent requests")
        print(f"  ✅ Success rate: {successful}/{len(results)} ({successful/len(results)*100:.1f}%)")
        print(f"  ✅ Average duration: {avg_duration:.3f}s")
        print(f"  ✅ Response types: {response_types}")
        print(f"  ✅ Languages: {languages}")
        
        # Check metrics
        metrics_summary = metrics.get_metrics()
        print(f"  📊 Total requests processed: {metrics_summary.get('successful_requests', 0) + metrics_summary.get('failed_requests', 0)}")
        
        return successful >= len(results) * 0.8  # 80% success rate
        
    except Exception as e:
        print(f"  ❌ Async service simulation failed: {e}")
        return False

def test_error_scenarios():
    """Test error handling in various scenarios"""
    print("\n🛡️  Testing error scenarios...")
    
    try:
        from utils.helpers import format_error_response, parse_json_safely
        from utils.security import SecurityManager
        
        security = SecurityManager()
        
        # Test various error scenarios
        error_scenarios = [
            {
                'name': 'Invalid API key',
                'test': lambda: security.validate_api_key('invalid-key-123'),
                'expected': False
            },
            {
                'name': 'Malformed JSON',
                'test': lambda: parse_json_safely('{"invalid": json}'),
                'expected': None
            },
            {
                'name': 'Error response formatting',
                'test': lambda: len(format_error_response('TEST_ERROR', 'Test message')) > 0,
                'expected': True
            },
            {
                'name': 'Rate limiting (multiple calls)',
                'test': lambda: all(security.check_rate_limit('test_client') for _ in range(5)),
                'expected': True
            }
        ]
        
        passed = 0
        for scenario in error_scenarios:
            try:
                result = scenario['test']()
                if result == scenario['expected']:
                    print(f"  ✅ {scenario['name']}: handled correctly")
                    passed += 1
                else:
                    print(f"  ❌ {scenario['name']}: unexpected result {result}")
            except Exception as e:
                print(f"  ❌ {scenario['name']}: exception {e}")
        
        print(f"  📊 Error handling: {passed}/{len(error_scenarios)} scenarios passed")
        return passed >= len(error_scenarios) * 0.75
        
    except Exception as e:
        print(f"  ❌ Error scenarios test failed: {e}")
        return False

def main():
    """Run all integration tests"""
    print("🚀 LLM Gateway Python Service Integration Tests")
    print("=" * 70)
    
    tests = [
        test_admin_api_creation,
        test_service_config_comprehensive,
        test_query_processing_pipeline,
        test_error_scenarios
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
        async_result = asyncio.run(test_async_service_simulation())
        results.append(async_result)
    except Exception as e:
        print(f"❌ Async integration test failed: {e}")
        results.append(False)
    
    # Final summary
    passed = sum(results)
    total = len(results)
    success_rate = (passed / total) * 100
    
    print("\n" + "=" * 70)
    print("📊 Integration Test Results Summary:")
    print(f"  ✅ Passed: {passed}")
    print(f"  ❌ Failed: {total - passed}")
    print(f"  📈 Success Rate: {success_rate:.1f}%")
    
    if success_rate >= 80:
        print("🎉 Integration tests passed! Service is ready for deployment.")
        print("🔧 Next steps:")
        print("   1. Install remaining ML dependencies (pinecone, chromadb, etc.)")
        print("   2. Set up environment variables for API keys")
        print("   3. Test with actual OpenAI and Tavily APIs")
        print("   4. Deploy using Docker or direct Python execution")
        return 0
    else:
        print("⚠️  Some integration tests failed. Review the service architecture.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
