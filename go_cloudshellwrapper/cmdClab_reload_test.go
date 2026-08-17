package cloudshellwrapper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	topoengine "github.com/asadarafat/topoViewer/go_topoengine"
)

// topoJSON builds a minimal valid containerlab topology-data export.
func topoJSON(name string, nodeNames ...string) string {
	var nodes []string
	for _, n := range nodeNames {
		nodes = append(nodes, `{
			"id": "`+n+`",
			"shortname": "`+n+`",
			"longname": "clab-`+name+`-`+n+`",
			"fqdn": "`+n+`.`+name+`.io",
			"kind": "nokia_srlinux",
			"image": "ghcr.io/nokia/srlinux",
			"mgmt-ipv4-address": "10.0.0.2",
			"mgmt-ipv4-prefix-length": 24,
			"labels": {}
		}`)
	}
	return `{"name": "` + name + `", "type": "clab", "nodes": [` +
		strings.Join(nodes, ",") + `], "links": []}`
}

// setupReloadTest points the package config at a temp topology-data JSON file,
// resets the shared topology state, and chdirs into a temp dir so the reload's
// dataCytoMarshall.json write lands there. Tests sharing package globals must
// not run in parallel.
func setupReloadTest(t *testing.T, labName string) (topoPath string) {
	t.Helper()
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.MkdirAll(filepath.Join(dir, "html-public", labName), 0o755); err != nil {
		t.Fatal(err)
	}

	topoPath = filepath.Join(dir, "topology-data.json")
	if err := confClab["topology-file-json"].SetValue(topoPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = confClab["topology-file-json"].SetValue(".") })

	topoStateMu.Lock()
	cyTopo = topoengine.CytoTopology{}
	cyTopoJsonBytes = nil
	topoStateMu.Unlock()
	return topoPath
}

func servedPayload(t *testing.T) string {
	t.Helper()
	topoStateMu.RLock()
	defer topoStateMu.RUnlock()
	return string(cyTopoJsonBytes)
}

func writeTopo(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReloadTopoFile_SwapsServedState(t *testing.T) {
	// THE regression behind netpilot-labs/containerlab-mcp#122: /reload-topo
	// rebuilt the topology into discarded locals, so the served payload stayed
	// frozen at process start and nodes added by a redeploy had a dead
	// Terminal button.
	path := setupReloadTest(t, "citest")

	writeTopo(t, path, topoJSON("citest", "r1", "r2"))
	if err := reloadTopoFile(); err != nil {
		t.Fatalf("first reload failed: %v", err)
	}
	if p := servedPayload(t); !strings.Contains(p, "clab-citest-r1") {
		t.Fatalf("served payload missing r1 after first reload: %.200s", p)
	}

	writeTopo(t, path, topoJSON("citest", "r1", "r2", "r3"))
	if err := reloadTopoFile(); err != nil {
		t.Fatalf("second reload failed: %v", err)
	}
	if p := servedPayload(t); !strings.Contains(p, "clab-citest-r3") {
		t.Fatal("served payload does not contain the node added by the second reload")
	}

	topoStateMu.RLock()
	nodeCount := len(cyTopo.ClabTopoDataV2.Nodes)
	topoStateMu.RUnlock()
	if nodeCount != 3 {
		t.Fatalf("shared cyTopo has %d nodes after reload, want 3", nodeCount)
	}

	if _, err := os.Stat(filepath.Join("html-public", "citest", "dataCytoMarshall.json")); err != nil {
		t.Fatalf("reload did not write dataCytoMarshall.json: %v", err)
	}
}

func TestReloadTopoFile_MalformedLeavesPreviousState(t *testing.T) {
	path := setupReloadTest(t, "citest")

	writeTopo(t, path, topoJSON("citest", "r1"))
	if err := reloadTopoFile(); err != nil {
		t.Fatalf("good reload failed: %v", err)
	}
	before := servedPayload(t)

	writeTopo(t, path, `{"name": `)
	if err := reloadTopoFile(); err == nil {
		t.Fatal("malformed topology must fail the reload")
	}
	if servedPayload(t) != before {
		t.Fatal("failed reload must leave the previous served state untouched")
	}
}

func TestReloadTopoFile_EmptyTopologyRejected(t *testing.T) {
	path := setupReloadTest(t, "citest")

	writeTopo(t, path, topoJSON("citest", "r1"))
	if err := reloadTopoFile(); err != nil {
		t.Fatalf("good reload failed: %v", err)
	}
	before := servedPayload(t)

	// {} unmarshals without error but is not a real topology.
	writeTopo(t, path, `{}`)
	if err := reloadTopoFile(); err == nil {
		t.Fatal("empty topology must fail the reload")
	}
	if servedPayload(t) != before {
		t.Fatal("failed reload must leave the previous served state untouched")
	}
}

func TestReloadTopoFile_LabRenameRejected(t *testing.T) {
	path := setupReloadTest(t, "citest")

	writeTopo(t, path, topoJSON("citest", "r1"))
	if err := reloadTopoFile(); err != nil {
		t.Fatalf("good reload failed: %v", err)
	}

	// Lab-scoped assets are built once for the startup name; serving a renamed
	// lab from the same process would point handlers at missing paths.
	writeTopo(t, path, topoJSON("renamed", "r1"))
	err := reloadTopoFile()
	if err == nil {
		t.Fatal("lab rename must fail the reload")
	}
	if !strings.Contains(err.Error(), "restart topoviewer") {
		t.Fatalf("rename rejection should direct to a restart; got: %v", err)
	}

	topoStateMu.RLock()
	name := cyTopo.ClabTopoDataV2.Name
	topoStateMu.RUnlock()
	if name != "citest" {
		t.Fatalf("served lab name changed to %q on a rejected reload", name)
	}
}
