"""Admin module for FastAPI endpoints."""

try:
    from .api import get_admin_app, create_admin_app
    __all__ = ["get_admin_app", "create_admin_app"]
except ImportError:
    # FastAPI not available, provide dummy functions
    def get_admin_app():
        """Dummy admin app when FastAPI is not available."""
        return None
    
    def create_admin_app():
        """Dummy admin app creation when FastAPI is not available."""
        return None
    
    __all__ = ["get_admin_app", "create_admin_app"]
