#!/usr/bin/env python3
"""
Event Pipeline Stages Diagram
Shows the Watermill-based event processing pipeline.
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.programming.language import Go
from diagrams.onprem.queue import Nats
from diagrams.onprem.database import PostgreSQL
from diagrams.onprem.inmemory import Redis
from diagrams.generic.storage import Storage

graph_attr = {
    "fontsize": "20",
    "bgcolor": "white",
    "pad": "0.5",
    "splines": "spline",
    "nodesep": "1.0",
    "ranksep": "1.0",
}

with Diagram(
    "Order Processing Pipeline",
    filename="output/pipeline_stages",
    show=False,
    direction="LR",
    graph_attr=graph_attr,
):
    
    ingest = Nats("orders.ingest")
    
    with Cluster("Stage 1: Validate"):
        validate = Go("Validate\nOrder")
        validate_desc = Storage("• Check required fields\n• Verify amounts\n• Validate customer")
    
    validated = Nats("orders.validated")
    
    with Cluster("Stage 2: Enrich"):
        enrich = Go("Enrich\nOrder")
        enrich_desc = Storage("• Customer tier lookup\n• Fraud score\n• Inventory check")
        redis = Redis("Redis\nCache")
    
    enriched = Nats("orders.enriched")
    
    with Cluster("Stage 3: Route"):
        route = Go("Route\nOrder")
        route_desc = Storage("• Apply routing rules\n• Determine destination\n• Set priority")
    
    routed = Nats("orders.routed")
    
    with Cluster("Persistence"):
        postgres = PostgreSQL("PostgreSQL")
    
    dlq = Storage("DLQ\norders.dlq")
    
    # Main flow
    ingest >> Edge(color="blue", label="consume") >> validate
    validate >> Edge(color="green", label="publish") >> validated
    
    validated >> Edge(color="blue", label="consume") >> enrich
    enrich >> Edge(color="green", label="publish") >> enriched
    enrich >> Edge(style="dashed", color="gray") >> redis
    
    enriched >> Edge(color="blue", label="consume") >> route
    route >> Edge(color="green", label="publish") >> routed
    
    routed >> Edge(color="green", label="persist") >> postgres
    
    # Error paths
    validate >> Edge(color="red", style="dashed", label="error") >> dlq
    enrich >> Edge(color="red", style="dashed", label="error") >> dlq
    route >> Edge(color="red", style="dashed", label="error") >> dlq

print("✓ Generated: output/pipeline_stages.png")
