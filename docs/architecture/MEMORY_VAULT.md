# Memory Vault Architecture

## Purpose

The memory vault is UCLAW's knowledge graph — a local-first, markdown-native store of everything the habitat knows:
decisions, sources, prompts, agent logs, meeting notes, research, and links between them.

---

## Storage Structure

```
~/.uclaw/vault/
├── decisions/          # Architecture decisions, policy changes
├── prompts/            # Saved prompt templates, agent system prompts
├── sources/            # Cached source documents, citations
├── logs/               # Agent action logs per mission
├── notes/              # Freeform engineering notes
├── research/           # Research outputs, summaries
└── graph.db            # SQLite graph index (nodes + edges)
```

---

## Node Types

| Type | Description |
|---|---|
| `decision` | Architectural or policy choice |
| `prompt` | Agent or user prompt template |
| `source` | External URL, paper, doc reference |
| `log` | Agent action log entry |
| `note` | Engineer freeform note |
| `artifact` | Link to a tracked artifact |
| `mission` | Link to a mission record |

---

## Edge Types

| Edge | Meaning |
|---|---|
| `caused-by` | This node was created as a result of another |
| `verified-by` | This claim is supported by a verifier node |
| `supersedes` | This node replaces an older one |
| `cites` | This node references an external source |
| `part-of` | This node belongs to a parent mission/artifact |
| `contradicts` | This node is in tension with another |

---

## Query API

```bash
# Find all decisions that cite a source
uclaw memory query --type decision --edge cites

# Get everything related to a mission
uclaw memory graph --mission <id> --depth 2

# Search full-text across vault
uclaw memory search "capability layer security"

# Show unverified claims
uclaw memory audit
```

---

## Obsidian Compatibility

The vault folder uses standard Obsidian frontmatter + wikilinks.
You can open `~/.uclaw/vault/` directly in Obsidian as a side window.
UCLAW will not overwrite Obsidian's `.obsidian/` config folder.
