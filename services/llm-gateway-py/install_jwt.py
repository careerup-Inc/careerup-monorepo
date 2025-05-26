#!/usr/bin/env python3
"""
Simple script to install PyJWT and test the security module
"""

import subprocess
import sys
import os

def install_pyjwt():
    """Install PyJWT using pip"""
    try:
        result = subprocess.run([sys.executable, '-m', 'pip', 'install', 'PyJWT'], 
                              capture_output=True, text=True)
        if result.returncode == 0:
            print("✅ PyJWT installed successfully")
            return True
        else:
            print(f"❌ Failed to install PyJWT: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ Exception installing PyJWT: {e}")
        return False

def test_jwt_import():
    """Test if JWT can be imported"""
    try:
        import jwt
        print(f"✅ PyJWT import successful, version: {jwt.__version__}")
        return True
    except ImportError as e:
        print(f"❌ PyJWT import failed: {e}")
        return False

def main():
    print("🔧 Installing and testing PyJWT...")
    
    # Try to import first
    if test_jwt_import():
        print("✅ PyJWT already available")
        return 0
    
    # Install if not available
    if install_pyjwt():
        if test_jwt_import():
            print("✅ PyJWT installation and test complete")
            return 0
    
    print("❌ Failed to install or import PyJWT")
    return 1

if __name__ == "__main__":
    sys.exit(main())
