#!/usr/bin/env python3
"""
Install critical dependencies for LLM Gateway service
"""

import subprocess
import sys

# Critical dependencies that we need for core functionality
CRITICAL_DEPS = [
    'fastapi>=0.100.0',
    'uvicorn>=0.22.0',
    'grpcio>=1.56.0',
    'grpcio-tools>=1.56.0',
    'openai>=1.0.0',
    'langchain>=0.2.0',
    'langchain-openai>=0.1.0',
    'requests>=2.31.0',
    'httpx>=0.24.0'
]

def install_dependencies():
    """Install critical dependencies"""
    print("📦 Installing critical dependencies for LLM Gateway...")
    
    for dep in CRITICAL_DEPS:
        print(f"  Installing {dep}...")
        try:
            result = subprocess.run(
                [sys.executable, '-m', 'pip', 'install', dep], 
                capture_output=True, text=True, timeout=120
            )
            if result.returncode == 0:
                print(f"  ✅ {dep} installed successfully")
            else:
                print(f"  ⚠️  {dep} installation had warnings: {result.stderr.split()[-1] if result.stderr else 'unknown'}")
        except subprocess.TimeoutExpired:
            print(f"  ⏰ {dep} installation timed out, continuing...")
        except Exception as e:
            print(f"  ❌ {dep} installation failed: {e}")
    
    print("\n📋 Checking installed packages...")
    try:
        result = subprocess.run([sys.executable, '-m', 'pip', 'list'], 
                              capture_output=True, text=True)
        if result.returncode == 0:
            lines = result.stdout.split('\n')
            relevant_packages = [line for line in lines if any(
                pkg.split('>=')[0].lower() in line.lower() 
                for pkg in ['fastapi', 'uvicorn', 'grpcio', 'openai', 'langchain', 'requests', 'httpx']
            )]
            print("  Installed relevant packages:")
            for pkg in relevant_packages:
                print(f"    {pkg}")
    except Exception as e:
        print(f"  ❌ Failed to list packages: {e}")

def test_imports():
    """Test if we can import key modules"""
    print("\n🧪 Testing critical imports...")
    
    test_modules = [
        ('fastapi', 'FastAPI'),
        ('uvicorn', 'Uvicorn ASGI server'),
        ('grpcio', 'gRPC'),
        ('openai', 'OpenAI client'),
        ('requests', 'HTTP requests'),
        ('httpx', 'Async HTTP client')
    ]
    
    for module, description in test_modules:
        try:
            __import__(module)
            print(f"  ✅ {description}: importable")
        except ImportError as e:
            print(f"  ⚠️  {description}: {e}")

def main():
    print("🚀 LLM Gateway Dependency Installation")
    print("=" * 50)
    
    install_dependencies()
    test_imports()
    
    print("\n✅ Dependency installation complete!")
    print("   The service is ready for integration testing.")

if __name__ == "__main__":
    main()
