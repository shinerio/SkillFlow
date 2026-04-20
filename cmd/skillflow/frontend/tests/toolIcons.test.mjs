import test from 'node:test'
import assert from 'node:assert/strict'
import { getToolIconConfig } from '../.tmp-tests/src/config/toolIcons.js'

test('getToolIconConfig returns branded config for Copilot', () => {
  const config = getToolIconConfig('copilot')

  assert.equal(config.color, '#000000')
  assert.equal(typeof config.svg, 'function')
})

test('getToolIconConfig falls back for unknown tools', () => {
  const config = getToolIconConfig('unknown-tool')

  assert.equal(config.color, '#6b7280')
  assert.equal(typeof config.svg, 'function')
})
