"""Helper utilities for the LLM Gateway service."""

import json
import re
from typing import Any, Optional, List, Dict
import html
import urllib.parse
from datetime import datetime
import base64


def parse_json_safely(json_str: str, default: Any = None) -> Any:
    """Safely parse JSON string.
    
    Args:
        json_str: JSON string to parse
        default: Default value if parsing fails
        
    Returns:
        Parsed JSON object or default value
    """
    if not json_str or not isinstance(json_str, str):
        return default
    
    try:
        return json.loads(json_str.strip())
    except (json.JSONDecodeError, ValueError):
        return default


def sanitize_text(text: str, max_length: int = 10000) -> str:
    """Sanitize text input for safe processing.
    
    Args:
        text: Text to sanitize
        max_length: Maximum allowed length
        
    Returns:
        Sanitized text
    """
    if not isinstance(text, str):
        return ""
    
    # Truncate if too long
    if len(text) > max_length:
        text = text[:max_length] + "..."
    
    # HTML escape to prevent XSS
    text = html.escape(text)
    
    # Remove null bytes and control characters
    text = re.sub(r'[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]', '', text)
    
    # Normalize whitespace
    text = re.sub(r'\s+', ' ', text.strip())
    
    return text


def truncate_text(text: str, max_length: int, suffix: str = "...") -> str:
    """Truncate text to specified length.
    
    Args:
        text: Text to truncate
        max_length: Maximum length
        suffix: Suffix to add when truncating
        
    Returns:
        Truncated text
    """
    if not text or len(text) <= max_length:
        return text
    
    # Try to truncate at word boundary
    truncated = text[:max_length - len(suffix)]
    last_space = truncated.rfind(' ')
    
    if last_space > max_length * 0.7:  # If we found a space reasonably close
        truncated = truncated[:last_space]
    
    return truncated + suffix


def extract_urls(text: str) -> List[str]:
    """Extract URLs from text.
    
    Args:
        text: Text to extract URLs from
        
    Returns:
        List of extracted URLs
    """
    url_pattern = re.compile(
        r'http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+'
    )
    return url_pattern.findall(text)


def validate_email(email: str) -> bool:
    """Validate email address format.
    
    Args:
        email: Email address to validate
        
    Returns:
        True if email format is valid
    """
    if not email or not isinstance(email, str):
        return False
    
    email_pattern = re.compile(
        r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    )
    return bool(email_pattern.match(email.strip()))


def format_duration(seconds: float) -> str:
    """Format duration in seconds to human-readable string.
    
    Args:
        seconds: Duration in seconds
        
    Returns:
        Formatted duration string
    """
    if seconds < 1:
        return f"{seconds * 1000:.1f}ms"
    elif seconds < 60:
        return f"{seconds:.2f}s"
    elif seconds < 3600:
        minutes = int(seconds // 60)
        remaining_seconds = seconds % 60
        return f"{minutes}m {remaining_seconds:.1f}s"
    else:
        hours = int(seconds // 3600)
        remaining_minutes = int((seconds % 3600) // 60)
        return f"{hours}h {remaining_minutes}m"


def format_bytes(bytes_count: int) -> str:
    """Format byte count to human-readable string.
    
    Args:
        bytes_count: Number of bytes
        
    Returns:
        Formatted bytes string
    """
    for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
        if bytes_count < 1024.0:
            return f"{bytes_count:.1f}{unit}"
        bytes_count /= 1024.0
    return f"{bytes_count:.1f}PB"


def encode_base64(data: str) -> str:
    """Encode string to base64.
    
    Args:
        data: String to encode
        
    Returns:
        Base64 encoded string
    """
    return base64.b64encode(data.encode('utf-8')).decode('ascii')


def decode_base64(data: str) -> Optional[str]:
    """Decode base64 string.
    
    Args:
        data: Base64 string to decode
        
    Returns:
        Decoded string or None if invalid
    """
    try:
        return base64.b64decode(data).decode('utf-8')
    except Exception:
        return None


def slugify(text: str) -> str:
    """Convert text to URL-friendly slug.
    
    Args:
        text: Text to slugify
        
    Returns:
        URL-friendly slug
    """
    # Convert to lowercase and replace spaces with hyphens
    slug = re.sub(r'[^\w\s-]', '', text.lower())
    slug = re.sub(r'[-\s]+', '-', slug)
    return slug.strip('-')


def camel_to_snake(name: str) -> str:
    """Convert camelCase to snake_case.
    
    Args:
        name: CamelCase string
        
    Returns:
        snake_case string
    """
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def snake_to_camel(name: str) -> str:
    """Convert snake_case to camelCase.
    
    Args:
        name: snake_case string
        
    Returns:
        camelCase string
    """
    components = name.split('_')
    return components[0] + ''.join(x.title() for x in components[1:])


def deep_merge_dicts(dict1: Dict[str, Any], dict2: Dict[str, Any]) -> Dict[str, Any]:
    """Deep merge two dictionaries.
    
    Args:
        dict1: First dictionary
        dict2: Second dictionary (takes precedence)
        
    Returns:
        Merged dictionary
    """
    result = dict1.copy()
    
    for key, value in dict2.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge_dicts(result[key], value)
        else:
            result[key] = value
    
    return result


def flatten_dict(d: Dict[str, Any], parent_key: str = '', sep: str = '.') -> Dict[str, Any]:
    """Flatten nested dictionary.
    
    Args:
        d: Dictionary to flatten
        parent_key: Parent key prefix
        sep: Separator for nested keys
        
    Returns:
        Flattened dictionary
    """
    items = []
    for k, v in d.items():
        new_key = f"{parent_key}{sep}{k}" if parent_key else k
        if isinstance(v, dict):
            items.extend(flatten_dict(v, new_key, sep=sep).items())
        else:
            items.append((new_key, v))
    return dict(items)


def retry_with_backoff(
    func,
    max_retries: int = 3,
    backoff_factor: float = 1.0,
    exceptions: tuple = (Exception,)
):
    """Decorator for retrying function calls with exponential backoff.
    
    Args:
        func: Function to retry
        max_retries: Maximum number of retries
        backoff_factor: Backoff multiplier
        exceptions: Exception types to catch and retry
        
    Returns:
        Decorated function
    """
    import functools
    import time
    import random
    
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        for attempt in range(max_retries + 1):
            try:
                return func(*args, **kwargs)
            except exceptions as e:
                if attempt == max_retries:
                    raise e
                
                # Calculate backoff time with jitter
                backoff_time = backoff_factor * (2 ** attempt) + random.uniform(0, 1)
                time.sleep(backoff_time)
        
        return None
    
    return wrapper


def chunk_list(lst: List[Any], chunk_size: int) -> List[List[Any]]:
    """Split list into chunks of specified size.
    
    Args:
        lst: List to chunk
        chunk_size: Size of each chunk
        
    Returns:
        List of chunks
    """
    return [lst[i:i + chunk_size] for i in range(0, len(lst), chunk_size)]


def normalize_whitespace(text: str) -> str:
    """Normalize whitespace in text.
    
    Args:
        text: Text to normalize
        
    Returns:
        Text with normalized whitespace
    """
    if not text:
        return ""
    
    # Replace multiple whitespace characters with single space
    normalized = re.sub(r'\s+', ' ', text)
    
    # Remove leading/trailing whitespace
    return normalized.strip()


def extract_numbers(text: str) -> List[float]:
    """Extract numbers from text.
    
    Args:
        text: Text to extract numbers from
        
    Returns:
        List of extracted numbers
    """
    number_pattern = re.compile(r'-?\d+(?:\.\d+)?')
    matches = number_pattern.findall(text)
    return [float(match) for match in matches]


def is_valid_uuid(uuid_string: str) -> bool:
    """Check if string is a valid UUID.
    
    Args:
        uuid_string: String to check
        
    Returns:
        True if valid UUID format
    """
    uuid_pattern = re.compile(
        r'^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
        re.IGNORECASE
    )
    return bool(uuid_pattern.match(uuid_string))


def get_timestamp() -> str:
    """Get current timestamp in ISO format.
    
    Returns:
        ISO formatted timestamp string
    """
    return datetime.utcnow().isoformat()


def parse_timestamp(timestamp_str: str) -> Optional[datetime]:
    """Parse ISO timestamp string.
    
    Args:
        timestamp_str: ISO timestamp string
        
    Returns:
        Datetime object or None if invalid
    """
    try:
        return datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
    except (ValueError, AttributeError):
        return None


def extract_keywords(text: str, max_keywords: int = 10) -> List[str]:
    """Extract keywords from text using simple regex.
    
    Args:
        text: Text to extract keywords from
        max_keywords: Maximum number of keywords to return
        
    Returns:
        List of extracted keywords
    """
    if not isinstance(text, str) or not text.strip():
        return []
    
    # Simple keyword extraction using word patterns
    # Remove common stop words and extract meaningful words
    stop_words = {
        'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 
        'of', 'with', 'by', 'is', 'are', 'was', 'were', 'be', 'been', 'being',
        'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could',
        'should', 'may', 'might', 'can', 'this', 'that', 'these', 'those'
    }
    
    # Extract words (alphanumeric, 3+ characters)
    words = re.findall(r'\b[a-zA-Z0-9]{3,}\b', text.lower())
    
    # Filter out stop words and get unique keywords
    keywords = []
    seen = set()
    for word in words:
        if word not in stop_words and word not in seen:
            keywords.append(word)
            seen.add(word)
            if len(keywords) >= max_keywords:
                break
    
    return keywords


def validate_query(query: Any) -> bool:
    """Validate if a query is valid for processing.
    
    Args:
        query: Query to validate
        
    Returns:
        True if query is valid, False otherwise
    """
    if not query:
        return False
    
    if not isinstance(query, str):
        return False
    
    # Check if query has meaningful content
    cleaned_query = query.strip()
    if not cleaned_query or len(cleaned_query) < 2:
        return False
    
    # Check for reasonable length (not too long)
    if len(cleaned_query) > 10000:
        return False
    
    return True


def format_error_response(error_msg: str, error_code: str = "INTERNAL_ERROR") -> Dict[str, Any]:
    """Format error response consistently.
    
    Args:
        error_msg: Error message
        error_code: Error code
        
    Returns:
        Formatted error response dictionary
    """
    return {
        "error": {
            "message": error_msg,
            "code": error_code,
            "timestamp": get_timestamp()
        }
    }


def truncate_text(text: str, max_length: int = 1000, suffix: str = "...") -> str:
    """Truncate text to maximum length.
    
    Args:
        text: Text to truncate
        max_length: Maximum length
        suffix: Suffix to add when truncated
        
    Returns:
        Truncated text
    """
    if not isinstance(text, str):
        return ""
    
    if len(text) <= max_length:
        return text
    
    return text[:max_length - len(suffix)] + suffix
