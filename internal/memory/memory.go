package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stizzfer36-del/uclaw/internal/audit"
	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/ids"
	"github.com/stizzfer36-del/uclaw/internal/sqlitepy"
)

type Node struct { ID string `json:"id"`; Type string `json:"type"`; Title string `json:"title"`; MissionID string `json:"mission_id,omitempty"`; AgentID string `json:"agent_id,omitempty"`; Path string `json:"path"`; Content string `json:"content"`; Verified bool `json:"verified"`; CreatedAt string `json:"created_at"` }
type Edge struct { ID string `json:"id"`; FromID string `json:"from_id"`; ToID string `json:"to_id"`; EdgeType string `json:"edge_type"`; CreatedAt string `json:"created_at"` }
func Ensure(ctx context.Context, cfg config.Config) error { return sqlitepy.Exec(ctx, cfg.GraphDBPath, `CREATE TABLE IF NOT EXISTS nodes (id TEXT PRIMARY KEY,type TEXT NOT NULL,title TEXT NOT NULL,mission_id TEXT,agent_id TEXT,path TEXT NOT NULL,content TEXT NOT NULL,verified INTEGER NOT NULL DEFAULT 0,created_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS edges (id TEXT PRIMARY KEY,from_id TEXT NOT NULL,to_id TEXT NOT NULL,edge_type TEXT NOT NULL,created_at TEXT NOT NULL);`) }
func CreateNode(ctx context.Context, cfg config.Config, node Node) (Node, error) { if err:=Ensure(ctx,cfg); err!=nil { return Node{}, err }; if node.ID=="" { node.ID=ids.New("node") }; node.CreatedAt=time.Now().UTC().Format(time.RFC3339); if node.Path=="" { node.Path=filepath.Join(folderForType(cfg.VaultPath,node.Type), safeFileName(node.Title,node.ID)) }; if err:=os.MkdirAll(filepath.Dir(node.Path),0o755); err!=nil { return Node{}, err }; if err:=os.WriteFile(node.Path, []byte(markdown(node)), 0o644); err!=nil { return Node{}, err }; if err:=sqlitepy.ExecParams(ctx,cfg.GraphDBPath,`INSERT INTO nodes(id, type, title, mission_id, agent_id, path, content, verified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, node.ID,node.Type,node.Title,node.MissionID,node.AgentID,node.Path,node.Content,boolInt(node.Verified),node.CreatedAt); err!=nil { return Node{}, err }; _ = audit.Write(ctx,cfg.AuditPath,audit.Event{AgentID:defaultAgent(node.AgentID,"memory"),Action:"memory_node_create",Target:node.Path,MissionID:node.MissionID,Outcome:"success",ApprovalRequired:false}); return node,nil }
func AddEdge(ctx context.Context, cfg config.Config, fromID,toID,edgeType string) (Edge, error) { if err:=Ensure(ctx,cfg); err!=nil { return Edge{}, err }; edge:=Edge{ID:ids.New("edge"),FromID:fromID,ToID:toID,EdgeType:edgeType,CreatedAt:time.Now().UTC().Format(time.RFC3339)}; if err:=sqlitepy.ExecParams(ctx,cfg.GraphDBPath,`INSERT INTO edges(id, from_id, to_id, edge_type, created_at) VALUES (?, ?, ?, ?, ?)`,edge.ID,edge.FromID,edge.ToID,edge.EdgeType,edge.CreatedAt); err!=nil { return Edge{}, err }; _ = audit.Write(ctx,cfg.AuditPath,audit.Event{AgentID:"memory",Action:"memory_edge_create",Target:fromID+"->"+toID,Outcome:"success",ApprovalRequired:false}); return edge,nil }
func Query(ctx context.Context, cfg config.Config, nodeType, edgeType string) ([]Node, error) { if err:=Ensure(ctx,cfg); err!=nil { return nil, err }; sql:=`SELECT DISTINCT n.id, n.type, n.title, n.mission_id, n.agent_id, n.path, n.content, n.verified, n.created_at FROM nodes n`; params:=[]interface{}{}; if edgeType!="" { sql += ` JOIN edges e ON e.from_id = n.id OR e.to_id = n.id WHERE e.edge_type = ?`; params=append(params, edgeType); if nodeType!="" { sql += ` AND n.type = ?`; params=append(params,nodeType) } } else if nodeType!="" { sql += ` WHERE n.type = ?`; params=append(params,nodeType) }; sql += ` ORDER BY n.created_at ASC`; rows, err:=sqlitepy.Query(ctx,cfg.GraphDBPath,sql,params...); if err!=nil { return nil, err }; return mapNodes(rows), nil }
func Graph(ctx context.Context, cfg config.Config, missionID string, depth int) (map[string]interface{}, error) { if err:=Ensure(ctx,cfg); err!=nil { return nil, err }; if depth<1 { depth=1 }; nodes, err:=sqlitepy.Query(ctx,cfg.GraphDBPath,`SELECT id, type, title, mission_id, agent_id, path, content, verified, created_at FROM nodes WHERE mission_id = ? ORDER BY created_at ASC`, missionID); if err!=nil { return nil, err }; edges, err:=sqlitepy.Query(ctx,cfg.GraphDBPath,`SELECT id, from_id, to_id, edge_type, created_at FROM edges WHERE from_id IN (SELECT id FROM nodes WHERE mission_id = ?) OR to_id IN (SELECT id FROM nodes WHERE mission_id = ?) ORDER BY created_at ASC`, missionID, missionID); if err!=nil { return nil, err }; return map[string]interface{}{"depth":depth,"nodes":mapNodes(nodes),"edges":mapEdges(edges)}, nil }
func Search(ctx context.Context, cfg config.Config, query string) ([]Node, error) { if err:=Ensure(ctx,cfg); err!=nil { return nil, err }; value:="%"+strings.ToLower(query)+"%"; rows, err:=sqlitepy.Query(ctx,cfg.GraphDBPath,`SELECT id, type, title, mission_id, agent_id, path, content, verified, created_at FROM nodes WHERE LOWER(title) LIKE ? OR LOWER(content) LIKE ? ORDER BY created_at ASC`, value, value); if err!=nil { return nil, err }; return mapNodes(rows), nil }
func Unverified(ctx context.Context, cfg config.Config) ([]Node, error) { if err:=Ensure(ctx,cfg); err!=nil { return nil, err }; rows, err:=sqlitepy.Query(ctx,cfg.GraphDBPath,`SELECT id, type, title, mission_id, agent_id, path, content, verified, created_at FROM nodes WHERE verified = 0 ORDER BY created_at ASC`); if err!=nil { return nil, err }; return mapNodes(rows), nil }
func WriteAgentLog(ctx context.Context, cfg config.Config, missionID, agentID, content string) (Node, error) { return CreateNode(ctx,cfg,Node{Type:"log",Title:fmt.Sprintf("Log %s", missionID),MissionID:missionID,AgentID:agentID,Content:content,Verified:false}) }
func folderForType(vaultPath,nodeType string) string { switch nodeType { case "decision": return filepath.Join(vaultPath,"decisions"); case "prompt": return filepath.Join(vaultPath,"prompts"); case "source": return filepath.Join(vaultPath,"sources"); case "log": return filepath.Join(vaultPath,"logs"); case "note": return filepath.Join(vaultPath,"notes"); case "research": return filepath.Join(vaultPath,"research"); default: return filepath.Join(vaultPath,"notes") } }
func safeFileName(title,id string) string { base:=strings.ToLower(strings.ReplaceAll(title," ","-")); base = strings.Map(func(r rune) rune { switch { case r>='a'&&r<='z': return r; case r>='0'&&r<='9': return r; case r=='-': return r; default: return -1 } }, base); if base=="" { base=id }; return base+".md" }
func markdown(node Node) string { return fmt.Sprintf(`---
id: %s
type: %s
title: %s
mission_id: %s
agent_id: %s
verified: %t
created_at: %s
---

%s
`, node.ID,node.Type,node.Title,node.MissionID,node.AgentID,node.Verified,node.CreatedAt,node.Content) }
func mapNodes(rows []map[string]interface{}) []Node { nodes:=make([]Node,0,len(rows)); for _, row := range rows { nodes=append(nodes, Node{ID:fmt.Sprintf("%v", row["id"]),Type:fmt.Sprintf("%v", row["type"]),Title:fmt.Sprintf("%v", row["title"]),MissionID:fmt.Sprintf("%v", row["mission_id"]),AgentID:fmt.Sprintf("%v", row["agent_id"]),Path:fmt.Sprintf("%v", row["path"]),Content:fmt.Sprintf("%v", row["content"]),Verified:fmt.Sprintf("%v", row["verified"])=="1",CreatedAt:fmt.Sprintf("%v", row["created_at"])}) }; return nodes }
func defaultAgent(value,fallback string) string { if strings.TrimSpace(value)=="" { return fallback }; return value }
func mapEdges(rows []map[string]interface{}) []Edge { edges:=make([]Edge,0,len(rows)); for _, row := range rows { edges=append(edges, Edge{ID:fmt.Sprintf("%v", row["id"]),FromID:fmt.Sprintf("%v", row["from_id"]),ToID:fmt.Sprintf("%v", row["to_id"]),EdgeType:fmt.Sprintf("%v", row["edge_type"]),CreatedAt:fmt.Sprintf("%v", row["created_at"])}) }; return edges }
func boolInt(v bool) int { if v { return 1 }; return 0 }
