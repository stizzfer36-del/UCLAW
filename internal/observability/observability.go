package observability

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/stizzfer36-del/UCLAW/internal/artifacts"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/missions"
	"github.com/stizzfer36-del/UCLAW/internal/sqlitepy"
)

type State struct { Missions []missions.Mission `json:"missions"`; Artifacts []artifacts.Artifact `json:"artifacts"`; Agents []map[string]interface{} `json:"agents"`; Approvals []map[string]interface{} `json:"approvals"`; Review []map[string]interface{} `json:"review_queue"`; Health map[string]interface{} `json:"health"`; Budget map[string]interface{} `json:"budget"`; Errors []map[string]interface{} `json:"errors"`; Timeline []missions.TimelineEvent `json:"timeline"`; Workflow []map[string]interface{} `json:"workflow_queue"` }
func Status(ctx context.Context, cfg config.Config) (State, error) {
	missionList, err := missions.List(ctx, cfg); if err != nil { return State{}, err }
	artList, err := artifacts.List(ctx, cfg, ""); if err != nil { return State{}, err }
	agentRows, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT ap.id, m.name, ap.role, ap.status, ap.provider FROM agent_profiles ap JOIN members m ON m.id = ap.member_id ORDER BY m.created_at ASC`)
	approvals, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT * FROM approval_requests WHERE status = 'pending' ORDER BY requested_at ASC`)
	review, _ := missions.ReviewQueue(ctx, cfg); workflow := []map[string]interface{}{}; for _, item := range review { workflow = append(workflow, item) }; for _, item := range approvals { workflow = append(workflow, item) }
	health, _ := Health(ctx, cfg); budget, _ := Budget(ctx, cfg); errorsList, _ := Errors(cfg); var timeline []missions.TimelineEvent; if len(missionList) > 0 { timeline, _ = missions.Timeline(ctx, cfg, missionList[len(missionList)-1].ID) }
	return State{Missions:missionList,Artifacts:artList,Agents:agentRows,Approvals:approvals,Review:review,Health:health,Budget:budget,Errors:errorsList,Timeline:timeline,Workflow:workflow}, nil }
func Health(ctx context.Context, cfg config.Config) (map[string]interface{}, error) { agentRows, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT status FROM agent_profiles`); worldInfo := map[string]interface{}{"db_size":fileSize(cfg.DBPath),"vault_size":dirSize(cfg.VaultPath),"checkpoint_count":len(mustCheckpoints(ctx, cfg))}; return map[string]interface{}{"agent_count":len(agentRows),"provider_health":"local","world":worldInfo}, nil }
func Budget(ctx context.Context, cfg config.Config) (map[string]interface{}, error) { rows, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT provider, SUM(tokens) AS tokens, SUM(cost) AS cost FROM token_usage GROUP BY provider`); if err != nil { return nil, err }; totalTokens := 0; totalCost := 0.0; for _, row := range rows { totalTokens += intVal(row["tokens"]); totalCost += floatVal(row["cost"]) }; return map[string]interface{}{"providers":rows,"tokens":totalTokens,"cost":totalCost,"limit":cfg.BudgetLimit}, nil }
func Errors(cfg config.Config) ([]map[string]interface{}, error) { f, err := os.Open(cfg.AuditPath); if err != nil { if os.IsNotExist(err) { return nil, nil }; return nil, err }; defer f.Close(); var out []map[string]interface{}; scanner := bufio.NewScanner(f); for scanner.Scan() { var row map[string]interface{}; if err := json.Unmarshal(scanner.Bytes(), &row); err != nil { continue }; action := fmt.Sprintf("%v", row["action"]); outcome := fmt.Sprintf("%v", row["outcome"]); if strings.Contains(action, "blocked") || strings.Contains(action, "rejected") || outcome == "failed" { out = append(out, row) } }; return out, scanner.Err() }
func RenderTUI(state State, width int) string { if width < 80 { width = 80 }; left := fmt.Sprintf("Teams/Agents\nmissions:%d\nagents:%d", len(state.Missions), len(state.Agents)); center := fmt.Sprintf("Canvas\nartifacts:%d\nworkflow:%d\npanes: terminal code doc cad browser notebook", len(state.Artifacts), len(state.Workflow)); right := fmt.Sprintf("Right Rail\napprovals:%d\nerrors:%d\nbudget:%v", len(state.Approvals), len(state.Errors), state.Budget["tokens"]); return padBlock(left, width/3) + padBlock(center, width/3) + padBlock(right, width/3) }
func RenderHTML(state State, outputPath string) (string, error) { if outputPath == "" { outputPath = filepath.Join(os.TempDir(), "uclaw-desktop.html") }; tpl := `<!doctype html><html><head><meta charset="utf-8"><title>UCLAW</title><style>
body{font-family:Georgia,serif;background:linear-gradient(135deg,#f6efe0,#d7e7e2);color:#14212b;margin:0;padding:24px}
.grid{display:grid;grid-template-columns:1fr 2fr 1fr;gap:16px}
.panel{background:rgba(255,255,255,.75);padding:16px;border:1px solid #355c6f;border-radius:12px}
pre{white-space:pre-wrap}
</style></head><body><div class="grid">
<section class="panel"><h2>Sidebar</h2><pre>{{.Sidebar}}</pre></section>
<section class="panel"><h2>Canvas</h2><pre>{{.Canvas}}</pre></section>
<section class="panel"><h2>Right Rail</h2><pre>{{.Rail}}</pre></section>
</div><script id="state" type="application/json">{{.State}}</script></body></html>`; stateJSON, _ := json.Marshal(state); data := struct { Sidebar string; Canvas string; Rail string; State template.JS }{Sidebar:fmt.Sprintf("Missions: %d\nAgents: %d", len(state.Missions), len(state.Agents)),Canvas:fmt.Sprintf("Artifacts: %d\nWorkflow Queue: %d\nSix panes: terminal, code, doc, CAD, browser, notebook", len(state.Artifacts), len(state.Workflow)),Rail:fmt.Sprintf("Approvals: %d\nErrors: %d\nBudget tokens: %v", len(state.Approvals), len(state.Errors), state.Budget["tokens"]),State:template.JS(string(stateJSON))}; f, err := os.Create(outputPath); if err != nil { return "", err }; defer f.Close(); if err := template.Must(template.New("desktop").Parse(tpl)).Execute(f, data); err != nil { return "", err }; return outputPath, nil }
func padBlock(text string, width int) string { lines := strings.Split(text, "\n"); for i, line := range lines { if len(line) < width-1 { lines[i] = line + strings.Repeat(" ", width-1-len(line)) } }; return strings.Join(lines, "\n") + "\n" }
func dirSize(root string) int64 { var total int64; _ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error { if err == nil && !info.IsDir() { total += info.Size() }; return nil }); return total }
func fileSize(path string) int64 { info, err := os.Stat(path); if err != nil { return 0 }; return info.Size() }
func mustCheckpoints(ctx context.Context, cfg config.Config) []missions.Checkpoint { var out []missions.Checkpoint; list, _ := missions.List(ctx, cfg); for _, m := range list { cp, _ := missions.ListCheckpoints(ctx, cfg, m.ID); out = append(out, cp...) }; return out }
func intVal(v interface{}) int { var n int; fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &n); return n }
func floatVal(v interface{}) float64 { var n float64; fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &n); return n }
