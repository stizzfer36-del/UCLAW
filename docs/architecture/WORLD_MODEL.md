# UCLAW World Model

The world model is the hierarchical state graph at the center of UCLAW.
All CLI commands, agent actions, desktop views, and policy decisions operate on this graph.

---

## Node Hierarchy

```
World (singleton)
├── id: string
├── name: string
├── created_at: timestamp
├── vault_path: local path
└── Offices []
    ├── id, name, description
    └── Teams []
        ├── id, name, role (dev | verify | research | ops)
        ├── lead_agent_id
        └── Members []
            ├── id, name, type (human | agent)
            ├── engineering_handbook_path
            ├── capabilities []
            ├── allowed_paths []
            └── Machines []
                ├── id, hostname, os
                └── Rooms []
                    ├── id, name, type (mission | workspace | archive)
                    └── Missions []
                        ├── id, title, status
                        ├── created_by, assigned_to
                        ├── Memory links []
                        ├── Artifacts []
                        ├── Verification branches []
                        └── Policy bindings []
```

---

## State Graph Storage

- Format: SQLite (primary) + JSON snapshots (for portability)
- Location: `~/.uclaw/world.db`
- Replicated on each mission checkpoint to `~/.uclaw/checkpoints/`
- Queryable via CLI: `uclaw world inspect --node <id>` and mission/status subcommands through the same local state

---

## State Transitions

| Event | Transition |
|---|---|
| `uclaw init` | Create World node, default Office + Team |
| `uclaw agent spawn <name>` | Add Member (agent) to Team |
| `uclaw mission start <title>` | Create Mission in current Room |
| Agent completes task | Mission status → `complete`, artifact added |
| Verifier signs off | Artifact verification status → `verified` |
| Policy violation | Mission status → `blocked`, audit event emitted |

The current CLI does not yet expose first-class `office` or `team` creation commands.
Those remain part of the world model design, not the present command surface.
