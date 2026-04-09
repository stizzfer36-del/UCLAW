#!/usr/bin/env python3
"""AutoGen orchestration runner for UCLAW.

Full SDK: pip install autogen-agentchat autogen-ext[openai]
Docs: https://microsoft.github.io/autogen/
"""
import sys, json, asyncio

payload = json.loads(sys.argv[1] if len(sys.argv) > 1 else '{}')

async def main():
    task = payload.get('task', 'no task')
    model = payload.get('model', 'gpt-4o')
    agents = payload.get('agents', [])

    try:
        from autogen_agentchat.agents import AssistantAgent
        from autogen_agentchat.teams import RoundRobinGroupChat
        from autogen_agentchat.ui import Console
        from autogen_ext.models.openai import OpenAIChatCompletionClient

        client = OpenAIChatCompletionClient(model=model)
        agent_objs = [AssistantAgent(name=a, model_client=client) for a in (agents or ['assistant'])]
        team = RoundRobinGroupChat(agent_objs)
        await Console(team.run_stream(task=task))
    except ImportError:
        # Graceful stub when SDK not installed
        print(json.dumps({'result': f'[stub] would run task: {task} with agents {agents}'}))

asyncio.run(main())
