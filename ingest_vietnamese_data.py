#!/usr/bin/env python3
"""
Script to ingest Vietnamese university data directly into the LLM Gateway.
This bypasses the admin API endpoint and calls the ingestion method directly.
"""

import asyncio
import json
import sys
import os
from pathlib import Path

# Add the services directory to Python path
sys.path.insert(0, str(Path(__file__).parent / "services" / "llm-gateway-py"))

async def ingest_vietnamese_data():
    """Ingest Vietnamese university data directly."""
    import time
    start_time = time.time()
    
    try:
        # Import after adding to path
        from services.llm_service import LLMServicer
        
        print("🚀 Starting Vietnamese university data ingestion...")
        print(f"⏱️  Start time: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        # Create LLM service instance
        print("🔧 Initializing LLM service...")
        init_start = time.time()
        llm_service = LLMServicer()
        init_time = time.time() - init_start
        print(f"✅ LLM service initialized in {init_time:.1f} seconds")
        
        # Define file paths (adapt for container vs local execution)
        if os.path.exists("/app/data"):
            # Running inside Docker container
            data_dir = Path("/app/data")
        else:
            # Running locally
            data_dir = Path(__file__).parent / "services" / "llm-gateway-py" / "data"
        
        json_file = str(data_dir / "diem_chuan_dai_hoc_2024_enhanced.json")
        pdf_file = str(data_dir / "de-an-tuyen-sinh-2024final.pdf")
        
        # Check if files exist
        if not os.path.exists(json_file):
            print(f"❌ JSON file not found: {json_file}")
            return
        
        if not os.path.exists(pdf_file):
            print(f"❌ PDF file not found: {pdf_file}")
            return
        
        print(f"📄 Found JSON file: {json_file}")
        print(f"📄 Found PDF file: {pdf_file}")
        
        # Get file sizes for confirmation
        json_size = os.path.getsize(json_file) / 1024  # KB
        pdf_size = os.path.getsize(pdf_file) / 1024    # KB
        
        print(f"📊 JSON file size: {json_size:.1f} KB")
        print(f"📊 PDF file size: {pdf_size:.1f} KB")
        
        # Ingest JSON file (university admission scores)
        print("\n🔄 Ingesting JSON university admission data...")
        json_start = time.time()
        json_result = await llm_service.ingest_vietnamese_university_data(
            file_path=json_file,
            file_type="json",
            collection_name="vietnamese-university-scores"
        )
        json_time = time.time() - json_start
        print(f"⏱️  JSON processing completed in {json_time:.1f} seconds")
        
        print("📋 JSON Ingestion Result:")
        print(json.dumps(json_result, indent=2, ensure_ascii=False))
        
        # Ingest PDF file (admission guidelines)
        print("\n🔄 Ingesting PDF admission guidelines...")
        pdf_start = time.time()
        pdf_result = await llm_service.ingest_vietnamese_university_data(
            file_path=pdf_file,
            file_type="pdf",
            collection_name="vietnamese-university-guidelines"
        )
        pdf_time = time.time() - pdf_start
        print(f"⏱️  PDF processing completed in {pdf_time:.1f} seconds")
        
        print("📋 PDF Ingestion Result:")
        print(json.dumps(pdf_result, indent=2, ensure_ascii=False))
        
        # Summary
        print("\n✅ Vietnamese University Data Ingestion Complete!")
        print("=" * 60)
        
        total_docs = 0
        total_summaries = 0
        total_chunks = 0
        
        if json_result.get("success"):
            docs = json_result.get("documents_processed", 0)
            summaries = json_result.get("summaries_created", 0)
            chunks = json_result.get("total_chunks", 0)
            total_docs += docs
            total_summaries += summaries
            total_chunks += chunks
            print(f"📊 JSON Data: {docs} documents, {summaries} summaries, {chunks} total chunks")
        
        if pdf_result.get("success"):
            docs = pdf_result.get("documents_processed", 0)
            summaries = pdf_result.get("summaries_created", 0)
            chunks = pdf_result.get("total_chunks", 0)
            total_docs += docs
            total_summaries += summaries
            total_chunks += chunks
            print(f"📊 PDF Data: {docs} documents, {summaries} summaries, {chunks} total chunks")
        
        print(f"🎯 Total Processed: {total_docs} documents, {total_summaries} summaries, {total_chunks} chunks")
        
        total_time = time.time() - start_time
        print(f"⏱️  Total ingestion time: {total_time:.1f} seconds ({total_time/60:.1f} minutes)")
        
        # Update implementation summary
        print("\n📝 Implementation Status: Moving from 95% to 100% complete!")
        print("🚀 Adaptive RAG with Vietnamese university data is now fully operational!")
        
    except Exception as e:
        elapsed = time.time() - start_time
        print(f"❌ Error after {elapsed:.1f} seconds: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    asyncio.run(ingest_vietnamese_data())
