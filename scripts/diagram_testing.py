#!/usr/bin/env python3
"""
Testing Philosophy Diagram
Shows the testing pyramid and conformance testing approach.
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.programming.language import Go
from diagrams.onprem.container import Docker
from diagrams.onprem.queue import Nats
from diagrams.onprem.database import PostgreSQL
from diagrams.onprem.inmemory import Redis
from diagrams.programming.framework import FastAPI
from diagrams.generic.storage import Storage

graph_attr = {
    "fontsize": "20",
    "bgcolor": "white",
    "pad": "0.5",
    "splines": "spline",
    "nodesep": "0.8",
    "ranksep": "1.0",
}

with Diagram(
    "Testing Strategy & Conformance",
    filename="output/testing_strategy",
    show=False,
    direction="TB",
    graph_attr=graph_attr,
):
    
    with Cluster("Specifications"):
        openapi = FastAPI("OpenAPI 3.1")
        asyncapi = FastAPI("AsyncAPI 3.0")
    
    with Cluster("Contract Testing"):
        with Cluster("HTTP Conformance"):
            http_tests = Go("OpenAPI\nValidator")
            http_checks = Storage("• Response structure\n• Status codes\n• Content types")
        
        with Cluster("Event Conformance"):
            event_tests = Go("AsyncAPI\nValidator")
            event_checks = Storage("• Payload schema\n• Required fields\n• Data types")
    
    with Cluster("Integration Testing"):
        with Cluster("Testcontainers"):
            docker = Docker("Docker")
            nats = Nats("NATS")
            postgres = PostgreSQL("PostgreSQL")
            redis = Redis("Redis")
        
        pipeline_tests = Go("Pipeline\nIntegration Tests")
    
    with Cluster("Test Execution"):
        test_runner = Go("go test ./...")
    
    # Connections
    openapi >> Edge(color="purple", label="validate against") >> http_tests
    asyncapi >> Edge(color="purple", label="validate against") >> event_tests
    
    http_tests >> Edge(color="green") >> test_runner
    event_tests >> Edge(color="green") >> test_runner
    
    docker >> Edge(style="dashed", color="gray") >> nats
    docker >> Edge(style="dashed", color="gray") >> postgres
    docker >> Edge(style="dashed", color="gray") >> redis
    
    nats >> Edge(color="blue") >> pipeline_tests
    postgres >> Edge(color="blue") >> pipeline_tests
    redis >> Edge(color="blue") >> pipeline_tests
    
    pipeline_tests >> Edge(color="green") >> test_runner

print("✓ Generated: output/testing_strategy.png")
