#!/usr/bin/env python3
"""
Service Structure Validation Test
Tests the service structure and gRPC interface without external dependencies
"""

import sys
import os
import asyncio
from pathlib import Path
from unittest.mock import Mock, patch

# Add the project root to the Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def test_proto_structure():
    """Test proto file structure and imports"""
    print("📋 Testing proto structure...")
    
    try:
        # Check if proto files exist
        proto_root = project_root.parent.parent / "proto"
        llm_proto = proto_root / "llm" / "v1" / "llm.proto"
        
        if llm_proto.exists():
            print(f"  ✅ Proto file found: {llm_proto}")
            
            # Read proto content
            with open(llm_proto, 'r') as f:
                content = f.read()
                
            # Check for required services
            services = []
            if "service LLMService" in content:
                services.append("LLMService")
            if "rpc GenerateStream" in content:
                services.append("GenerateStream")
            if "rpc GenerateWithRAG" in content:
                services.append("GenerateWithRAG")
                
            print(f"  ✅ Found services: {services}")
            return len(services) > 0
        else:
            print(f"  ❌ Proto file not found: {llm_proto}")
            return False
            
    except Exception as e:
        print(f"❌ Proto structure test failed: {e}")
        return False

def test_service_structure():
    """Test service module structure"""
    print("\n🏗️  Testing service structure...")
    
    try:
        # Test if service modules exist
        services_dir = project_root / "services"
        if not services_dir.exists():
            print(f"  ❌ Services directory not found: {services_dir}")
            return False
            
        llm_service_file = services_dir / "llm_service.py"
        if not llm_service_file.exists():
            print(f"  ❌ LLM service file not found: {llm_service_file}")
            return False
            
        print(f"  ✅ LLM service file found: {llm_service_file}")
        
        # Read service content to check structure
        with open(llm_service_file, 'r') as f:
            content = f.read()
            
        # Check for required classes and methods
        checks = {
            'QueryRoute enum': 'class QueryRoute' in content,
            'RAGState dataclass': 'class RAGState' in content,
            'LLMService class': 'class LLMService' in content,
            'GenerateStream method': 'GenerateStream' in content,
            'GenerateWithRAG method': 'GenerateWithRAG' in content,
        }
        
        for check, status in checks.items():
            status_icon = "✅" if status else "❌"
            print(f"  {status_icon} {check}: {'Found' if status else 'Missing'}")
            
        return all(checks.values())
        
    except Exception as e:
        print(f"❌ Service structure test failed: {e}")
        return False

def test_main_service_entry():
    """Test main service entry point"""
    print("\n🚀 Testing main service entry...")
    
    try:
        main_file = project_root / "main.py"
        if not main_file.exists():
            print(f"  ❌ Main file not found: {main_file}")
            return False
            
        print(f"  ✅ Main file found: {main_file}")
        
        # Read main content
        with open(main_file, 'r') as f:
            content = f.read()
            
        # Check for required imports and functionality
        checks = {
            'asyncio import': 'import asyncio' in content,
            'grpc import': 'import grpc' in content,
            'settings import': 'from config.settings import' in content,
            'LLMService import': 'from services.llm_service import LLMService' in content,
            'main function': 'async def main():' in content,
            'gRPC server setup': 'grpc.aio.server' in content,
            'server start': 'await server.start()' in content,
        }
        
        for check, status in checks.items():
            status_icon = "✅" if status else "❌"
            print(f"  {status_icon} {check}: {'Found' if status else 'Missing'}")
            
        return all(checks.values())
        
    except Exception as e:
        print(f"❌ Main service test failed: {e}")
        return False

def test_configuration_completeness():
    """Test configuration completeness"""
    print("\n⚙️  Testing configuration completeness...")
    
    try:
        from config.settings import ServiceConfig
        config = ServiceConfig()
        
        # Check required configuration fields
        required_fields = [
            'grpc_port', 'http_port', 'log_level', 'debug',
            'max_workers', 'environment', 'enable_admin_api'
        ]
        
        missing_fields = []
        for field in required_fields:
            if not hasattr(config, field):
                missing_fields.append(field)
            else:
                value = getattr(config, field)
                print(f"  ✅ {field}: {value}")
        
        if missing_fields:
            print(f"  ❌ Missing fields: {missing_fields}")
            return False
            
        return True
        
    except Exception as e:
        print(f"❌ Configuration test failed: {e}")
        return False

def test_dependency_structure():
    """Test dependency structure without importing heavy libs"""
    print("\n📦 Testing dependency structure...")
    
    try:
        # Check requirements file
        req_file = project_root / "requirements.txt"
        if not req_file.exists():
            print(f"  ❌ Requirements file not found: {req_file}")
            return False
            
        with open(req_file, 'r') as f:
            requirements = f.read()
            
        # Check for critical dependencies
        critical_deps = [
            'grpcio', 'langchain', 'openai', 'pinecone-client',
            'fastapi', 'uvicorn', 'pydantic', 'structlog'
        ]
        
        found_deps = []
        missing_deps = []
        
        for dep in critical_deps:
            if dep in requirements.lower():
                found_deps.append(dep)
                print(f"  ✅ {dep}: Listed in requirements")
            else:
                missing_deps.append(dep)
                print(f"  ⚠️  {dep}: Not found in requirements")
        
        print(f"  📊 Found {len(found_deps)}/{len(critical_deps)} critical dependencies")
        
        return len(found_deps) >= len(critical_deps) * 0.8  # 80% threshold
        
    except Exception as e:
        print(f"❌ Dependency structure test failed: {e}")
        return False

def test_admin_api_structure():
    """Test admin API structure"""
    print("\n🔧 Testing admin API structure...")
    
    try:
        admin_dir = project_root / "admin"
        if not admin_dir.exists():
            print(f"  ❌ Admin directory not found: {admin_dir}")
            return False
            
        admin_api_file = admin_dir / "api.py"
        if not admin_api_file.exists():
            print(f"  ❌ Admin API file not found: {admin_api_file}")
            return False
            
        print(f"  ✅ Admin API file found: {admin_api_file}")
        
        # Check admin init file
        admin_init_file = admin_dir / "__init__.py"
        if admin_init_file.exists():
            with open(admin_init_file, 'r') as f:
                init_content = f.read()
                
            if 'get_admin_app' in init_content:
                print("  ✅ Admin app function exported")
                return True
            else:
                print("  ⚠️  Admin app function not properly exported")
                return False
        else:
            print("  ❌ Admin __init__.py not found")
            return False
            
    except Exception as e:
        print(f"❌ Admin API structure test failed: {e}")
        return False

def main():
    """Run all service structure validation tests"""
    print("🚀 Running Service Structure Validation Tests")
    print("=" * 60)
    
    tests = [
        test_proto_structure,
        test_service_structure,
        test_main_service_entry,
        test_configuration_completeness,
        test_dependency_structure,
        test_admin_api_structure
    ]
    
    results = []
    for test in tests:
        try:
            result = test()
            results.append(result)
        except Exception as e:
            print(f"❌ Test {test.__name__} failed with exception: {e}")
            results.append(False)
    
    # Print summary
    passed = sum(results)
    total = len(results)
    success_rate = (passed / total) * 100
    
    print("\n" + "=" * 60)
    print("📊 Service Structure Validation Summary:")
    print(f"  ✅ Passed: {passed}")
    print(f"  ❌ Failed: {total - passed}")
    print(f"  📈 Success Rate: {success_rate:.1f}%")
    
    if success_rate >= 80:
        print("🎉 Service structure is ready for deployment!")
        print("💡 Next steps:")
        print("   1. Install missing dependencies: pip install -r requirements.txt")
        print("   2. Set up environment variables")
        print("   3. Generate gRPC proto files")
        print("   4. Start the service: python main.py")
        return 0
    else:
        print("⚠️  Service structure needs attention before deployment.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
