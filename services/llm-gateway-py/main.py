import os
import sys
from pathlib import Path

# Add the proto directory to the Python path
project_root = Path(__file__).parent.parent.parent
proto_path = project_root / "proto"
sys.path.insert(0, str(proto_path))

import asyncio
import logging
from concurrent.futures import ThreadPoolExecutor
import grpc
from grpc_reflection.v1alpha import reflection
import uvicorn
import threading
from datetime import datetime

# Import configuration and utilities
from config.settings import get_settings
from utils.logger import setup_logger, get_logger
from utils.metrics import get_metrics_collector
from admin.api import get_admin_app

# Configure logging
settings = get_settings()
logger = setup_logger("llm-gateway-main", level=settings.log_level)

async def start_admin_server():
    """Start the FastAPI admin server."""
    try:
        admin_app = get_admin_app()
        config = uvicorn.Config(
            admin_app,
            host="0.0.0.0",
            port=settings.http_port,
            log_level=settings.log_level.lower(),
            access_log=True
        )
        server = uvicorn.Server(config)
        logger.info(f"Starting admin HTTP server on port {settings.http_port}")
        await server.serve()
    except Exception as e:
        logger.error(f"Failed to start admin server: {str(e)}", exc_info=True)


async def main():
    """Main entry point for the LLM Gateway Python service."""
    try:
        # Import after adding proto path
        from services.llm_service import LLMServicer
        from llm.v1 import llm_pb2_grpc
        
        logger.info("Initializing LLM Gateway Python service...")
        
        # Initialize metrics collector
        metrics_collector = get_metrics_collector()
        logger.info("Metrics collector initialized")
        
        # Create gRPC server
        server = grpc.aio.server(ThreadPoolExecutor(max_workers=settings.max_workers))
        
        # Create and register the LLM service
        llm_service = LLMServicer()
        llm_pb2_grpc.add_LLMServiceServicer_to_server(llm_service, server)
        
        # Add reflection for debugging
        from llm.v1 import llm_pb2
        SERVICE_NAMES = (
            llm_pb2.DESCRIPTOR.services_by_name['LLMService'].full_name,
            reflection.SERVICE_NAME,
        )
        reflection.enable_server_reflection(SERVICE_NAMES, server)
        
        # Configure server address
        grpc_address = f"[::]:{settings.grpc_port}"
        server.add_insecure_port(grpc_address)
        
        logger.info(f"Starting LLM Gateway Python gRPC server on {grpc_address}")
        
        # Start the gRPC server
        await server.start()
        
        # Start HTTP admin server in background
        admin_task = None
        if settings.enable_admin_api:
            admin_task = asyncio.create_task(start_admin_server())
            logger.info(f"Admin API will be available at http://localhost:{settings.http_port}/admin/docs")
        
        logger.info("LLM Gateway Python service is ready")
        logger.info(f"Service configuration: environment={settings.environment}, debug={settings.debug}")
        
        try:
            # Keep the server running
            await server.wait_for_termination()
        finally:
            # Clean shutdown
            if admin_task:
                admin_task.cancel()
                try:
                    await admin_task
                except asyncio.CancelledError:
                    pass
        
    except Exception as e:
        logger.error(f"Failed to start LLM Gateway Python service: {e}", exc_info=True)
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())
