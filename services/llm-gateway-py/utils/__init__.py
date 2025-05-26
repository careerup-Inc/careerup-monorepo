"""Utility modules for the LLM Gateway service."""

from .logger import setup_logger, get_logger
from .vietnamese import (
    format_vietnamese_response,
    is_vietnamese_text,
    normalize_vietnamese_query,
    extract_vietnamese_keywords
)
from .metrics import MetricsCollector
from .helpers import (
    parse_json_safely,
    sanitize_text,
    truncate_text,
    extract_urls,
    validate_email
)

# Conditional import for security (requires PyJWT)
try:
    from .security import validate_api_key, generate_request_id
    _SECURITY_AVAILABLE = True
except ImportError:
    _SECURITY_AVAILABLE = False
    # Provide dummy functions
    def validate_api_key(api_key):
        return False
    def generate_request_id():
        import uuid
        return str(uuid.uuid4())

__all__ = [
    "setup_logger",
    "get_logger",
    "format_vietnamese_response",
    "is_vietnamese_text", 
    "normalize_vietnamese_query",
    "extract_vietnamese_keywords",
    "validate_api_key",
    "generate_request_id",
    "MetricsCollector",
    "parse_json_safely",
    "sanitize_text",
    "truncate_text",
    "extract_urls",
    "validate_email",
]
