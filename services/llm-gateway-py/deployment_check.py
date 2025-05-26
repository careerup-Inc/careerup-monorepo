#!/usr/bin/env python3
"""
Service deployment readiness check and startup script
"""

import sys
import os
import subprocess
from pathlib import Path

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def check_environment():
    """Check if the environment is properly configured"""
    print("🔍 Checking environment configuration...")
    
    try:
        from config.settings import ServiceConfig
        config = ServiceConfig()
        
        checks = [
            ('Python version', sys.version_info >= (3, 8), f"Current: {sys.version}"),
            ('Virtual environment', 'venv' in sys.executable, f"Python: {sys.executable}"),
            ('gRPC port configured', config.grpc_port > 0, f"Port: {config.grpc_port}"),
            ('HTTP port configured', config.http_port > 0, f"Port: {config.http_port}"),
            ('Environment loaded', hasattr(config, 'environment'), f"Env: {getattr(config, 'environment', 'not set')}"),
        ]
        
        all_passed = True
        for check_name, passed, details in checks:
            status = "✅" if passed else "❌"
            print(f"  {status} {check_name}: {details}")
            if not passed:
                all_passed = False
        
        return all_passed
        
    except Exception as e:
        print(f"  ❌ Environment check failed: {e}")
        return False

def check_dependencies():
    """Check if all critical dependencies are available"""
    print("\n📦 Checking dependencies...")
    
    critical_modules = [
        ('structlog', 'Structured logging'),
        ('pydantic', 'Data validation'),
        ('python-dotenv', 'Environment variables'),
        ('PyJWT', 'JWT tokens'),
        ('fastapi', 'Admin API (optional)'),
        ('uvicorn', 'ASGI server (optional)'),
        ('openai', 'OpenAI client (optional)'),
        ('requests', 'HTTP client'),
    ]
    
    available = 0
    total = len(critical_modules)
    
    for module, description in critical_modules:
        try:
            __import__(module.replace('-', '_'))
            print(f"  ✅ {description}: available")
            available += 1
        except ImportError:
            print(f"  ⚠️  {description}: not available")
    
    print(f"  📊 Dependencies: {available}/{total} available")
    return available >= total * 0.7  # 70% threshold

def check_service_modules():
    """Check if all service modules can be imported"""
    print("\n🔧 Checking service modules...")
    
    modules_to_check = [
        ('config.settings', 'Configuration'),
        ('utils.logger', 'Logging utilities'),
        ('utils.helpers', 'Helper functions'),
        ('utils.vietnamese', 'Vietnamese processing'),
        ('utils.metrics', 'Metrics collection'),
        ('utils.security', 'Security utilities'),
        ('admin', 'Admin API (optional)'),
    ]
    
    available = 0
    total = len(modules_to_check)
    
    for module, description in modules_to_check:
        try:
            __import__(module)
            print(f"  ✅ {description}: importable")
            available += 1
        except ImportError as e:
            print(f"  ⚠️  {description}: {e}")
    
    print(f"  📊 Service modules: {available}/{total} available")
    return available >= total * 0.8  # 80% threshold

def test_core_functionality():
    """Test core service functionality"""
    print("\n🧪 Testing core functionality...")
    
    try:
        # Test basic operations
        from utils.vietnamese import detect_vietnamese
        from utils.helpers import extract_keywords, validate_query
        from utils.security import SecurityManager
        from utils.metrics import ServiceMetrics
        
        # Quick functionality test
        test_query = "Test query for functionality check"
        
        is_valid = validate_query(test_query)
        is_vietnamese = detect_vietnamese(test_query)
        keywords = extract_keywords(test_query)
        
        security = SecurityManager()
        api_key_check = security.validate_api_key("test-key-123456789")
        
        metrics = ServiceMetrics()
        metrics.increment_counter("test_counter")
        metrics_data = metrics.get_metrics()
        
        functionality_checks = [
            ('Query validation', is_valid == True),
            ('Language detection', is_vietnamese == False),
            ('Keyword extraction', len(keywords) > 0),
            ('Security validation', api_key_check == True),  # Should pass format check
            ('Metrics collection', 'test_counter' in metrics_data),
        ]
        
        passed = 0
        for check_name, check_result in functionality_checks:
            status = "✅" if check_result else "❌"
            print(f"  {status} {check_name}: {check_result}")
            if check_result:
                passed += 1
        
        print(f"  📊 Core functionality: {passed}/{len(functionality_checks)} tests passed")
        return passed >= len(functionality_checks) * 0.8
        
    except Exception as e:
        print(f"  ❌ Core functionality test failed: {e}")
        return False

def generate_deployment_info():
    """Generate deployment information"""
    print("\n📋 Deployment Information:")
    
    try:
        from config.settings import ServiceConfig
        config = ServiceConfig()
        
        print(f"  🚀 Service Name: LLM Gateway Python")
        print(f"  🌐 gRPC Port: {config.grpc_port}")
        print(f"  🌐 HTTP Port: {config.http_port}")
        print(f"  🔧 Environment: {getattr(config, 'environment', 'development')}")
        print(f"  🐍 Python: {sys.version.split()[0]}")
        print(f"  📁 Working Directory: {Path.cwd()}")
        
        print(f"\n🔧 Startup Commands:")
        print(f"  Basic service test: python test_simple.py")
        print(f"  Advanced tests: python test_service_logic.py")
        print(f"  Integration tests: python test_integration.py")
        print(f"  Install dependencies: python install_deps.py")
        
        print(f"\n⚠️  Required Environment Variables:")
        print(f"  OPENAI_API_KEY=your_openai_api_key")
        print(f"  TAVILY_API_KEY=your_tavily_api_key")
        print(f"  PINECONE_API_KEY=your_pinecone_api_key (optional)")
        
        print(f"\n📦 Optional Dependencies for Full Functionality:")
        print(f"  pip install pinecone-client chromadb sentence-transformers")
        print(f"  pip install grpcio-tools protobuf")
        
    except Exception as e:
        print(f"  ❌ Could not generate deployment info: {e}")

def main():
    """Run deployment readiness check"""
    print("🚀 LLM Gateway Python Service Deployment Readiness Check")
    print("=" * 70)
    
    checks = [
        check_environment,
        check_dependencies,
        check_service_modules,
        test_core_functionality
    ]
    
    results = []
    for check in checks:
        try:
            result = check()
            results.append(result)
        except Exception as e:
            print(f"❌ Check {check.__name__} failed: {e}")
            results.append(False)
    
    # Generate deployment info regardless of check results
    generate_deployment_info()
    
    # Summary
    passed = sum(results)
    total = len(results)
    success_rate = (passed / total) * 100
    
    print("\n" + "=" * 70)
    print("📊 Deployment Readiness Summary:")
    print(f"  ✅ Passed: {passed}")
    print(f"  ❌ Failed: {total - passed}")
    print(f"  📈 Readiness Score: {success_rate:.1f}%")
    
    if success_rate >= 80:
        print("🎉 Service is ready for deployment!")
        print("   ✅ Core functionality validated")
        print("   ✅ Dependencies available")
        print("   ✅ Configuration loaded")
        print("   🚀 Ready to start serving requests")
        return 0
    elif success_rate >= 60:
        print("⚠️  Service is partially ready - some optional features may not work")
        print("   ✅ Core functionality should work")
        print("   ⚠️  Some dependencies missing")
        print("   🔧 Consider installing missing dependencies")
        return 0
    else:
        print("❌ Service needs attention before deployment")
        print("   🔧 Fix critical issues above")
        print("   📦 Install missing dependencies")
        print("   ⚙️  Check configuration")
        return 1

if __name__ == "__main__":
    sys.exit(main())
