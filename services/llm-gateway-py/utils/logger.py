"""Logging configuration and utilities for the LLM Gateway service."""

import logging
import logging.handlers
import sys
from typing import Optional
import json
from datetime import datetime
import traceback

# Global logger cache
_loggers = {}


class JSONFormatter(logging.Formatter):
    """Custom JSON formatter for structured logging."""
    
    def format(self, record):
        log_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = {
                "type": record.exc_info[0].__name__,
                "message": str(record.exc_info[1]),
                "traceback": traceback.format_exception(*record.exc_info)
            }
        
        # Add extra fields
        for key, value in record.__dict__.items():
            if key not in ['name', 'msg', 'args', 'levelname', 'levelno', 'pathname', 
                          'filename', 'module', 'lineno', 'funcName', 'created', 
                          'msecs', 'relativeCreated', 'thread', 'threadName', 
                          'processName', 'process', 'message', 'exc_info', 'exc_text', 'stack_info']:
                log_entry[key] = value
        
        return json.dumps(log_entry, ensure_ascii=False)


def setup_logger(
    name: str = "llm-gateway",
    level: str = "INFO",
    log_file: Optional[str] = None,
    max_bytes: int = 10 * 1024 * 1024,  # 10MB
    backup_count: int = 5,
    use_json: bool = True
) -> logging.Logger:
    """Set up a logger with file and console handlers.
    
    Args:
        name: Logger name
        level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_file: Path to log file (optional)
        max_bytes: Maximum bytes per log file before rotation
        backup_count: Number of backup files to keep
        use_json: Whether to use JSON formatting
        
    Returns:
        Configured logger instance
    """
    logger = logging.getLogger(name)
    
    # Avoid duplicate handlers
    if logger.handlers:
        return logger
    
    # Set level
    log_level = getattr(logging, level.upper(), logging.INFO)
    logger.setLevel(log_level)
    
    # Create formatter
    if use_json:
        formatter = JSONFormatter()
    else:
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(module)s:%(funcName)s:%(lineno)d - %(message)s'
        )
    
    # Console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(log_level)
    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)
    
    # File handler (optional)
    if log_file:
        file_handler = logging.handlers.RotatingFileHandler(
            log_file,
            maxBytes=max_bytes,
            backupCount=backup_count,
            encoding='utf-8'
        )
        file_handler.setLevel(log_level)
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)
    
    # Cache logger
    _loggers[name] = logger
    
    logger.info(f"Logger '{name}' initialized with level {level}")
    return logger


def get_logger(name: str = "llm-gateway") -> logging.Logger:
    """Get an existing logger or create a new one.
    
    Args:
        name: Logger name
        
    Returns:
        Logger instance
    """
    if name in _loggers:
        return _loggers[name]
    
    # Return root logger or create new one
    logger = logging.getLogger(name)
    if not logger.handlers:
        return setup_logger(name)
    return logger


class LoggerContextManager:
    """Context manager for adding request context to logs."""
    
    def __init__(self, logger: logging.Logger, **context):
        self.logger = logger
        self.context = context
        self.old_factory = None
    
    def __enter__(self):
        self.old_factory = logging.getLogRecordFactory()
        
        def record_factory(*args, **kwargs):
            record = self.old_factory(*args, **kwargs)
            for key, value in self.context.items():
                setattr(record, key, value)
            return record
        
        logging.setLogRecordFactory(record_factory)
        return self.logger
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        logging.setLogRecordFactory(self.old_factory)


def log_function_call(func):
    """Decorator to log function calls with parameters and execution time."""
    import functools
    import time
    
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        logger = get_logger(func.__module__)
        start_time = time.time()
        
        # Log function entry
        logger.debug(
            f"Calling {func.__name__}",
            extra={
                "function": func.__name__,
                "args_count": len(args),
                "kwargs_keys": list(kwargs.keys()),
                "event": "function_start"
            }
        )
        
        try:
            result = func(*args, **kwargs)
            execution_time = time.time() - start_time
            
            # Log successful completion
            logger.debug(
                f"Completed {func.__name__} in {execution_time:.3f}s",
                extra={
                    "function": func.__name__,
                    "execution_time": execution_time,
                    "event": "function_success"
                }
            )
            
            return result
            
        except Exception as e:
            execution_time = time.time() - start_time
            
            # Log exception
            logger.error(
                f"Error in {func.__name__} after {execution_time:.3f}s: {str(e)}",
                extra={
                    "function": func.__name__,
                    "execution_time": execution_time,
                    "error_type": type(e).__name__,
                    "error_message": str(e),
                    "event": "function_error"
                },
                exc_info=True
            )
            
            raise
    
    return wrapper
