import test from 'node:test'
import assert from 'node:assert/strict'
import { buildMemoryPushStatusEntries, getMemoryAgentLabel, getMemoryDrawerMetrics } from '../.tmp-tests/src/lib/memoryUi.js'

test('getMemoryDrawerMetrics widens the drawer but keeps practical bounds', () => {
  assert.deepEqual(getMemoryDrawerMetrics(1600), {
    width: 960,
    maxWidth: 960,
    minWidth: 520,
  })

  assert.deepEqual(getMemoryDrawerMetrics(1000), {
    width: 720,
    maxWidth: 960,
    minWidth: 520,
  })

  assert.deepEqual(getMemoryDrawerMetrics(480), {
    width: 520,
    maxWidth: 960,
    minWidth: 520,
  })
})

test('buildMemoryPushStatusEntries preserves enabled-agent order and labels', () => {
  assert.deepEqual(
    buildMemoryPushStatusEntries(
      [{ name: 'codex' }, { name: 'claude-code' }, { name: 'custom-agent' }],
      { codex: 'synced', 'claude-code': 'pendingPush' },
    ),
    [
      { agentType: 'codex', label: 'Codex', status: 'synced' },
      { agentType: 'claude-code', label: 'Claude Code', status: 'pendingPush' },
      { agentType: 'custom-agent', label: 'custom-agent', status: 'neverPushed' },
    ],
  )
})

test('getMemoryAgentLabel returns Copilot brand name', () => {
  assert.equal(getMemoryAgentLabel('copilot'), 'Copilot')
  assert.equal(getMemoryAgentLabel('custom-agent'), 'custom-agent')
})
