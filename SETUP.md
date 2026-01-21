# TopoViewer Setup Guide

## Installation

```bash
# Clone repository
git clone https://github.com/asadarafat/topoViewer.git /opt/topoviewer-dev

# Deploy to /opt/topoviewer
sudo mkdir -p /opt/topoviewer
sudo cp -r /opt/topoviewer-dev/dist/* /opt/topoviewer/
sudo cp -r /opt/topoviewer-dev/html-template /opt/topoviewer/
sudo mkdir -p /opt/topoviewer/logs

# Optional: create symlink
sudo ln -sf /opt/topoviewer/topoviewer /usr/local/bin/topoviewer
```

**Note:** `html-template/` is not included in `dist/` - must copy separately.

## Running TopoViewer

```bash
cd /opt/topoviewer && sudo ./topoviewer clab \
  --topology-file-yaml /path/to/topology.clab.yml \
  --server-port 50080 \
  --deployment-type colocated \
  --clab-user $(whoami) \
  --allowed-hostnames localhost,127.0.0.1,YOUR_IP \
  --clab-server-address YOUR_IP
```

Access: `http://YOUR_IP:50080`

### Example (clab1 VM)

```bash
# Stop existing instance
sudo pkill -f "topoviewer clab"

# Start TopoViewer
cd /opt/topoviewer && sudo ./topoviewer clab \
  --topology-file-yaml /opt/containerlab/simple-demo/simple-demo.clab.yml \
  --server-port 8080 \
  --deployment-type colocated \
  --clab-user linzhu \
  --allowed-hostnames localhost,127.0.0.1,34.42.33.0,clab1-topo.netpilot.io \
  --clab-server-address 34.42.33.0 &
```

Access: `https://clab1-topo.netpilot.io`

## Key Options

| Flag | Description |
|------|-------------|
| `--topology-file-yaml` | Path to Container Lab YAML topology |
| `--server-port` | HTTP server port (default: 8080) |
| `--deployment-type` | `colocated` (same host as clab) or `container` |
| `--allowed-hostnames` | Hostnames allowed for WebSocket (comma-separated) |
| `--clab-server-address` | IP for SSH terminal connections |
| `--clab-user` | SSH user for container access |

## Important Notes

- Run binary from `/opt/topoviewer/` directory (uses relative paths for templates)
- `--allowed-hostnames` must include any reverse proxy/tunnel hostnames for terminal to work
- See [BUGFIXES.md](BUGFIXES.md) for known issues and fixes
