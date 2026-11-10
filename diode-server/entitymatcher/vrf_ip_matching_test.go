package entitymatcher

// POC + PROOF: does the diode graph matcher honour VRF when matching IP addresses?
//
// Full write-up: .session/diode-graph-vrf-proof/PROOF.md
//
// These tests exercise the real Matcher.FindBestMatch against a stateful
// in-memory NodeFinder whose stored node Data mirrors what graph.Service
// actually persists (protojson of the entity wrapper, camelCase keys).
//
//   - TestIPMatch_AddressOnly_IgnoresVRF  -> proof: an address-only rule
//     collapses same-address / different-VRF IPs to one node.
//   - TestIPMatch_ShippedSnakeCaseRuleIsInert -> drift note: the *shipped*
//     `ip_address.address` (snake_case) path never fires, because the
//     matcher serialises the entity with Go field names (`IpAddress`).
//   - TestIPMatch_VRFAware_DistinctNodes  -> POC: adding `ipAddress.vrf.name`
//     makes same-address / different-VRF IPs distinct.

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

// memFinder is a stateful in-memory NodeFinder that evaluates nested-path
// field matches against stored node Data, standing in for the Postgres repo.
type memFinder struct{ nodes []graph.Node }

func (f *memFinder) FindNodesByFieldMatch(_ context.Context, arg graph.FindNodesByFieldMatchParams) ([]graph.Node, error) {
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
			if v, ok := getPath(m, arg.NestedPath); ok && matchStr(v) == arg.NestedValue {
				out = append(out, n)
			}
		} else if arg.JSONField != "" {
			if v, ok := m[arg.JSONField]; ok && matchStr(v) == arg.FieldValue {
				out = append(out, n)
			}
		}
	}
	return out, nil
}

func (f *memFinder) GetNodesByType(_ context.Context, arg graph.GetNodesByTypeParams) ([]graph.Node, error) {
	var out []graph.Node
	for _, n := range f.nodes {
		if n.NodeType == arg.NodeType {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *memFinder) FindNodeByMetadata(_ context.Context, _ graph.FindNodeByMetadataParams) (graph.Node, error) {
	return graph.Node{}, nil
}

func getPath(data map[string]any, parts []string) (any, bool) {
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

func matchStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), "\"")
}

func ipEntity(addr, vrf string) *diodepb.Entity {
	return &diodepb.Entity{Entity: &diodepb.Entity_IpAddress{IpAddress: &diodepb.IPAddress{
		Address: addr,
		Vrf:     &diodepb.VRF{Name: vrf},
	}}}
}

// storedIPNode simulates what graph.Service stores for a node: protojson of the
// entity wrapper (camelCase keys, e.g. {"ipAddress":{"address":...,"vrf":...}}).
func storedIPNode(id int64, ext, addr, vrf string) graph.Node {
	b, _ := protojson.Marshal(ipEntity(addr, vrf))
	return graph.Node{ID: id, ExternalID: ext, NodeType: "IPAddress", Data: b, DuplicateCount: 1}
}

func ipRuleConfig(requireAll bool, paths ...string) *matching.EntityMatchingConfig {
	prs := make([]matching.FieldMatchRule, 0, len(paths))
	for _, p := range paths {
		prs = append(prs, matching.FieldMatchRule{
			FieldPath:  p,
			MatchType:  matching.MatchExact,
			Weight:     1.0,
			Required:   true,
			Confidence: matching.ConfidenceHigh,
		})
	}
	return &matching.EntityMatchingConfig{
		Rules: map[string]*matching.EntityMatchingRule{
			"IPAddress": {
				EntityType:        "IPAddress",
				PrimaryRules:      prs,
				MinConfidence:     matching.ConfidenceMedium,
				RequireAllPrimary: &requireAll,
			},
		},
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
	}
}

func newIPMatcher(cfg *matching.EntityMatchingConfig) *Matcher {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	finder := &memFinder{nodes: []graph.Node{storedIPNode(1, "ip-1", "10.0.0.1/24", "vrf-A")}}
	return NewMatcher(finder, cfg, logger)
}

func matchedExternalID(t *testing.T, m *Matcher, addr, vrf string) (string, bool) {
	t.Helper()
	best, err := m.FindBestMatch(context.Background(), ipEntity(addr, vrf))
	if err != nil {
		t.Fatalf("FindBestMatch(%s,%s): %v", addr, vrf, err)
	}
	if best == nil || best.ExternalID == nil {
		return "", false
	}
	return *best.ExternalID, true
}

// PROOF: with a working address-only rule, an incoming IP with the same
// address matches the stored node regardless of VRF -> same-address /
// different-VRF IPs collapse to one graph node (VRF not honoured).
func TestIPMatch_AddressOnly_IgnoresVRF(t *testing.T) {
	m := newIPMatcher(ipRuleConfig(true, "ipAddress.address"))

	sameVRF, ok := matchedExternalID(t, m, "10.0.0.1/24", "vrf-A")
	if !ok {
		t.Fatal("expected same-address/same-VRF IP to match the stored node")
	}
	diffVRF, ok := matchedExternalID(t, m, "10.0.0.1/24", "vrf-B")
	if !ok {
		t.Fatal("expected same-address/DIFFERENT-VRF IP to match the stored node (proves VRF ignored)")
	}
	if sameVRF != diffVRF || diffVRF != "ip-1" {
		t.Fatalf("both should collapse onto ip-1; got sameVRF=%q diffVRF=%q", sameVRF, diffVRF)
	}
	if _, ok := matchedExternalID(t, m, "10.0.0.2/24", "vrf-A"); ok {
		t.Fatal("a different address must not match")
	}
}

// DRIFT NOTE: the shipped snake_case path `ip_address.address` never fires,
// because the matcher serialises the entity with Go field names (top-level
// key `IpAddress`), and EqualFold cannot bridge `_` vs `A`. Today IP identity
// therefore falls through to the content-hash fallback (which happens to
// include VRF), not to this rule.
func TestIPMatch_ShippedSnakeCaseRuleIsInert(t *testing.T) {
	m := newIPMatcher(ipRuleConfig(true, "ip_address.address"))
	if _, ok := matchedExternalID(t, m, "10.0.0.1/24", "vrf-A"); ok {
		t.Fatal("shipped snake_case path unexpectedly matched; drift finding no longer holds")
	}
}

// POC: adding `ipAddress.vrf.name` as a required primary rule makes IP identity
// VRF-aware. Same address + same VRF still collapses; same address + different
// VRF becomes a DISTINCT node.
func TestIPMatch_VRFAware_DistinctNodes(t *testing.T) {
	m := newIPMatcher(ipRuleConfig(true, "ipAddress.address", "ipAddress.vrf.name"))

	if id, ok := matchedExternalID(t, m, "10.0.0.1/24", "vrf-A"); !ok || id != "ip-1" {
		t.Fatalf("same address + same VRF should still collapse onto ip-1; got id=%q ok=%v", id, ok)
	}
	if id, ok := matchedExternalID(t, m, "10.0.0.1/24", "vrf-B"); ok {
		t.Fatalf("same address + DIFFERENT VRF must be a distinct node; unexpectedly matched %q", id)
	}
}
