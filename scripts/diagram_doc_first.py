#!/usr/bin/env python3
"""
Doc-First Development Lifecycle Diagram
Shows the specification-driven development workflow.
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.programming.language import Go
from diagrams.generic.blank import Blank
from diagrams.onprem.vcs import Git
from diagrams.programming.framework import FastAPI
from diagrams.custom import Custom

graph_attr = {
    "fontsize": "20",
    "bgcolor": "white",
    "pad": "0.5",
    "splines": "ortho",
    "nodesep": "0.8",
    "ranksep": "1.2",
}

with Diagram(
    "Doc-First Development Lifecycle",
    filename="output/doc_first_lifecycle",
    show=False,
    direction="TB",
    graph_attr=graph_attr,
):
    
    with Cluster("1. SPECIFICATIONS (Source of Truth)"):
        with Cluster("API Contracts"):
            openapi = FastAPI("OpenAPI 3.1\nREST Endpoints")
            asyncapi = FastAPI("AsyncAPI 3.0\nEvent Schemas")
    
    with Cluster("2. CODE GENERATION"):
        generator = Go("synctl\nCode Generator")
    
    with Cluster("3. GENERATED CODE"):
        with Cluster("Domain Types"):
            types = Go("types.gen.go\n31 Structs")
        with Cluster("Interfaces"):
            server = Go("server.gen.go\nHTTP Interface")
            client = Go("client.gen.go\nHTTP Client")
            events = Go("events.gen.go\nEvent Handlers")
    
    with Cluster("4. IMPLEMENTATION"):
        handlers = Go("handlers/\nBusiness Logic")
        pipeline = Go("pipeline/\nStage Logic")
    
    with Cluster("5. CONFORMANCE TESTING"):
        contract_tests = Go("Contract Tests\nSpec Validation")
    
    # Flow
    openapi >> Edge(label="parse", color="blue") >> generator
    asyncapi >> Edge(label="parse", color="blue") >> generator
    
    generator >> Edge(color="green") >> types
    generator >> Edge(color="green") >> server
    generator >> Edge(color="green") >> client
    generator >> Edge(color="green") >> events
    
    types >> Edge(style="dashed", color="gray") >> handlers
    server >> Edge(style="dashed", color="gray") >> handlers
    events >> Edge(style="dashed", color="gray") >> pipeline
    
    handlers >> Edge(label="validate", color="purple") >> contract_tests
    pipeline >> Edge(label="validate", color="purple") >> contract_tests
    
    contract_tests >> Edge(label="verify against", color="purple", style="dashed") >> openapi
    contract_tests >> Edge(label="verify against", color="purple", style="dashed") >> asyncapi

print("✓ Generated: output/doc_first_lifecycle.png")
