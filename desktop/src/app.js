async function loadState() {
  const response = await fetch("./state.json");
  if (!response.ok) {
    throw new Error("failed to load state");
  }
  return response.json();
}

function render(state) {
  const sidebar = document.getElementById("sidebar");
  const canvas = document.getElementById("canvas");
  const rail = document.getElementById("rail");

  sidebar.textContent = [
    `missions: ${state.missions.length}`,
    `agents: ${state.agents.length}`,
    `approvals: ${state.approvals.length}`
  ].join("\n");

  canvas.textContent = [
    `artifacts: ${state.artifacts.length}`,
    `workflow: ${state.workflow_queue.length}`,
    "panes: terminal code doc cad browser notebook"
  ].join("\n");

  rail.textContent = [
    `errors: ${state.errors.length}`,
    `tokens: ${state.budget.tokens || 0}`,
    `timeline: ${state.timeline.length}`
  ].join("\n");
}

loadState().then(render).catch((error) => {
  document.body.dataset.error = "true";
  document.getElementById("canvas").textContent = error.message;
});
