import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildToolSkillsGroupSidebarItems,
  buildToolSkillsNavItems,
  extractConfiguredAgentSkillGroups,
  filterAllSkillsByGroup,
  filterManagedSkills,
  getSettledValue,
  getToolSkillsContentScrollMode,
  groupManagedSkills,
  isMemoryPanelAvailable,
  resolveSkillGroupDropTarget,
  splitAgentSkillEntries,
} from '../.tmp-tests/src/lib/toolSkillsManagement.js'

test('buildToolSkillsNavItems prepends all entry', () => {
  assert.deepEqual(
    buildToolSkillsNavItems([{ name: 'codex' }, { name: 'claude-code' }]).map(item => item.key),
    ['all', 'codex', 'claude-code'],
  )
})

test('isMemoryPanelAvailable returns false for all', () => {
  assert.equal(isMemoryPanelAvailable('all'), false)
  assert.equal(isMemoryPanelAvailable('codex'), true)
})

test('groupManagedSkills uses Ungrouped fallback', () => {
  const groups = groupManagedSkills([
    { name: 'react-expert', groupName: 'frontend' },
    { name: 'go-reviewer' },
  ])
  assert.equal(groups[0].groupName, 'frontend')
  assert.equal(groups[1].groupName, 'Ungrouped')
})

test('buildToolSkillsGroupSidebarItems merges config groups with skills', () => {
  const sidebar = buildToolSkillsGroupSidebarItems(['frontend'], [
    { name: 'react-expert', groupName: 'frontend' },
    { name: 'go-reviewer', groupName: 'backend' },
    { name: 'prompt-writer' },
  ])
  assert.equal(sidebar[0].kind, 'all')
  assert.equal(sidebar[1].kind, 'ungrouped')
  assert.deepEqual(sidebar.slice(2).map(item => item.label), ['frontend', 'backend'])
})

test('extractConfiguredAgentSkillGroups supports daemon config payload field casing', () => {
  assert.deepEqual(
    extractConfiguredAgentSkillGroups({
      agentSkillManagement: {
        Groups: ['frontend', 'backend'],
      },
    }),
    ['frontend', 'backend'],
  )
  assert.deepEqual(
    extractConfiguredAgentSkillGroups({
      agentSkillManagement: {
        groups: ['design'],
      },
    }),
    ['design'],
  )
})

test('getToolSkillsContentScrollMode uses internal pane scrolling for all-agents view only', () => {
  assert.equal(getToolSkillsContentScrollMode('all'), 'pane')
  assert.equal(getToolSkillsContentScrollMode('codex'), 'page')
  assert.equal(getToolSkillsContentScrollMode('claude-code'), 'page')
})

test('filterAllSkillsByGroup restricts based on the selected group', () => {
  const skills = [
    { name: 'frontend-master', groupName: 'frontend' },
    { name: 'backend-master', groupName: 'backend' },
    { name: 'legacy-ungrouped', groupName: 'Ungrouped' },
    { name: 'detached-skill' },
  ]
  assert.equal(filterAllSkillsByGroup(skills, 'all').length, 4)
  assert.deepEqual(filterAllSkillsByGroup(skills, 'frontend').map(skill => skill.name), ['frontend-master'])
  assert.deepEqual(filterAllSkillsByGroup(skills, 'Ungrouped').map(skill => skill.name), ['legacy-ungrouped', 'detached-skill'])
})

test('resolveSkillGroupDropTarget switches between assign and clear', () => {
  assert.deepEqual(resolveSkillGroupDropTarget('frontend'), { type: 'assign', groupName: 'frontend' })
  assert.deepEqual(resolveSkillGroupDropTarget('Ungrouped'), { type: 'clear' })
  assert.deepEqual(resolveSkillGroupDropTarget(' all '), { type: 'clear' })
})

test('filterManagedSkills searches name only', () => {
  const filtered = filterManagedSkills([
    { name: 'React Expert', groupName: 'frontend' },
    { name: 'Go Reviewer', groupName: 'backend' },
  ], 'react', 'asc')
  assert.deepEqual(filtered.map(item => item.name), ['React Expert'])
})

test('splitAgentSkillEntries separates push and scan-only entries', () => {
  const split = splitAgentSkillEntries([
    { name: 'react-expert', pushed: true, seenInAgentScan: true },
    { name: 'go-reviewer', pushed: false, seenInAgentScan: true },
    { name: 'prompt-writer', pushed: false, seenInAgentScan: false },
  ])

  assert.deepEqual(split.pushSkills.map(item => item.name), ['react-expert'])
  assert.deepEqual(split.scanOnlySkills.map(item => item.name), ['go-reviewer'])
})

test('getSettledValue keeps fulfilled values when sibling requests fail', () => {
  assert.deepEqual(
    getSettledValue({ status: 'fulfilled', value: [{ name: 'codex' }] }, []),
    [{ name: 'codex' }],
  )
  assert.deepEqual(
    getSettledValue({ status: 'rejected', reason: new Error('boom') }, []),
    [],
  )
})
