# Integration Catalogue

| Integration | Source Repo | Language | UCLAW Tool(s) | Risk |
|---|---|---|---|---|
| **Fabric** | [danielmiessler/fabric](https://github.com/danielmiessler/fabric) | Go | `fabric_run_pattern`, `fabric_list_patterns` | low |
| **Codex** | [openai/codex](https://github.com/openai/codex) | TS/Go | `codex_edit`, `codex_build` | medium |
| **Ollama / LM Studio** | [ollama/ollama](https://github.com/ollama/ollama) | Go | `llm_generate`, `llm_embed` | low |
| **VoltAgent** | [voltagent/voltagent](https://github.com/voltagent/voltagent) | TS | `voltagent_spawn`, `voltagent_status` | medium |
| **PraisonAI** | [MervinPraison/PraisonAI](https://github.com/MervinPraison/PraisonAI) | Python | `praison_run_crew` | medium |
| **AutoGen** | [microsoft/autogen](https://github.com/microsoft/autogen) | Python | via `agent_framework` | medium |
| **Observer** | local model orchestrator | Go/REST | `observer_run_task` | medium |
| **ROS2** | [ros2/ros2](https://github.com/ros2/ros2) | C++/Python | `ros_call_service`, `ros_publish_topic` | high |
| **IoT/MQTT** | mosquitto + paho | Go/CLI | `iot_toggle_device`, `iot_read_sensor` | high/low |
| **Doc Parser** | poppler + pandoc | CLI | `doc_parse` | low |
| **CAD** | [FreeCAD/FreeCAD](https://github.com/FreeCAD/FreeCAD) | Python/C++ | `cad_inspect` | low |

## Adding a New Integration

1. `mkdir core/integrations/<name>`
2. Write `adapter.go` (or `adapter.ts` / `runner.py` for non-Go).
3. Add tool entries to `core/policies/integrations.yaml`.
4. Add adapter docs to this file.
5. Ensure all calls pass through `agents.CheckTool()` in the agent runtime.
