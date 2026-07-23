package entitymatcher

// Reproduces: the graph entity matcher does not honour VRF when matching IP
// addresses. Two allocations of the same address in different VRFs collapse
// onto a single graph node.
//
// This test is EXPECTED TO FAIL on current code — it asserts the correct
// behaviour (same address + different VRF => distinct nodes) and fails because
// the IPAddress match rule is address-only.
//
//   go test ./entitymatcher/ -run TestIPMatch_VRFIgnored_Bug -v
//
// The matcher runs against a stateful in-memory NodeFinder whose stored node
// Data mirrors what graph.Service persists (protojson of the entity wrapper).
// The IPAddress rule below matches on address only, exactly like the shipped
// config (examples/entity_matching_config.yaml).

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/graph"
	"github.com/netboxlabs/diode/diode-server/matching"
)

type vrfBugFinder struct{ nodes []graph.Node }

func (f *vrfBugFinder) FindNodesByFieldMatch(_ context.Context, arg graph.FindNodesByFieldMatchParams) ([]graph.Node, error) {
	var out []graph.Node
	for _, n := range f.nodes {
		if n.NodeType != arg.NodeType {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(n.Data, &m); err != nil {
			continue
		}
		if len(arg.NestedPath) > 0 {
			if v, ok := vrfBugGetPath(m, arg.NestedPath); ok && vrfBugStr(v) == arg.NestedValue {
				out = append(out, n)
			}
		} else if arg.JSONField != "" {
			if v, ok := m[arg.JSONField]; ok && vrfBugStr(v) == arg.FieldValue {
				out = append(out, n)
			}
		}
	}
	return out, nil
}

func (f *vrfBugFinder) GetNodesByType(_ context.Context, arg graph.GetNodesByTypeParams) ([]graph.Node, error) {
	var out []graph.Node
	for _, n := range f.nodes {
		if n.NodeType == arg.NodeType {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *vrfBugFinder) FindNodeByMetadata(_ context.Context, _ graph.FindNodeByMetadataParams) (graph.Node, error) {
	return graph.Node{}, nil
}

func vrfBugGetPath(data map[string]any, parts []string) (any, bool) {
	var cur any = data
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		if cur, ok = m[p]; !ok {
			return nil, false
		}
	}
	return cur, true
}

func vrfBugStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), "\"")
}

func vrfBugIP(addr, vrf string) *diodepb.Entity {
	return &diodepb.Entity{Entity: &diodepb.Entity_IpAddress{IpAddress: &diodepb.IPAddress{
		Address: addr,
		Vrf:     &diodepb.VRF{Name: vrf},
	}}}
}

// TestIPMatch_VRFIgnored_Bug asserts the desired VRF-aware behaviour and fails
// against current code, demonstrating the bug.
func TestIPMatch_VRFIgnored_Bug(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Address-only IPAddress rule (same identity axis as the shipped config).
	requireAll := true
	cfg := &matching.EntityMatchingConfig{
		Rules: map[string]*matching.EntityMatchingRule{
			"IPAddress": {
				EntityType: "IPAddress",
				PrimaryRules: []matching.FieldMatchRule{
					{FieldPath: "ipAddress.address", MatchType: matching.MatchExact, Weight: 1.0, Required: true, Confidence: matching.ConfidenceHigh},
				},
				MinConfidence:     matching.ConfidenceMedium,
				RequireAllPrimary: &requireAll,
			},
		},
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
	}

	// Existing graph node: 10.0.0.1/24 in vrf-A (stored as protojson, like graph.Service).
	stored, _ := protojson.Marshal(vrfBugIP("10.0.0.1/24", "vrf-A"))
	finder := &vrfBugFinder{nodes: []graph.Node{
		{ID: 1, ExternalID: "ip-1", NodeType: "IPAddress", Data: stored, DuplicateCount: 1},
	}}
	m := NewMatcher(finder, cfg, logger)

	// Incoming: same address, DIFFERENT VRF. It must NOT match the vrf-A node.
	best, err := m.FindBestMatch(context.Background(), vrfBugIP("10.0.0.1/24", "vrf-B"))
	if err != nil {
		t.Fatalf("FindBestMatch: %v", err)
	}
	if best != nil {
		t.Fatalf("BUG: 10.0.0.1/24 in vrf-B collapsed onto the vrf-A node %q (VRF not honoured); "+
			"expected a distinct node (no match)", *best.ExternalID)
	}
}
