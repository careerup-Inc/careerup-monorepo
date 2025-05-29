#!/usr/bin/env python3
"""
CareerUP Unified Python Test Suite
Consolidated testing for Python components and services
"""

import sys
import os
import asyncio
import json
import requests
from pathlib import Path
from typing import Dict, Any, Optional
import time

# Add the project root to the Python path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'

def log(message: str):
    timestamp = time.strftime('%Y-%m-%d %H:%M:%S')
    print(f"{Colors.BLUE}[{timestamp}]{Colors.NC} {message}")

def success(message: str):
    print(f"{Colors.GREEN}✅ {message}{Colors.NC}")

def warning(message: str):
    print(f"{Colors.YELLOW}⚠️  {message}{Colors.NC}")

def error(message: str):
    print(f"{Colors.RED}❌ {message}{Colors.NC}")

def info(message: str):
    print(f"{Colors.CYAN}ℹ️  {message}{Colors.NC}")

class CareerUPTestSuite:
    def __init__(self):
        self.admin_host = "http://localhost:8091"
        self.admin_api_key = "admin-secret-key-change-in-production"
        self.llm_service_path = project_root / "services" / "llm-gateway-py"
        
    def test_basic_imports(self) -> bool:
        """Test basic Python imports and modules"""
        print(f"\n{Colors.CYAN}📦 Testing Basic Imports{Colors.NC}")
        print("=" * 50)
        
        try:
            # Test standard library imports
            import json
            import os
            import sys
            import asyncio
            success("Standard library imports: OK")
            
            # Test common third-party imports
            try:
                import requests
                success("Requests library: OK")
            except ImportError:
                warning("Requests library not available")
            
            try:
                import numpy as np
                success("NumPy library: OK")
            except ImportError:
                warning("NumPy library not available")
                
            return True
            
        except Exception as e:
            error(f"Basic imports failed: {e}")
            return False
    
    def test_service_configuration(self) -> bool:
        """Test service configuration loading"""
        print(f"\n{Colors.CYAN}⚙️  Testing Service Configuration{Colors.NC}")
        print("=" * 50)
        
        try:
            # Test if we can import and use the service configuration
            config_path = self.llm_service_path / "config" / "settings.py"
            if not config_path.exists():
                warning("Service configuration file not found")
                return False
                
            # Add service path to sys.path temporarily
            sys.path.insert(0, str(self.llm_service_path))
            
            try:
                from config.settings import ServiceConfig
                config = ServiceConfig()
                
                # Basic configuration checks
                assert hasattr(config, 'grpc_port'), "GRPC port not configured"
                assert hasattr(config, 'http_port'), "HTTP port not configured"
                
                success(f"Service configuration loaded (GRPC: {config.grpc_port}, HTTP: {config.http_port})")
                return True
                
            except ImportError as e:
                warning(f"Could not import service configuration: {e}")
                return False
            except Exception as e:
                error(f"Configuration error: {e}")
                return False
            finally:
                # Remove service path from sys.path
                if str(self.llm_service_path) in sys.path:
                    sys.path.remove(str(self.llm_service_path))
                    
        except Exception as e:
            error(f"Configuration test failed: {e}")
            return False
    
    def test_admin_api_health(self) -> bool:
        """Test admin API health endpoint"""
        print(f"\n{Colors.CYAN}🏥 Testing Admin API Health{Colors.NC}")
        print("=" * 50)
        
        try:
            response = requests.get(f"{self.admin_host}/health", timeout=5)
            
            if response.status_code == 200:
                health_data = response.json()
                if health_data.get("status") == "healthy":
                    success("Admin API is healthy")
                    return True
                else:
                    warning(f"Admin API unhealthy: {health_data}")
                    return False
            else:
                error(f"Admin API health check failed: {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            error(f"Could not connect to Admin API: {e}")
            return False
        except Exception as e:
            error(f"Admin API health test failed: {e}")
            return False
    
    def test_embedding_system(self) -> bool:
        """Test embedding system functionality"""
        print(f"\n{Colors.CYAN}🔢 Testing Embedding System{Colors.NC}")
        print("=" * 50)
        
        try:
            # Test if we can access embedding functionality
            sys.path.insert(0, str(self.llm_service_path))
            
            try:
                # Test basic embedding concepts
                test_query = "Điểm chuẩn ngành Tài chính"
                log(f"Testing with query: {test_query}")
                
                # Try to load embedding service configuration
                from config.settings import ServiceConfig
                config = ServiceConfig()
                
                if hasattr(config, 'embedding_dimensions'):
                    success(f"Embedding dimensions configured: {config.embedding_dimensions}")
                else:
                    warning("Embedding dimensions not found in configuration")
                
                if hasattr(config, 'embedding_model'):
                    success(f"Embedding model configured: {config.embedding_model}")
                else:
                    warning("Embedding model not found in configuration")
                
                return True
                
            except ImportError as e:
                warning(f"Could not test embedding system: {e}")
                return False
            finally:
                if str(self.llm_service_path) in sys.path:
                    sys.path.remove(str(self.llm_service_path))
                    
        except Exception as e:
            error(f"Embedding system test failed: {e}")
            return False
    
    def test_pinecone_connectivity(self) -> bool:
        """Test Pinecone connectivity (if available)"""
        print(f"\n{Colors.CYAN}🌲 Testing Pinecone Connectivity{Colors.NC}")
        print("=" * 50)
        
        try:
            import pinecone
            success("Pinecone library available")
            
            # Check for API key in environment
            api_key = os.getenv('PINECONE_API_KEY')
            if api_key:
                success("Pinecone API key found in environment")
                
                # Try to initialize Pinecone client
                try:
                    pc = pinecone.Pinecone(api_key=api_key)
                    success("Pinecone client initialized")
                    
                    # List available indexes
                    try:
                        indexes = pc.list_indexes()
                        if indexes:
                            success(f"Found {len(indexes)} Pinecone indexes")
                            for idx in indexes:
                                info(f"  - {idx.name} ({idx.dimension}D)")
                        else:
                            warning("No Pinecone indexes found")
                        return True
                    except Exception as e:
                        warning(f"Could not list Pinecone indexes: {e}")
                        return False
                        
                except Exception as e:
                    error(f"Could not initialize Pinecone client: {e}")
                    return False
            else:
                warning("Pinecone API key not found in environment")
                return False
                
        except ImportError:
            warning("Pinecone library not available")
            return False
        except Exception as e:
            error(f"Pinecone connectivity test failed: {e}")
            return False
    
    def test_admin_endpoints(self) -> bool:
        """Test admin API endpoints"""
        print(f"\n{Colors.CYAN}🔧 Testing Admin Endpoints{Colors.NC}")
        print("=" * 50)
        
        try:
            headers = {"X-Admin-Key": self.admin_api_key}
            
            # Test status endpoint
            response = requests.get(f"{self.admin_host}/admin/status", 
                                  headers=headers, timeout=5)
            
            if response.status_code == 200:
                success("Admin status endpoint accessible")
                status_data = response.json()
                info(f"Service status: {status_data}")
            else:
                warning(f"Admin status endpoint returned: {response.status_code}")
            
            # Test metrics endpoint
            try:
                response = requests.get(f"{self.admin_host}/admin/metrics", 
                                      headers=headers, timeout=5)
                if response.status_code == 200:
                    success("Admin metrics endpoint accessible")
                else:
                    warning(f"Admin metrics endpoint returned: {response.status_code}")
            except Exception as e:
                warning(f"Could not test metrics endpoint: {e}")
            
            return True
            
        except requests.exceptions.RequestException as e:
            error(f"Could not connect to admin endpoints: {e}")
            return False
        except Exception as e:
            error(f"Admin endpoints test failed: {e}")
            return False
    
    def run_test_suite(self, test_type: str = "all") -> bool:
        """Run the complete test suite"""
        print(f"{Colors.BLUE}🚀 CareerUP Python Test Suite{Colors.NC}")
        print("=" * 60)
        print(f"Test type: {test_type}")
        print(f"Timestamp: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print("")
        
        tests = {
            "imports": self.test_basic_imports,
            "config": self.test_service_configuration,
            "health": self.test_admin_api_health,
            "embedding": self.test_embedding_system,
            "pinecone": self.test_pinecone_connectivity,
            "admin": self.test_admin_endpoints
        }
        
        results = {}
        
        if test_type == "all":
            test_list = list(tests.keys())
        elif test_type == "core":
            test_list = ["imports", "config", "health"]
        elif test_type == "services":
            test_list = ["health", "admin", "embedding"]
        elif test_type in tests:
            test_list = [test_type]
        else:
            error(f"Unknown test type: {test_type}")
            return False
        
        # Run selected tests
        for test_name in test_list:
            try:
                results[test_name] = tests[test_name]()
            except Exception as e:
                error(f"Test {test_name} crashed: {e}")
                results[test_name] = False
        
        # Summary
        print("\n" + "=" * 60)
        print(f"{Colors.CYAN}📊 Test Results Summary{Colors.NC}")
        print("=" * 60)
        
        passed = sum(1 for result in results.values() if result)
        total = len(results)
        
        for test_name, result in results.items():
            status = "PASS" if result else "FAIL"
            color = Colors.GREEN if result else Colors.RED
            print(f"{color}{test_name.upper()}: {status}{Colors.NC}")
        
        print("-" * 60)
        if passed == total:
            success(f"All tests passed ({passed}/{total})")
            return True
        else:
            error(f"Some tests failed ({passed}/{total} passed)")
            return False

def main():
    """Main entry point"""
    import argparse
    
    parser = argparse.ArgumentParser(description="CareerUP Python Test Suite")
    parser.add_argument("test_type", nargs="?", default="all",
                       choices=["all", "core", "services", "imports", "config", 
                               "health", "embedding", "pinecone", "admin"],
                       help="Type of tests to run")
    
    args = parser.parse_args()
    
    suite = CareerUPTestSuite()
    success = suite.run_test_suite(args.test_type)
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
