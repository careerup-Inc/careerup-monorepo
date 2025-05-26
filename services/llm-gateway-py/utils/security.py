"""Security utilities for the LLM Gateway service."""

import hashlib
import hmac
import secrets
import uuid
from typing import Optional, Dict, Any
import time
import jwt
from datetime import datetime, timedelta


def generate_request_id() -> str:
    """Generate a unique request ID.
    
    Returns:
        Unique request ID string
    """
    return str(uuid.uuid4())


def validate_api_key(api_key: str, valid_keys: Optional[set] = None) -> bool:
    """Validate an API key.
    
    Args:
        api_key: API key to validate
        valid_keys: Set of valid API keys (if None, basic format validation only)
        
    Returns:
        True if API key is valid
    """
    if not api_key or not isinstance(api_key, str):
        return False
    
    # Basic format validation
    if len(api_key) < 16:
        return False
    
    # If valid keys provided, check against them
    if valid_keys is not None:
        return api_key in valid_keys
    
    # Basic format check (alphanumeric + some special chars)
    import re
    if not re.match(r'^[A-Za-z0-9_\-\.]+$', api_key):
        return False
    
    return True


def hash_sensitive_data(data: str, salt: Optional[str] = None) -> str:
    """Hash sensitive data using SHA-256.
    
    Args:
        data: Data to hash
        salt: Optional salt (if None, a random salt is generated)
        
    Returns:
        Hashed data as hex string
    """
    if salt is None:
        salt = secrets.token_hex(16)
    
    salted_data = f"{salt}{data}"
    return hashlib.sha256(salted_data.encode()).hexdigest()


def verify_signature(payload: str, signature: str, secret: str) -> bool:
    """Verify HMAC signature.
    
    Args:
        payload: Original payload
        signature: Signature to verify
        secret: Secret key for HMAC
        
    Returns:
        True if signature is valid
    """
    expected_signature = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    
    return hmac.compare_digest(signature, expected_signature)


def create_jwt_token(payload: Dict[str, Any], secret: str, expires_in: int = 3600) -> str:
    """Create a JWT token.
    
    Args:
        payload: Token payload
        secret: Secret key for signing
        expires_in: Token expiration time in seconds
        
    Returns:
        JWT token string
    """
    now = datetime.utcnow()
    payload.update({
        'iat': now,
        'exp': now + timedelta(seconds=expires_in),
        'jti': str(uuid.uuid4())  # JWT ID
    })
    
    return jwt.encode(payload, secret, algorithm='HS256')


def verify_jwt_token(token: str, secret: str) -> Optional[Dict[str, Any]]:
    """Verify and decode a JWT token.
    
    Args:
        token: JWT token to verify
        secret: Secret key for verification
        
    Returns:
        Decoded payload if valid, None otherwise
    """
    try:
        payload = jwt.decode(token, secret, algorithms=['HS256'])
        return payload
    except jwt.ExpiredSignatureError:
        return None
    except jwt.InvalidTokenError:
        return None


def sanitize_input(text: str, max_length: int = 10000) -> str:
    """Sanitize user input.
    
    Args:
        text: Input text to sanitize
        max_length: Maximum allowed length
        
    Returns:
        Sanitized text
    """
    if not isinstance(text, str):
        return ""
    
    # Truncate if too long
    if len(text) > max_length:
        text = text[:max_length]
    
    # Remove null bytes
    text = text.replace('\x00', '')
    
    # Remove control characters except common whitespace
    import re
    text = re.sub(r'[\x01-\x08\x0B\x0C\x0E-\x1F\x7F]', '', text)
    
    return text.strip()


def mask_sensitive_data(data: str, visible_chars: int = 4) -> str:
    """Mask sensitive data for logging.
    
    Args:
        data: Sensitive data to mask
        visible_chars: Number of characters to keep visible at the end
        
    Returns:
        Masked data string
    """
    if not data or len(data) <= visible_chars:
        return "*" * len(data) if data else ""
    
    return "*" * (len(data) - visible_chars) + data[-visible_chars:]


class RateLimiter:
    """Simple in-memory rate limiter."""
    
    def __init__(self, max_requests: int = 100, window_seconds: int = 60):
        """Initialize rate limiter.
        
        Args:
            max_requests: Maximum requests per window
            window_seconds: Time window in seconds
        """
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self.requests = {}  # {client_id: [(timestamp, count), ...]}
    
    def is_allowed(self, client_id: str) -> bool:
        """Check if request is allowed for client.
        
        Args:
            client_id: Unique client identifier
            
        Returns:
            True if request is allowed
        """
        now = time.time()
        window_start = now - self.window_seconds
        
        # Clean old entries
        if client_id in self.requests:
            self.requests[client_id] = [
                (timestamp, count) for timestamp, count in self.requests[client_id]
                if timestamp > window_start
            ]
        else:
            self.requests[client_id] = []
        
        # Count requests in current window
        total_requests = sum(count for _, count in self.requests[client_id])
        
        if total_requests >= self.max_requests:
            return False
        
        # Add current request
        self.requests[client_id].append((now, 1))
        return True
    
    def get_remaining_requests(self, client_id: str) -> int:
        """Get remaining requests for client.
        
        Args:
            client_id: Unique client identifier
            
        Returns:
            Number of remaining requests
        """
        now = time.time()
        window_start = now - self.window_seconds
        
        if client_id not in self.requests:
            return self.max_requests
        
        # Count current requests
        current_requests = sum(
            count for timestamp, count in self.requests[client_id]
            if timestamp > window_start
        )
        
        return max(0, self.max_requests - current_requests)


def generate_secure_token(length: int = 32) -> str:
    """Generate a cryptographically secure random token.
    
    Args:
        length: Token length in bytes
        
    Returns:
        Secure token as hex string
    """
    return secrets.token_hex(length)


def constant_time_compare(a: str, b: str) -> bool:
    """Compare two strings in constant time to prevent timing attacks.
    
    Args:
        a: First string
        b: Second string
        
    Returns:
        True if strings are equal
    """
    return hmac.compare_digest(a, b)


class SecurityHeaders:
    """Security headers for HTTP responses."""
    
    @staticmethod
    def get_default_headers() -> Dict[str, str]:
        """Get default security headers.
        
        Returns:
            Dictionary of security headers
        """
        return {
            'X-Content-Type-Options': 'nosniff',
            'X-Frame-Options': 'DENY',
            'X-XSS-Protection': '1; mode=block',
            'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
            'Content-Security-Policy': "default-src 'self'",
            'Referrer-Policy': 'strict-origin-when-cross-origin'
        }


class SecurityManager:
    """Security manager for the LLM Gateway service."""
    
    def __init__(self, jwt_secret: Optional[str] = None, rate_limit_requests: int = 100):
        """Initialize security manager.
        
        Args:
            jwt_secret: Secret for JWT token signing
            rate_limit_requests: Maximum requests per minute
        """
        self.jwt_secret = jwt_secret or generate_secure_token()
        self.rate_limiter = RateLimiter(max_requests=rate_limit_requests, window_seconds=60)
        self.security_headers = SecurityHeaders()
    
    def generate_api_key(self) -> str:
        """Generate a new API key.
        
        Returns:
            New API key
        """
        return generate_secure_token(32)
    
    def validate_request(self, api_key: str, client_id: str) -> bool:
        """Validate a request.
        
        Args:
            api_key: API key to validate
            client_id: Client identifier for rate limiting
            
        Returns:
            True if request is valid
        """
        # Validate API key format
        if not validate_api_key(api_key):
            return False
        
        # Check rate limit
        if not self.rate_limiter.is_allowed(client_id):
            return False
        
        return True
    
    def validate_api_key(self, api_key: str) -> bool:
        """Validate an API key (alias for validate_api_key function).
        
        Args:
            api_key: API key to validate
            
        Returns:
            True if API key is valid
        """
        return validate_api_key(api_key)
    
    def check_rate_limit(self, client_id: str) -> bool:
        """Check if client is within rate limits.
        
        Args:
            client_id: Client identifier
            
        Returns:
            True if client is allowed to make request
        """
        return self.rate_limiter.is_allowed(client_id)
    
    def create_token(self, payload: Dict[str, Any], expires_in: int = 3600) -> str:
        """Create a JWT token.
        
        Args:
            payload: Token payload
            expires_in: Expiration time in seconds
            
        Returns:
            JWT token
        """
        return create_jwt_token(payload, self.jwt_secret, expires_in)
    
    def verify_token(self, token: str) -> Optional[Dict[str, Any]]:
        """Verify a JWT token.
        
        Args:
            token: JWT token to verify
            
        Returns:
            Decoded payload if valid
        """
        return verify_jwt_token(token, self.jwt_secret)
    
    def sanitize_user_input(self, text: str) -> str:
        """Sanitize user input.
        
        Args:
            text: Input text to sanitize
            
        Returns:
            Sanitized text
        """
        return sanitize_input(text)
    
    def mask_sensitive_info(self, data: str) -> str:
        """Mask sensitive information for logging.
        
        Args:
            data: Sensitive data to mask
            
        Returns:
            Masked data
        """
        return mask_sensitive_data(data)
