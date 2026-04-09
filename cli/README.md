# UCLAW CLI

The CLI is the launch and scripting surface for the UCLAW habitat.

## Entry Point

```bash
uclaw [command] [flags]
```

## Core Commands

### World
```bash
uclaw init                          # Initialize a new UCLAW world
uclaw world inspect                 # Show world state summary
uclaw world inspect --node <id>     # Inspect a specific node
```

### Agents
```bash
uclaw agent spawn <name> --team <team_id> --role dev   # Spawn an agent
uclaw agent list                    # List active agents
uclaw agent inspect <id>            # Show agent details + handbook
uclaw agent pause <id>              # Pause an agent
uclaw agent resume <id>             # Resume an agent
uclaw agent retire <id>             # Retire an agent (preserve state)
```

### Missions
```bash
uclaw mission start <title>         # Start a new mission
uclaw mission list                  # List missions
uclaw mission status <id>           # Show mission status + checks
uclaw mission checkpoints <id>      # List checkpoints
uclaw mission replay <id> --from <checkpoint_id>  # Replay from checkpoint
uclaw mission rollback <id>         # Roll back mission state
```

### Memory
```bash
uclaw memory create --type decision --title "ADR" --content "..."   # Create a vault node
uclaw memory link --from <node_a> --to <node_b> --edge caused-by    # Link two nodes
uclaw memory search <query>         # Full-text vault search
uclaw memory query --type decision  # Query by node type
uclaw memory graph --mission <id>   # Show mission memory graph
uclaw memory audit                  # Show unverified claims
```

### Artifacts
```bash
uclaw artifact create --title "Spec" --type doc --path /tmp/spec.md --agent <id> --mission <id> --sources url:https://example.com
uclaw artifact list --mission <id>  # List artifacts for a mission
uclaw artifact show <id>            # Show artifact + checks
uclaw artifact verify <id> --verifier <id> --test-command "true" --workspace ~/.uclaw
uclaw artifact flag <id>            # Flag artifact for re-verification
uclaw artifact sign <id> --by <agent_id>  # Sign off on artifact
uclaw artifact revert <id> --to <sha>  # Revert artifact to git SHA
```

### Observability
```bash
uclaw status                        # Mission + agent dashboard
uclaw audit                         # Show full audit log
uclaw audit verify                  # Verify audit chain integrity
uclaw budget                        # Show token + cost usage
uclaw health                        # Show agent + provider health
uclaw pause --all                   # Pause all agents
uclaw resume --agent <id>           # Resume an agent
uclaw override --artifact <id>      # Override a blocked artifact
uclaw policy tighten --tool <name>  # Increase tool risk level
```

### Voice
```bash
uclaw voice --text "start mission alpha"                   # Dispatch from transcript
uclaw voice --file /tmp/brief.txt                         # Dispatch from transcript file
uclaw voice --live --capture-command "..." --stt-command "..."  # Capture and transcribe locally
uclaw voice --audio /tmp/clip.wav --stt-command "..." --transcript-only  # Transcribe only
```

### Sync
```bash
uclaw sync export peer-a                                  # Export peer package to disk
uclaw sync import /tmp/sync-peer-a.json                   # Import peer package from disk
uclaw sync serve --listen 127.0.0.1:44144                 # Serve HTTP peer transport
uclaw sync pull --from http://127.0.0.1:44144 --peer peer-a  # Pull package over HTTP
uclaw sync push --to http://127.0.0.1:44144 --peer peer-a    # Push package over HTTP
```
