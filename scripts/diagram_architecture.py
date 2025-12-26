#!/usr/bin/env python3
"""
Synapse System Architecture Diagram
Generates a visual representation of the complete system architecture.
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.programming.language import Go
from diagrams.onprem.queue import Nats
from diagrams.onprem.database import PostgreSQL
from diagrams.onprem.inmemory import Redis
from diagrams.onprem.client import Users
from diagrams.onprem.container import Docker
from diagrams.generic.storage import Storage
from diagrams.programming.framework import React
from diagrams.custom import Custom

# Graph attributes for aesthetics
graph_attr = {
    "fontsize": "20",
    "bgcolor": "white",
    "pad": "0.5",
    "splines": "spline",
    "nodesep": "1.0",
    "ranksep": "1.5",
}

node_attr = {
    "fontsize": "14",
}

edge_attr = {
    "fontsize": "12",
}

with Diagram(
    "Synapse - Event-Driven Order Processing",
    filename="output/architecture",
    show=False,
    direction="LR",
    graph_attr=graph_attr,
    node_attr=node_attr,
    edge_attr=edge_attr,
):
    # External
    client = Users("API Clients")
    
    with Cluster("Synapse Platform"):
        
        with Cluster("API Layer"):
            api = Go("HTTP API\n(Chi Router)")
        
        with Cluster("Event Bus"):
            nats = Nats("NATS\nJetStream")
        
        with Cluster("Pipeline Stages (Watermill)"):
            validate = Go("Validate")
            enrich = Go("Enrich")
            route = Go("Route")
        
        with Cluster("Data Layer"):
            postgres = PostgreSQL("PostgreSQL\nOrders & Events")
            redis = Redis("Redis\nCache & State")
        
        with Cluster("Dead Letter Queue"):
            dlq = Storage("DLQ\nFailed Events")

    # Connections
    client >> Edge(label="REST", color="darkgreen") >> api
    
    api >> Edge(label="publish", color="blue") >> nats
    
    nats >> Edge(label="subscribe", color="blue") >> validate
    validate >> Edge(label="validated", color="green") >> enrich
    enrich >> Edge(label="enriched", color="green") >> route
    route >> Edge(label="routed", color="green") >> postgres
    
    validate >> Edge(label="lookup", style="dashed", color="gray") >> redis
    enrich >> Edge(label="cache", style="dashed", color="gray") >> redis
    
    validate >> Edge(label="errors", color="red", style="dashed") >> dlq
    enrich >> Edge(label="errors", color="red", style="dashed") >> dlq
    route >> Edge(label="errors", color="red", style="dashed") >> dlq

print("✓ Generated: output/architecture.png")
