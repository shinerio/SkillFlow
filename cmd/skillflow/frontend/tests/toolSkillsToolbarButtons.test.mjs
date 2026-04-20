import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/pages/ToolSkills.tsx', import.meta.url), 'utf8')

test('tool skills toolbar actions reuse one shared secondary button style', () => {
  assert.match(
    source,
    /const toolbarSecondaryButtonClassName = 'btn-secondary flex items-center gap-1\.5 px-3 py-1\.5 text-sm rounded-lg disabled:opacity-40'/,
  )

  const usages = source.match(/className=\{toolbarSecondaryButtonClassName\}/g) ?? []
  assert.ok(usages.length >= 3, 'expected manage, manual pull, and batch delete to share the same button class')
})
