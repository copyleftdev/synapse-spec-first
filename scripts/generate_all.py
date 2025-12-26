#!/usr/bin/env python3
"""
Generate All Diagrams
Master script to generate all architecture diagrams for the article.
"""

import os
import subprocess
import sys

SCRIPTS_DIR = os.path.dirname(os.path.abspath(__file__))
OUTPUT_DIR = os.path.join(SCRIPTS_DIR, "output")

def main():
    # Ensure output directory exists
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    
    print("=" * 60)
    print("Synapse Architecture Diagram Generator")
    print("=" * 60)
    print()
    
    diagrams = [
        ("diagram_architecture.py", "System Architecture"),
        ("diagram_doc_first.py", "Doc-First Lifecycle"),
        ("diagram_pipeline.py", "Pipeline Stages"),
        ("diagram_testing.py", "Testing Strategy"),
        ("diagram_philosophy.py", "Philosophy & Principles"),
    ]
    
    success = 0
    failed = 0
    
    for script, name in diagrams:
        print(f"Generating: {name}...")
        script_path = os.path.join(SCRIPTS_DIR, script)
        
        try:
            result = subprocess.run(
                [sys.executable, script_path],
                cwd=SCRIPTS_DIR,
                capture_output=True,
                text=True
            )
            
            if result.returncode == 0:
                print(f"  ✓ {name}")
                success += 1
            else:
                print(f"  ✗ {name}: {result.stderr}")
                failed += 1
                
        except Exception as e:
            print(f"  ✗ {name}: {e}")
            failed += 1
    
    print()
    print("=" * 60)
    print(f"Complete: {success} succeeded, {failed} failed")
    print(f"Output directory: {OUTPUT_DIR}")
    print("=" * 60)
    
    # List generated files
    if os.path.exists(OUTPUT_DIR):
        files = os.listdir(OUTPUT_DIR)
        if files:
            print("\nGenerated files:")
            for f in sorted(files):
                size = os.path.getsize(os.path.join(OUTPUT_DIR, f))
                print(f"  • {f} ({size:,} bytes)")

if __name__ == "__main__":
    main()
