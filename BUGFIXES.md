# TopoViewer Bug Fixes

## BUG-001: Orphaned Parent References Break Cytoscape Rendering

**Status:** Documented (not yet fixed in source)

**Problem:** When removing group nodes (`topoviewer`, `TopoViewer:1`), child nodes retain orphaned `parent` references, causing Cytoscape.js rendering to fail.

**Location:** `html-static/js/dev.js` around line 697

**Original code:**
```javascript
// remove node topoviewer
topoViewerNode = cy.filter('node[name = "topoviewer"]');
topoViewerNode.remove();

// remove node TopoViewerParentNode
topoViewerParentNode = cy.filter('node[name = "TopoViewer:1"]');
topoViewerParentNode.remove();
```

**Fixed code:**
```javascript
// remove node topoviewer and clean up orphaned parent references
topoViewerNode = cy.filter('node[name = "topoviewer"]');
if (topoViewerNode.length > 0) {
    const topoViewerNodeId = topoViewerNode.id();
    cy.nodes().forEach(function(node) {
        if (node.data('parent') === topoViewerNodeId) {
            node.move({ parent: null });
        }
    });
    topoViewerNode.remove();
}

// remove node TopoViewerParentNode and clean up orphaned parent references
topoViewerParentNode = cy.filter('node[name = "TopoViewer:1"]');
if (topoViewerParentNode.length > 0) {
    const parentNodeId = topoViewerParentNode.id();
    cy.nodes().forEach(function(node) {
        if (node.data('parent') === parentNodeId) {
            node.move({ parent: null });
        }
    });
    topoViewerParentNode.remove();
}
```

---

## BUG-002: WebSocket Terminal Fails Through Cloudflare Tunnel

**Status:** Fixed (configuration)

**Problem:** Terminal shows "connection has been terminated from the server-side" when accessing via Cloudflare tunnel URL.

**Cause:** WebSocket origin check in `go_xtermjs/utils.go:16-28` rejects connections from hostnames not in `--allowed-hostnames`.

**Fix:** Add tunnel hostname to startup command:
```bash
--allowed-hostnames localhost,127.0.0.1,YOUR_IP,your-tunnel-hostname.example.com
```

---

## BUG-003: Welcome Panel Shows on Every Page Load

**Status:** Fixed

**Problem:** Introduction/welcome panel displays every time the page loads.

**Location:** `html-template/clab/dev.html.tmpl` and `index.html.tmpl` line ~1012

**Fix:** Changed `style="display: block;"` to `style="display: none;"` on `#panel-introduction` div.
