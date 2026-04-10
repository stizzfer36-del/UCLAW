package memory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/testingx"
)

func setup(t *testing.T) config.Config { t.Helper(); testingx.TempHome(t); cfg, err:=config.Load(); if err!=nil { t.Fatal(err) }; if err:=config.EnsureLayout(cfg); err!=nil { t.Fatal(err) }; return cfg }
func TestNodeCreateQueryRoundTrip(t *testing.T) { cfg:=setup(t); node, err:=CreateNode(context.Background(), cfg, Node{Type:"decision",Title:"Phase 2 Decision",Content:"Memory vault round trip",Verified:true}); if err!=nil { t.Fatal(err) }; nodes, err:=Query(context.Background(), cfg, "decision", ""); if err!=nil { t.Fatal(err) }; if len(nodes)!=1 || nodes[0].ID != node.ID { t.Fatalf("expected created node, got %+v", nodes) } }
func TestEdgeTraversalDepth(t *testing.T) { cfg:=setup(t); a, _:=CreateNode(context.Background(), cfg, Node{Type:"decision",Title:"A",MissionID:"mission-1",Content:"a"}); b, _:=CreateNode(context.Background(), cfg, Node{Type:"decision",Title:"B",MissionID:"mission-1",Content:"b"}); if _, err:=AddEdge(context.Background(), cfg, a.ID, b.ID, "caused-by"); err!=nil { t.Fatal(err) }; graph, err:=Graph(context.Background(), cfg, "mission-1", 2); if err!=nil { t.Fatal(err) }; if len(graph["nodes"].([]Node)) != 2 || len(graph["edges"].([]Edge)) != 1 { t.Fatalf("unexpected graph: %+v", graph) } }
func TestFullTextSearchAccuracy(t *testing.T) { cfg:=setup(t); if _, err:=CreateNode(context.Background(), cfg, Node{Type:"note",Title:"Compiler",Content:"capability layer security"}); err!=nil { t.Fatal(err) }; if _, err:=CreateNode(context.Background(), cfg, Node{Type:"note",Title:"Other",Content:"nothing relevant"}); err!=nil { t.Fatal(err) }; nodes, err:=Search(context.Background(), cfg, "capability layer"); if err!=nil { t.Fatal(err) }; if len(nodes)!=1 || nodes[0].Title != "Compiler" { t.Fatalf("unexpected search results: %+v", nodes) } }
func TestObsidianCompatibility(t *testing.T) { cfg:=setup(t); obsidianPath:=filepath.Join(cfg.VaultPath, ".obsidian"); if err:=os.MkdirAll(obsidianPath, 0o755); err!=nil { t.Fatal(err) }; sentinel:=filepath.Join(obsidianPath, "workspace.json"); if err:=os.WriteFile(sentinel, []byte(`{"sentinel":true}`), 0o644); err!=nil { t.Fatal(err) }; node, err:=CreateNode(context.Background(), cfg, Node{Type:"decision",Title:"Obsidian",Content:"frontmatter safe"}); if err!=nil { t.Fatal(err) }; body, err:=os.ReadFile(node.Path); if err!=nil { t.Fatal(err) }; if len(body)==0 || string(body[:3]) != "---" { t.Fatalf("expected frontmatter markdown, got %q", string(body)) }; sentinelBody, err:=os.ReadFile(sentinel); if err!=nil { t.Fatal(err) }; if string(sentinelBody) != `{"sentinel":true}` { t.Fatalf("obsidian config was modified: %s", string(sentinelBody)) } }
func TestMemoryWritesAreAudited(t *testing.T) { cfg:=setup(t); a, err:=CreateNode(context.Background(), cfg, Node{Type:"decision",Title:"A",MissionID:"mission-1",AgentID:"agent-1",Content:"a"}); if err!=nil { t.Fatal(err) }; if _, err:=AddEdge(context.Background(), cfg, a.ID, a.ID, "references"); err!=nil { t.Fatal(err) }; body, err:=os.ReadFile(cfg.AuditPath); if err!=nil { t.Fatal(err) }; if !strings.Contains(string(body), "memory_node_create") || !strings.Contains(string(body), "memory_edge_create") { t.Fatalf("expected memory audit events, got %s", string(body)) } }
