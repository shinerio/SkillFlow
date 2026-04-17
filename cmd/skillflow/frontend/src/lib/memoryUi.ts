export type MemoryPushStatus = 'synced' | 'pendingPush' | 'neverPushed'
export type MemoryModuleStatus = 'enabled' | 'disabled'

export type MemoryAgentLike = {
  name: string
}

export type MemoryPushStatusEntry = {
  agentType: string
  label: string
  status: MemoryPushStatus
}

const agentDisplayName: Record<string, string> = {
  'claude-code': 'Claude Code',
  copilot: 'Copilot',
  codex: 'Codex',
  'gemini-cli': 'Gemini CLI',
  opencode: 'OpenCode',
  openclaw: 'OpenClaw',
}

export function getMemoryDrawerMetrics(viewportWidth: number) {
  const maxWidth = 960
  const minWidth = 520
  return {
    width: Math.max(minWidth, Math.min(maxWidth, Math.round(viewportWidth * 0.72))),
    maxWidth,
    minWidth,
  }
}

export function getMemoryAgentLabel(agentName: string): string {
  return agentDisplayName[agentName] ?? agentName
}

export function buildMemoryPushStatusEntries(
  agents: MemoryAgentLike[],
  pushStatuses: Record<string, MemoryPushStatus>,
): MemoryPushStatusEntry[] {
  return agents.map(agent => ({
    agentType: agent.name,
    label: getMemoryAgentLabel(agent.name),
    status: pushStatuses[agent.name] ?? 'neverPushed',
  }))
}

export function getMemoryModuleStatus(enabled: boolean): MemoryModuleStatus {
  return enabled ? 'enabled' : 'disabled'
}

export function getMemoryModuleStatusColor(enabled: boolean): string {
  return enabled ? 'var(--color-success, #22c55e)' : 'var(--text-muted)'
}
