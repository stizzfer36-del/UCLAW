#!/usr/bin/env node
// Minimal VoltAgent runner for UCLAW.
// Full SDK: npm install @voltagent/core @voltagent/vercel-ai-provider
// This shim allows UCLAW Go code to call VoltAgent tasks via subprocess.

const arg = JSON.parse(process.argv[2] || '{}');

async function main() {
  if (arg.action === 'spawn') {
    // TODO: replace with real VoltAgent SDK call once installed
    // const { VoltAgent } = require('@voltagent/core');
    // const agent = new VoltAgent({ model: arg.model, tools: arg.tools });
    // const result = await agent.run(arg.task);
    const taskId = 'task_' + Date.now();
    console.log(JSON.stringify({ task_id: taskId, status: 'queued' }));
  } else if (arg.action === 'status') {
    // TODO: query real task store
    console.log(JSON.stringify({ status: 'running' }));
  } else {
    console.error('unknown action:', arg.action);
    process.exit(1);
  }
}

main().catch(e => { console.error(e); process.exit(1); });
