"""Vietnamese language utilities for the LLM Gateway service."""

import re
import unicodedata
from typing import List, Optional, Dict, Any
import json


# Vietnamese vowels and consonants for text detection
VIETNAMESE_CHARS = set("àáãạảăắằẵặẳâấầẫậẩèéẹẻẽêềếễệểìíịỉĩòóõọỏôốồỗộổơớờỡợởùúũụủưứừữựửỳýỵỷỹđĐ")
VIETNAMESE_DIACRITICS = "àáãạảăắằẵặẳâấầẫậẩèéẹẻẽêềếễệểìíịỉĩòóõọỏôốồỗộổơớờỡợởùúũụủưứừữựửỳýỵỷỹđ"

# Common Vietnamese words for language detection
VIETNAMESE_COMMON_WORDS = {
    "và", "của", "có", "là", "trong", "để", "với", "được", "các", "một",
    "từ", "cho", "về", "như", "khi", "trên", "tại", "người", "này", "đó",
    "việc", "thì", "cũng", "theo", "sẽ", "đã", "sau", "nếu", "bằng", "những",
    "tôi", "bạn", "anh", "chị", "em", "chúng", "ta", "họ", "nó", "gì",
    "làm", "đi", "nói", "biết", "thấy", "nghĩ", "cần", "muốn", "không"
}

# Vietnamese formatting patterns
VIETNAMESE_SENTENCE_ENDERS = [".", "!", "?", "...", "…"]
VIETNAMESE_PUNCTUATION = [".", ",", "!", "?", ":", ";", "-", "–", "—", "(", ")", "[", "]", "{", "}", "\"", "'", "…"]


def is_vietnamese_text(text: str, threshold: float = 0.1) -> bool:
    """Check if text contains Vietnamese characters.
    
    Args:
        text: Text to check
        threshold: Minimum ratio of Vietnamese characters to consider text Vietnamese
        
    Returns:
        True if text appears to be Vietnamese
    """
    if not text or not text.strip():
        return False
    
    # Remove whitespace and punctuation for analysis
    clean_text = re.sub(r'[^\w]', '', text.lower())
    if not clean_text:
        return False
    
    # Count Vietnamese characters
    vietnamese_char_count = sum(1 for char in clean_text if char in VIETNAMESE_CHARS)
    vietnamese_ratio = vietnamese_char_count / len(clean_text)
    
    # Check for common Vietnamese words
    words = text.lower().split()
    vietnamese_word_count = sum(1 for word in words if word in VIETNAMESE_COMMON_WORDS)
    vietnamese_word_ratio = vietnamese_word_count / len(words) if words else 0
    
    # Consider it Vietnamese if either character ratio or word ratio exceeds threshold
    return vietnamese_ratio >= threshold or vietnamese_word_ratio >= threshold * 2


# Alias for the test script
def detect_vietnamese(text: str) -> bool:
    """Detect if text is Vietnamese (alias for is_vietnamese_text).
    
    Args:
        text: Text to check
        
    Returns:
        True if text appears to be Vietnamese
    """
    return is_vietnamese_text(text)


def normalize_vietnamese_text(text: str) -> str:
    """Normalize Vietnamese text for better processing.
    
    Args:
        text: Original text
        
    Returns:
        Normalized text
    """
    if not text:
        return ""
    
    # Remove extra whitespace
    normalized = re.sub(r'\s+', ' ', text.strip())
    
    # Normalize unicode characters
    normalized = unicodedata.normalize('NFC', normalized)
    
    # Fix common Vietnamese typing errors
    replacements = {
        'òa': 'oà',  # Common typing error
        'úy': 'uý',  # Common typing error
        'ùy': 'uỳ',  # Common typing error
    }
    
    for wrong, correct in replacements.items():
        normalized = normalized.replace(wrong, correct)
    
    return normalized
    vietnamese_word_ratio = vietnamese_word_count / len(words) if words else 0
    
    # Consider it Vietnamese if either character ratio or word ratio exceeds threshold
    return vietnamese_ratio >= threshold or vietnamese_word_ratio >= threshold * 2


def normalize_vietnamese_query(query: str) -> str:
    """Normalize Vietnamese text for better processing.
    
    Args:
        query: Original query text
        
    Returns:
        Normalized query text
    """
    if not query:
        return ""
    
    # Remove extra whitespace
    normalized = re.sub(r'\s+', ' ', query.strip())
    
    # Normalize unicode characters
    normalized = unicodedata.normalize('NFC', normalized)
    
    # Fix common Vietnamese typing errors
    replacements = {
        'òa': 'oà',  # Common typing error
        'úy': 'uý',  # Common typing error
        'ùy': 'uỳ',  # Common typing error
    }
    
    for wrong, correct in replacements.items():
        normalized = normalized.replace(wrong, correct)
    
    return normalized


def extract_vietnamese_keywords(text: str) -> List[str]:
    """Extract meaningful Vietnamese keywords from text.
    
    Args:
        text: Text to extract keywords from
        
    Returns:
        List of extracted keywords
    """
    if not text:
        return []
    
    # Normalize text
    normalized = normalize_vietnamese_query(text.lower())
    
    # Split into words and filter
    words = re.findall(r'\b\w+\b', normalized)
    
    # Remove common stop words and keep meaningful words
    stop_words = {
        'và', 'của', 'có', 'là', 'trong', 'để', 'với', 'được', 'các', 'một',
        'từ', 'cho', 'về', 'như', 'khi', 'trên', 'tại', 'này', 'đó', 'việc',
        'thì', 'cũng', 'theo', 'sẽ', 'đã', 'sau', 'nếu', 'bằng', 'những',
        'rất', 'nhiều', 'lại', 'còn', 'chỉ', 'đều', 'đang', 'vào', 'ra',
        'lên', 'xuống', 'mà', 'nhưng', 'hoặc', 'hay', 'vì', 'do', 'nên'
    }
    
    keywords = [word for word in words if len(word) > 1 and word not in stop_words]
    
    # Remove duplicates while preserving order
    unique_keywords = []
    for keyword in keywords:
        if keyword not in unique_keywords:
            unique_keywords.append(keyword)
    
    return unique_keywords


def format_vietnamese_response(response: str, style: str = "formal") -> str:
    """Format response text according to Vietnamese language conventions.
    
    Args:
        response: Response text to format
        style: Formatting style ("formal", "casual", "professional")
        
    Returns:
        Formatted response text
    """
    if not response:
        return ""
    
    # Normalize unicode
    formatted = unicodedata.normalize('NFC', response)
    
    # Fix spacing around punctuation
    formatted = re.sub(r'\s+([,.!?;:])', r'\1', formatted)
    formatted = re.sub(r'([,.!?;:])\s*', r'\1 ', formatted)
    formatted = re.sub(r'\s+', ' ', formatted)
    
    # Ensure proper sentence capitalization
    sentences = re.split(r'([.!?]+)', formatted)
    formatted_sentences = []
    
    for i, sentence in enumerate(sentences):
        if i % 2 == 0:  # Actual sentence content
            sentence = sentence.strip()
            if sentence:
                # Capitalize first letter
                sentence = sentence[0].upper() + sentence[1:] if len(sentence) > 1 else sentence.upper()
        formatted_sentences.append(sentence)
    
    formatted = ''.join(formatted_sentences)
    
    # Style-specific adjustments
    if style == "formal":
        # Use formal pronouns and expressions
        formatted = formatted.replace(' tôi ', ' tôi ')
        formatted = formatted.replace(' mình ', ' tôi ')
        formatted = formatted.replace(' tao ', ' tôi ')
    elif style == "casual":
        # Keep casual tone
        pass
    elif style == "professional":
        # Add professional courtesies
        if not formatted.startswith(('Xin chào', 'Chào', 'Kính chào')):
            formatted = f"Xin chào! {formatted}"
        if not formatted.endswith(('Cảm ơn!', 'Trân trọng!', 'Chúc bạn một ngày tốt lành!')):
            formatted = f"{formatted} Cảm ơn bạn!"
    
    return formatted.strip()


def detect_vietnamese_intent(query: str) -> Dict[str, Any]:
    """Detect intent and entities from Vietnamese query.
    
    Args:
        query: Vietnamese query to analyze
        
    Returns:
        Dictionary containing intent and entities
    """
    normalized_query = normalize_vietnamese_query(query.lower())
    
    # Define intent patterns
    intent_patterns = {
        'question': [
            r'\b(gì|sao|thế nào|tại sao|vì sao|khi nào|ở đâu|ai|như thế nào)\b',
            r'^(có|được|phải|nên|cần)\b.*\?',
            r'\?$'
        ],
        'request': [
            r'\b(giúp|hãy|xin|vui lòng|có thể|làm ơn)\b',
            r'\b(tìm|tìm kiếm|cho biết|cung cấp|đưa ra)\b'
        ],
        'greeting': [
            r'\b(chào|xin chào|hello|hi|chúc)\b',
            r'\b(buổi sáng|buổi chiều|buổi tối|ngày mới)\b'
        ],
        'thanks': [
            r'\b(cảm ơn|cám ơn|thanks|thank you|merci)\b',
            r'\b(tuyệt vời|tốt quá|hay quá)\b'
        ],
        'complaint': [
            r'\b(tệ|kém|không tốt|không hay|chán|thất vọng)\b',
            r'\b(sai|lỗi|không đúng|không chính xác)\b'
        ]
    }
    
    detected_intent = 'unknown'
    confidence = 0.0
    
    for intent, patterns in intent_patterns.items():
        for pattern in patterns:
            if re.search(pattern, normalized_query):
                detected_intent = intent
                confidence = 0.8  # Simple confidence score
                break
        if detected_intent != 'unknown':
            break
    
    # Extract entities (simple keyword extraction)
    keywords = extract_vietnamese_keywords(query)
    
    return {
        'intent': detected_intent,
        'confidence': confidence,
        'keywords': keywords,
        'is_vietnamese': is_vietnamese_text(query),
        'query_length': len(query),
        'word_count': len(query.split())
    }


def create_vietnamese_prompt_template(base_prompt: str, context: str = "") -> str:
    """Create a Vietnamese-optimized prompt template.
    
    Args:
        base_prompt: Base prompt in English or Vietnamese
        context: Additional context to include
        
    Returns:
        Vietnamese-optimized prompt template
    """
    vietnamese_instructions = """
Hãy trả lời bằng tiếng Việt một cách tự nhiên và chính xác. 
Sử dụng ngôn ngữ lịch sự và phù hợp với văn hóa Việt Nam.
Nếu thông tin không đầy đủ, hãy thừa nhận và đề xuất cách tìm hiểu thêm.
"""
    
    if context:
        context_section = f"\nThông tin tham khảo:\n{context}\n"
    else:
        context_section = ""
    
    full_prompt = f"""
{vietnamese_instructions}

{base_prompt}
{context_section}

Hãy trả lời một cách chi tiết và hữu ích:
"""
    
    return full_prompt.strip()


def validate_vietnamese_response(response: str) -> Dict[str, Any]:
    """Validate if a response is appropriate for Vietnamese users.
    
    Args:
        response: Response to validate
        
    Returns:
        Validation results
    """
    if not response:
        return {
            'is_valid': False,
            'issues': ['Empty response'],
            'suggestions': ['Provide a meaningful response']
        }
    
    issues = []
    suggestions = []
    
    # Check if response is in Vietnamese when query was Vietnamese
    if not is_vietnamese_text(response):
        issues.append('Response not in Vietnamese')
        suggestions.append('Respond in Vietnamese for Vietnamese queries')
    
    # Check response length
    if len(response) < 10:
        issues.append('Response too short')
        suggestions.append('Provide more detailed information')
    
    # Check for proper formatting
    if not response.strip().endswith(tuple(VIETNAMESE_SENTENCE_ENDERS)):
        issues.append('Missing proper sentence ending')
        suggestions.append('End response with appropriate punctuation')
    
    # Check for politeness markers
    polite_markers = ['xin', 'cảm ơn', 'vui lòng', 'ạ', 'ơi', 'nhé']
    has_polite_markers = any(marker in response.lower() for marker in polite_markers)
    
    if not has_polite_markers and len(response) > 50:
        suggestions.append('Consider adding polite expressions for better tone')
    
    return {
        'is_valid': len(issues) == 0,
        'issues': issues,
        'suggestions': suggestions,
        'character_count': len(response),
        'word_count': len(response.split()),
        'has_vietnamese_chars': any(char in VIETNAMESE_CHARS for char in response)
    }
