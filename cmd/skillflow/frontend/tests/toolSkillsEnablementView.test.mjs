import test from 'node:test'
import assert from 'node:assert/strict'
import { buildManagedSkillEnablementSections } from '../.tmp-tests/src/lib/toolSkillsManagement.js'

test('buildManagedSkillEnablementSections splits skills by enabled state and groups them', () => {
  const sections = buildManagedSkillEnablementSections([
    { name: 'frontend-master', groupName: 'frontend', enabled: true },
    { name: 'backend-master', groupName: 'backend', enabled: false },
    { name: 'uncategorized', enabled: true },
    { name: 'backend-helper', groupName: 'backend', enabled: false },
  ])

  assert.equal(sections.enabled.length, 2)
  assert.deepEqual(sections.enabled.map(group => group.groupName), ['frontend', 'Ungrouped'])
  const ungrouped = sections.enabled.find(group => group.groupName === 'Ungrouped')
  assert.ok(ungrouped)
  assert.deepEqual(ungrouped.skills.map(skill => skill.name), ['uncategorized'])

  assert.equal(sections.disabled.length, 1)
  assert.equal(sections.disabled[0].groupName, 'backend')
  assert.deepEqual(sections.disabled[0].skills.map(skill => skill.name), ['backend-helper', 'backend-master'])
})
