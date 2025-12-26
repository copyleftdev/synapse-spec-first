#!/usr/bin/env python3
"""
Doc-First Philosophy Diagram
Visual representation of the core principles.
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.generic.blank import Blank
from diagrams.programming.language import Go
from diagrams.programming.framework import FastAPI
from diagrams.onprem.vcs import Git
from diagrams.generic.storage import Storage

graph_attr = {
    "fontsize": "22",
    "bgcolor": "white",
    "pad": "0.8",
    "splines": "spline",
    "nodesep": "1.2",
    "ranksep": "1.5",
}

with Diagram(
    "Doc-First Philosophy",
    filename="output/philosophy",
    show=False,
    direction="TB",
    graph_attr=graph_attr,
):
    
    with Cluster("Single Source of Truth"):
        specs = FastAPI("API Specifications\nOpenAPI + AsyncAPI")
    
    with Cluster("Automated Generation"):
        with Cluster("From Specs"):
            gen_types = Go("Domain Types")
            gen_server = Go("Server Interface")
            gen_client = Go("Client SDK")
            gen_events = Go("Event Handlers")
    
    with Cluster("Core Principles"):
        with Cluster("1. Contracts First"):
            principle1 = Storage("Define behavior\nbefore implementation")
        
        with Cluster("2. Generated > Handwritten"):
            principle2 = Storage("Reduce drift\nbetween spec and code")
        
        with Cluster("3. Conformance Testing"):
            principle3 = Storage("Verify implementation\nmatches specification")
        
        with Cluster("4. Continuous Validation"):
            principle4 = Storage("Tests run on\nevery change")
    
    with Cluster("Benefits"):
        benefit1 = Storage("✓ API consistency")
        benefit2 = Storage("✓ Type safety")
        benefit3 = Storage("✓ Auto documentation")
        benefit4 = Storage("✓ Client generation")
    
    # Connections
    specs >> Edge(color="blue", label="generates") >> gen_types
    specs >> Edge(color="blue") >> gen_server
    specs >> Edge(color="blue") >> gen_client
    specs >> Edge(color="blue") >> gen_events
    
    principle1 >> Edge(style="dashed", color="green") >> benefit1
    principle2 >> Edge(style="dashed", color="green") >> benefit2
    principle3 >> Edge(style="dashed", color="green") >> benefit3
    principle4 >> Edge(style="dashed", color="green") >> benefit4

print("✓ Generated: output/philosophy.png")
