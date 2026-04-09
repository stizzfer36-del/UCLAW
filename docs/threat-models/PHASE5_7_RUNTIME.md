# Threat Model: Desktop, Voice, and Peer Sync

## Components

- Electron desktop shell assets under `desktop/`
- local desktop state export from `uclaw desktop build`
- voice transcript dispatch and live capture/STT command execution
- peer sync package export/import and merge path

## Trust Boundaries

- local operator to desktop shell
- desktop shell to local runtime state
- microphone/audio device to capture process
- local STT command to runtime mission creation
- remote peer package to local state merge path

## Key Threats

- spoofed or stale desktop state rendered as authoritative
- malicious STT command or capture command execution
- sync package path traversal or conflict laundering
- secret leakage through sync payloads or voice artifacts
- missing audit events on late-phase runtime actions

## Existing Controls

- command-level trace IDs in audit log
- sync import path confinement to `UCLAW_HOME`
- conflict artifact recording for unresolved merges
- explicit external STT opt-in audit event
- local-first runtime state and secret scan path

## Remaining Gaps

- packaged Electron build cannot be verified in this workspace until Electron dependencies are installed
- live STT quality depends on the local transcription command provided by the operator
- binary database merge semantics remain conservative and conflict-oriented
