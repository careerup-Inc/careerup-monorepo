"""Metrics collection utilities for the LLM Gateway service."""

import time
from typing import Dict, Any, Optional, List
from collections import defaultdict, deque
from dataclasses import dataclass, field
from datetime import datetime, timedelta
import threading
import json


@dataclass
class RequestMetrics:
    """Metrics for a single request."""
    request_id: str
    timestamp: datetime
    method: str
    duration: float
    success: bool
    error_type: Optional[str] = None
    tokens_used: Optional[int] = None
    model_used: Optional[str] = None
    query_type: Optional[str] = None
    language: Optional[str] = None


@dataclass
class AggregatedMetrics:
    """Aggregated metrics over a time period."""
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    average_duration: float = 0.0
    total_tokens: int = 0
    error_distribution: Dict[str, int] = field(default_factory=dict)
    method_distribution: Dict[str, int] = field(default_factory=dict)
    model_distribution: Dict[str, int] = field(default_factory=dict)
    query_type_distribution: Dict[str, int] = field(default_factory=dict)
    language_distribution: Dict[str, int] = field(default_factory=dict)


class MetricsCollector:
    """Thread-safe metrics collector for the LLM Gateway service."""
    
    def __init__(self, max_history: int = 10000, aggregation_window: int = 300):
        """Initialize metrics collector.
        
        Args:
            max_history: Maximum number of individual metrics to keep
            aggregation_window: Time window for aggregation in seconds
        """
        self.max_history = max_history
        self.aggregation_window = aggregation_window
        self.metrics_history: deque = deque(maxlen=max_history)
        self.aggregated_metrics: Dict[str, AggregatedMetrics] = {}
        self.lock = threading.Lock()
        
        # Real-time counters
        self.total_requests = 0
        self.total_errors = 0
        self.total_tokens = 0
        self.total_duration = 0.0
        
        # Rate tracking
        self.request_times = deque(maxlen=100)  # Last 100 requests for rate calculation
    
    def record_request(
        self,
        request_id: str,
        method: str,
        duration: float,
        success: bool,
        error_type: Optional[str] = None,
        tokens_used: Optional[int] = None,
        model_used: Optional[str] = None,
        query_type: Optional[str] = None,
        language: Optional[str] = None
    ):
        """Record metrics for a single request.
        
        Args:
            request_id: Unique request identifier
            method: gRPC method name
            duration: Request duration in seconds
            success: Whether request was successful
            error_type: Type of error if failed
            tokens_used: Number of tokens consumed
            model_used: LLM model used
            query_type: Type of query (rag, web_search, direct)
            language: Detected language
        """
        now = datetime.utcnow()
        
        metrics = RequestMetrics(
            request_id=request_id,
            timestamp=now,
            method=method,
            duration=duration,
            success=success,
            error_type=error_type,
            tokens_used=tokens_used,
            model_used=model_used,
            query_type=query_type,
            language=language
        )
        
        with self.lock:
            # Store individual metrics
            self.metrics_history.append(metrics)
            
            # Update real-time counters
            self.total_requests += 1
            if not success:
                self.total_errors += 1
            if tokens_used:
                self.total_tokens += tokens_used
            self.total_duration += duration
            
            # Track request rate
            self.request_times.append(time.time())
            
            # Update aggregated metrics
            self._update_aggregated_metrics(metrics)
    
    def _update_aggregated_metrics(self, metrics: RequestMetrics):
        """Update aggregated metrics with new request data."""
        # Determine time bucket
        bucket_time = metrics.timestamp.replace(
            second=0, microsecond=0
        ) - timedelta(
            minutes=metrics.timestamp.minute % (self.aggregation_window // 60)
        )
        bucket_key = bucket_time.isoformat()
        
        if bucket_key not in self.aggregated_metrics:
            self.aggregated_metrics[bucket_key] = AggregatedMetrics()
        
        agg = self.aggregated_metrics[bucket_key]
        agg.total_requests += 1
        
        if metrics.success:
            agg.successful_requests += 1
        else:
            agg.failed_requests += 1
            if metrics.error_type:
                agg.error_distribution[metrics.error_type] = (
                    agg.error_distribution.get(metrics.error_type, 0) + 1
                )
        
        # Update average duration
        total_duration = agg.average_duration * (agg.total_requests - 1) + metrics.duration
        agg.average_duration = total_duration / agg.total_requests
        
        # Update distributions
        agg.method_distribution[metrics.method] = (
            agg.method_distribution.get(metrics.method, 0) + 1
        )
        
        if metrics.tokens_used:
            agg.total_tokens += metrics.tokens_used
        
        if metrics.model_used:
            agg.model_distribution[metrics.model_used] = (
                agg.model_distribution.get(metrics.model_used, 0) + 1
            )
        
        if metrics.query_type:
            agg.query_type_distribution[metrics.query_type] = (
                agg.query_type_distribution.get(metrics.query_type, 0) + 1
            )
        
        if metrics.language:
            agg.language_distribution[metrics.language] = (
                agg.language_distribution.get(metrics.language, 0) + 1
            )
        
        # Clean old aggregated metrics (keep last 24 hours)
        cutoff_time = datetime.utcnow() - timedelta(hours=24)
        old_buckets = [
            key for key in self.aggregated_metrics.keys()
            if datetime.fromisoformat(key) < cutoff_time
        ]
        for key in old_buckets:
            del self.aggregated_metrics[key]
    
    def get_current_stats(self) -> Dict[str, Any]:
        """Get current statistics.
        
        Returns:
            Dictionary with current statistics
        """
        with self.lock:
            success_rate = (
                (self.total_requests - self.total_errors) / self.total_requests * 100
                if self.total_requests > 0 else 0
            )
            
            average_duration = (
                self.total_duration / self.total_requests
                if self.total_requests > 0 else 0
            )
            
            # Calculate requests per minute
            now = time.time()
            recent_requests = [
                t for t in self.request_times
                if now - t <= 60  # Last minute
            ]
            requests_per_minute = len(recent_requests)
            
            return {
                'total_requests': self.total_requests,
                'total_errors': self.total_errors,
                'success_rate': round(success_rate, 2),
                'average_duration': round(average_duration, 3),
                'total_tokens': self.total_tokens,
                'requests_per_minute': requests_per_minute,
                'timestamp': datetime.utcnow().isoformat()
            }
    
    def get_recent_metrics(self, minutes: int = 60) -> List[RequestMetrics]:
        """Get recent request metrics.
        
        Args:
            minutes: Number of minutes to look back
            
        Returns:
            List of recent request metrics
        """
        cutoff_time = datetime.utcnow() - timedelta(minutes=minutes)
        
        with self.lock:
            return [
                metrics for metrics in self.metrics_history
                if metrics.timestamp >= cutoff_time
            ]
    
    def get_aggregated_metrics(self, hours: int = 24) -> Dict[str, AggregatedMetrics]:
        """Get aggregated metrics for specified time period.
        
        Args:
            hours: Number of hours to look back
            
        Returns:
            Dictionary of aggregated metrics by time bucket
        """
        cutoff_time = datetime.utcnow() - timedelta(hours=hours)
        
        with self.lock:
            return {
                key: metrics for key, metrics in self.aggregated_metrics.items()
                if datetime.fromisoformat(key) >= cutoff_time
            }
    
    def get_error_summary(self, hours: int = 24) -> Dict[str, Any]:
        """Get error summary for specified time period.
        
        Args:
            hours: Number of hours to look back
            
        Returns:
            Error summary statistics
        """
        recent_metrics = self.get_recent_metrics(hours * 60)
        
        error_counts = defaultdict(int)
        error_methods = defaultdict(set)
        total_errors = 0
        
        for metrics in recent_metrics:
            if not metrics.success and metrics.error_type:
                error_counts[metrics.error_type] += 1
                error_methods[metrics.error_type].add(metrics.method)
                total_errors += 1
        
        return {
            'total_errors': total_errors,
            'error_types': dict(error_counts),
            'error_methods': {
                error_type: list(methods)
                for error_type, methods in error_methods.items()
            },
            'error_rate': (
                total_errors / len(recent_metrics) * 100
                if recent_metrics else 0
            )
        }
    
    def export_metrics(self, format_type: str = "json") -> str:
        """Export metrics in specified format.
        
        Args:
            format_type: Export format ("json" or "prometheus")
            
        Returns:
            Exported metrics string
        """
        if format_type == "json":
            return self._export_json()
        elif format_type == "prometheus":
            return self._export_prometheus()
        else:
            raise ValueError(f"Unsupported format: {format_type}")
    
    def _export_json(self) -> str:
        """Export metrics as JSON."""
        with self.lock:
            data = {
                'current_stats': self.get_current_stats(),
                'recent_errors': self.get_error_summary(),
                'aggregated_metrics': {
                    key: {
                        'total_requests': agg.total_requests,
                        'successful_requests': agg.successful_requests,
                        'failed_requests': agg.failed_requests,
                        'average_duration': agg.average_duration,
                        'total_tokens': agg.total_tokens,
                        'error_distribution': agg.error_distribution,
                        'method_distribution': agg.method_distribution,
                        'model_distribution': agg.model_distribution,
                        'query_type_distribution': agg.query_type_distribution,
                        'language_distribution': agg.language_distribution,
                    }
                    for key, agg in self.aggregated_metrics.items()
                }
            }
            return json.dumps(data, indent=2, ensure_ascii=False)
    
    def _export_prometheus(self) -> str:
        """Export metrics in Prometheus format."""
        stats = self.get_current_stats()
        error_summary = self.get_error_summary()
        
        lines = [
            f"# HELP llm_gateway_requests_total Total number of requests",
            f"# TYPE llm_gateway_requests_total counter",
            f"llm_gateway_requests_total {stats['total_requests']}",
            f"",
            f"# HELP llm_gateway_errors_total Total number of errors",
            f"# TYPE llm_gateway_errors_total counter",
            f"llm_gateway_errors_total {stats['total_errors']}",
            f"",
            f"# HELP llm_gateway_success_rate Success rate percentage",
            f"# TYPE llm_gateway_success_rate gauge",
            f"llm_gateway_success_rate {stats['success_rate']}",
            f"",
            f"# HELP llm_gateway_duration_average Average request duration",
            f"# TYPE llm_gateway_duration_average gauge",
            f"llm_gateway_duration_average {stats['average_duration']}",
            f"",
            f"# HELP llm_gateway_tokens_total Total tokens consumed",
            f"# TYPE llm_gateway_tokens_total counter",
            f"llm_gateway_tokens_total {stats['total_tokens']}",
            f"",
            f"# HELP llm_gateway_requests_per_minute Current requests per minute",
            f"# TYPE llm_gateway_requests_per_minute gauge",
            f"llm_gateway_requests_per_minute {stats['requests_per_minute']}",
        ]
        
        return "\n".join(lines)
    
    def reset_metrics(self):
        """Reset all metrics (use with caution)."""
        with self.lock:
            self.metrics_history.clear()
            self.aggregated_metrics.clear()
            self.total_requests = 0
            self.total_errors = 0
            self.total_tokens = 0
            self.total_duration = 0.0
            self.request_times.clear()


# Global metrics collector instance
_metrics_collector = None


def get_metrics_collector() -> MetricsCollector:
    """Get the global metrics collector instance.
    
    Returns:
        MetricsCollector instance
    """
    global _metrics_collector
    if _metrics_collector is None:
        _metrics_collector = MetricsCollector()
    return _metrics_collector


# Alias for backward compatibility
class ServiceMetrics(MetricsCollector):
    """Simplified metrics interface for backward compatibility."""
    
    def __init__(self):
        super().__init__()
        self.counters = {}
        self.latencies = {}
    
    def increment_counter(self, name: str, value: int = 1):
        """Increment a named counter.
        
        Args:
            name: Counter name
            value: Value to increment by
        """
        self.counters[name] = self.counters.get(name, 0) + value
    
    def record_latency(self, operation: str, duration: float):
        """Record operation latency.
        
        Args:
            operation: Operation name
            duration: Duration in seconds
        """
        if operation not in self.latencies:
            self.latencies[operation] = []
        self.latencies[operation].append(duration)
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get metrics summary.
        
        Returns:
            Dictionary of metrics
        """
        metrics = dict(self.counters)
        
        # Add latency averages
        for operation, durations in self.latencies.items():
            avg_latency = sum(durations) / len(durations) if durations else 0
            metrics[f"{operation}_avg_latency"] = round(avg_latency, 3)
            metrics[f"{operation}_count"] = len(durations)
        
        return metrics
