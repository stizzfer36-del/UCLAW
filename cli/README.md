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
uclaw world restore --checkpoint <id> --confirm  # Restore world state
```

### Offices & Teams
```bash
uclaw office new <name>             # Create a new office
uclaw office list                   # List all offices
uclaw team new <name> --role dev    # Create a team in current office
uclaw team list                     # List teams
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
uclaw memory search <query>         # Full-text vault search
uclaw memory query --type decision  # Query by node type
uclaw memory graph --mission <id>   # Show mission memory graph
uclaw memory audit --unverified     # Show unverified claims
```

### Artifacts
```bash
uclaw artifact list                 # List artifacts for current mission
uclaw artifact show <id>            # Show artifact + checks
uclaw artifact flag <id>            # Flag artifact for re-verification
uclaw artifact sign <id>            # Sign off on artifact
uclaw artifact revert <id> --to <sha>  # Revert artifact to git SHA
```

### Observability
```bash
uclaw status                        # Mission + agent dashboard
uclaw audit --since 1h              # Show recent audit events
uclaw budget                        # Show token + cost usage
uclaw health                        # Show agent + provider health
uclaw pause --all                   # Pause all agents
uclaw resume --agent <id>           # Resume an agent
uclaw override --artifact <id>      # Override a blocked artifact
uclaw policy tighten --tool <name>  # Increase tool risk level
```

### Voice
```bash
uclaw voice                         # Start voice input mode
uclaw voice --tts                   # Enable TTS feedback
```
