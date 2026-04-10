package hardening

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/stizzfer36-del/UCLAW/internal/config"
)

type peerPackage struct {
	Peer       string     `json:"peer"`
	ExportedAt string     `json:"exported_at"`
	Files      []peerFile `json:"files"`
}

type peerFile struct {
	Path        string `json:"path"`
	Hash        string `json:"hash"`
	BaseHash    string `json:"base_hash,omitempty"`
	ContentB64  string `json:"content_b64"`
	BaseB64     string `json:"base_b64,omitempty"`
	ContentType string `json:"content_type"`
}

type peerState struct {
	Peer  string               `json:"peer"`
	Files map[string]stateFile `json:"files"`
}

type stateFile struct {
	Hash       string `json:"hash"`
	ContentB64 string `json:"content_b64"`
}

type mergeRule string

const (
	mergeRuleConflict    mergeRule = "conflict"
	mergeRuleText3Way    mergeRule = "text-3way"
	mergeRuleAppendLines mergeRule = "append-lines"
)

func exportPeerPackage(cfg config.Config, peer string) (string, error) {
	state, _ := loadPeerState(cfg, peer)
	pkg := peerPackage{Peer: peer, ExportedAt: timestamp(), Files: []peerFile{}}
	files, err := collectSyncFiles(cfg)
	if err != nil {
		return "", err
	}
	for _, rel := range files {
		body, err := os.ReadFile(filepath.Join(cfg.Home, rel))
		if err != nil {
			return "", err
		}
		entry := peerFile{
			Path:        rel,
			Hash:        hashBytes(body),
			ContentB64:  base64.StdEncoding.EncodeToString(body),
			ContentType: contentType(body),
		}
		if prior, ok := state.Files[rel]; ok {
			entry.BaseHash = prior.Hash
			entry.BaseB64 = prior.ContentB64
		}
		pkg.Files = append(pkg.Files, entry)
	}
	body, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(cfg.Home, "sync-"+peer+".json")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func importPeerPackage(cfg config.Config, path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return importPeerPackageBytes(cfg, body)
}

func importPeerPackageBytes(cfg config.Config, body []byte) error {
	var pkg peerPackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		return err
	}
	state, _ := loadPeerState(cfg, pkg.Peer)
	if state.Files == nil {
		state.Files = map[string]stateFile{}
	}
	conflictDir := filepath.Join(cfg.Home, "sync-conflicts", pkg.Peer)
	if err := os.MkdirAll(conflictDir, 0o755); err != nil {
		return err
	}
	for _, file := range pkg.Files {
		target := filepath.Join(cfg.Home, filepath.Clean(file.Path))
		if !strings.HasPrefix(target, filepath.Clean(cfg.Home)+string(os.PathSeparator)) && target != filepath.Clean(cfg.Home) {
			return fmt.Errorf("sync path escapes home: %s", file.Path)
		}
		remoteBody, err := base64.StdEncoding.DecodeString(file.ContentB64)
		if err != nil {
			return err
		}
		baseBody, _ := base64.StdEncoding.DecodeString(file.BaseB64)
		localBody, _ := os.ReadFile(target)
		localHash := hashBytes(localBody)
		remoteHash := file.Hash
		baseHash := file.BaseHash
		rule := classifyMergeRule(file.Path, file.ContentType)

		switch {
		case localHash == remoteHash:
		case localHash == "" || localHash == baseHash:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(target, remoteBody, 0o644); err != nil {
				return err
			}
		case remoteHash == baseHash:
		default:
			merged, ok := mergeByRule(rule, baseBody, localBody, remoteBody)
			if ok {
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(target, merged, 0o644); err != nil {
					return err
				}
			} else {
				if err := writeConflict(conflictDir, file.Path, baseBody, localBody, remoteBody, localHash, remoteHash); err != nil {
					return err
				}
				continue
			}
		}
		state.Files[file.Path] = stateFile{Hash: remoteHash, ContentB64: file.ContentB64}
	}
	return savePeerState(cfg, pkg.Peer, state)
}

func classifyMergeRule(rel, contentType string) mergeRule {
	rel = filepath.ToSlash(strings.ToLower(rel))
	base := filepath.Base(rel)
	ext := strings.ToLower(filepath.Ext(rel))
	switch {
	case base == "world.db" || base == "graph.db":
		return mergeRuleConflict
	case ext == ".db" || ext == ".sqlite" || ext == ".sqlite3":
		return mergeRuleConflict
	case base == "audit.jsonl" || ext == ".jsonl" || strings.Contains(rel, "/logs/"):
		return mergeRuleAppendLines
	case strings.HasPrefix(rel, "vault/notes/") && (ext == ".md" || ext == ".txt"):
		return mergeRuleAppendLines
	case ext == ".md" || ext == ".txt" || ext == ".html" || ext == ".css" || ext == ".js" || ext == ".go" || ext == ".sh" || ext == ".yaml" || ext == ".yml":
		return mergeRuleText3Way
	case contentType == "text":
		return mergeRuleText3Way
	default:
		return mergeRuleConflict
	}
}

func mergeByRule(rule mergeRule, base, local, remote []byte) ([]byte, bool) {
	switch rule {
	case mergeRuleAppendLines:
		return mergeAppendLines(base, local, remote)
	case mergeRuleText3Way:
		return mergeText(base, local, remote)
	default:
		return nil, false
	}
}

func collectSyncFiles(cfg config.Config) ([]string, error) {
	var files []string
	err := filepath.Walk(cfg.Home, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "sync-conflicts" || base == "peers" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Base(path) == ".env" || strings.HasPrefix(filepath.Base(path), "sync-") {
			return nil
		}
		rel, err := filepath.Rel(cfg.Home, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func loadPeerState(cfg config.Config, peer string) (peerState, error) {
	path := filepath.Join(cfg.PeersPath, peer, "state.json")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return peerState{Peer: peer, Files: map[string]stateFile{}}, nil
		}
		return peerState{}, err
	}
	var state peerState
	if err := json.Unmarshal(body, &state); err != nil {
		return peerState{}, err
	}
	if state.Files == nil {
		state.Files = map[string]stateFile{}
	}
	return state, nil
}

func savePeerState(cfg config.Config, peer string, state peerState) error {
	path := filepath.Join(cfg.PeersPath, peer, "state.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func mergeText(base, local, remote []byte) ([]byte, bool) {
	if !isText(local) || !isText(remote) || (len(base) > 0 && !isText(base)) {
		return nil, false
	}
	baseStr := string(base)
	localStr := string(local)
	remoteStr := string(remote)
	switch {
	case localStr == remoteStr:
		return local, true
	case localStr == baseStr:
		return remote, true
	case remoteStr == baseStr:
		return local, true
	case strings.HasPrefix(localStr, baseStr) && strings.HasPrefix(remoteStr, baseStr):
		localSuffix := strings.TrimPrefix(localStr, baseStr)
		remoteSuffix := strings.TrimPrefix(remoteStr, baseStr)
		return []byte(baseStr + localSuffix + dedupeSuffix(localSuffix, remoteSuffix)), true
	default:
		return nil, false
	}
}

func mergeAppendLines(base, local, remote []byte) ([]byte, bool) {
	if !isText(local) || !isText(remote) || (len(base) > 0 && !isText(base)) {
		return nil, false
	}
	switch {
	case string(local) == string(remote):
		return local, true
	case string(local) == string(base):
		return remote, true
	case string(remote) == string(base):
		return local, true
	}

	seen := map[string]bool{}
	var merged []string
	for _, body := range [][]byte{base, local, remote} {
		for _, line := range strings.Split(string(body), "\n") {
			if line == "" {
				continue
			}
			if seen[line] {
				continue
			}
			seen[line] = true
			merged = append(merged, line)
		}
	}
	if len(merged) == 0 {
		return []byte{}, true
	}
	return []byte(strings.Join(merged, "\n") + "\n"), true
}

func dedupeSuffix(localSuffix, remoteSuffix string) string {
	if remoteSuffix == "" || remoteSuffix == localSuffix {
		return ""
	}
	if localSuffix == "" {
		return remoteSuffix
	}
	lines := strings.Split(remoteSuffix, "\n")
	var out []string
	for _, line := range lines {
		if line == "" && len(out) == 0 {
			continue
		}
		if !strings.Contains(localSuffix, line) {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

func writeConflict(conflictDir, rel string, base, local, remote []byte, localHash, remoteHash string) error {
	stem := filepath.Join(conflictDir, strings.ReplaceAll(rel, string(os.PathSeparator), "_"))
	if err := os.MkdirAll(filepath.Dir(stem), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(stem+".base", base, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(stem+".local", local, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(stem+".remote", remote, 0o644); err != nil {
		return err
	}
	return os.WriteFile(stem+".conflict", []byte(fmt.Sprintf("local=%s\nremote=%s\npath=%s\n", localHash, remoteHash, rel)), 0o644)
}

func hashBytes(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return hashString(string(body))
}

func hashString(value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

func contentType(body []byte) string {
	if isText(body) {
		return "text"
	}
	return "binary"
}

func isText(body []byte) bool {
	for _, b := range body {
		if b == 0 {
			return false
		}
	}
	return true
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
