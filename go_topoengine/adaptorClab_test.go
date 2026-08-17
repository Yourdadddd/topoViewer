package topoengine

import (
	"strings"
	"testing"
)

// Minimal but structurally valid containerlab topology-data export.
func validTopoJSON(name string, nodeNames ...string) string {
	var nodes []string
	for i, n := range nodeNames {
		nodes = append(nodes, `{
			"id": "`+n+`",
			"index": "`+string(rune('0'+i))+`",
			"shortname": "`+n+`",
			"longname": "clab-`+name+`-`+n+`",
			"fqdn": "`+n+`.`+name+`.io",
			"kind": "nokia_srlinux",
			"image": "ghcr.io/nokia/srlinux",
			"mgmt-ipv4-address": "10.0.0.`+string(rune('2'+i))+`",
			"mgmt-ipv4-prefix-length": 24,
			"labels": {}
		}`)
	}
	return `{
		"name": "` + name + `",
		"type": "clab",
		"nodes": [` + strings.Join(nodes, ",") + `],
		"links": []
	}`
}

func TestUnmarshalContainerLabTopoV2_Valid(t *testing.T) {
	cyTopo := CytoTopology{}
	out, err := cyTopo.UnmarshalContainerLabTopoV2(
		[]byte(validTopoJSON("citest", "r1", "r2")), "clab", nil)
	if err != nil {
		t.Fatalf("valid topology returned error: %v", err)
	}
	if cyTopo.ClabTopoDataV2.Name != "citest" {
		t.Fatalf("lab name = %q, want citest", cyTopo.ClabTopoDataV2.Name)
	}
	if len(cyTopo.ClabTopoDataV2.Nodes) != 2 {
		t.Fatalf("parsed %d nodes, want 2", len(cyTopo.ClabTopoDataV2.Nodes))
	}
	if !strings.Contains(string(out), "clab-citest-r1") {
		t.Fatalf("cytoscape payload missing node longname; got: %.200s", out)
	}
}

func TestUnmarshalContainerLabTopoV2_MalformedJSON(t *testing.T) {
	// Regression for netpilot-labs/TopoViewer#6: the unmarshal error used to be
	// swallowed, serving a zero-valued topology as if the parse succeeded.
	cyTopo := CytoTopology{}
	if _, err := cyTopo.UnmarshalContainerLabTopoV2([]byte(`{"name": `), "clab", nil); err == nil {
		t.Fatal("malformed JSON must return an error")
	}
}

func TestUnmarshalContainerLabTopoV2_TypeInvalid(t *testing.T) {
	// A type-invalid field partially populates the struct; the error must still
	// surface (TopoViewer#6 round 2).
	bad := strings.Replace(validTopoJSON("citest", "r1"),
		`"mgmt-ipv4-prefix-length": 24`, `"mgmt-ipv4-prefix-length": "bad"`, 1)
	cyTopo := CytoTopology{}
	if _, err := cyTopo.UnmarshalContainerLabTopoV2([]byte(bad), "clab", nil); err == nil {
		t.Fatal("type-invalid JSON must return an error")
	}
}
