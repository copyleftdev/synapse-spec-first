# Synapse Diagrams

Architecture diagrams for the Synapse doc-first event processing system.

## Prerequisites

```bash
# Install Graphviz (required by diagrams library)
# Ubuntu/Debian
sudo apt-get install graphviz

# macOS
brew install graphviz

# Install Python dependencies
pip install -r requirements.txt
```

## Generate All Diagrams

```bash
cd scripts
python generate_all.py
```

## Individual Diagrams

| Script | Output | Description |
|--------|--------|-------------|
| `diagram_architecture.py` | `output/architecture.png` | Complete system architecture |
| `diagram_doc_first.py` | `output/doc_first_lifecycle.png` | Spec-first development workflow |
| `diagram_pipeline.py` | `output/pipeline_stages.png` | Event processing pipeline stages |
| `diagram_testing.py` | `output/testing_strategy.png` | Testing pyramid and conformance |
| `diagram_philosophy.py` | `output/philosophy.png` | Doc-first principles and benefits |

## Output

All diagrams are generated as PNG files in the `output/` directory.

## Customization

Each diagram script can be customized:
- `graph_attr`: Overall diagram styling
- `node_attr`: Node styling
- `edge_attr`: Connection styling
- `direction`: Layout direction (TB, LR, BT, RL)

## Icons

The diagrams library uses icons from:
- AWS, GCP, Azure (cloud providers)
- Kubernetes, Docker (containers)
- Programming languages (Go, Python, etc.)
- Databases (PostgreSQL, Redis, etc.)
- Generic icons (storage, network, etc.)

See: https://diagrams.mingrammer.com/docs/nodes/onprem
