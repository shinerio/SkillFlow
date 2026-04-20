import type { ToolSkillsPanel } from './toolSkillsPanels'

export type ToolSkillsNavItem = {
  key: string
  label: string
  kind: 'all' | 'agent'
}

export type ManagedSkillLike = {
  name: string
  groupName?: string
  enabled?: boolean
}

export type ManagedSkillGroup<TSkill extends ManagedSkillLike> = {
  groupName: string
  skills: TSkill[]
}

export type AgentSkillEntryLike = {
  pushed?: boolean
  seenInAgentScan?: boolean
}

export type ToolSkillsGroupSidebarItemKind = 'all' | 'ungrouped' | 'group'

export type ToolSkillsGroupSidebarItem = {
  key: string
  label: string
  kind: ToolSkillsGroupSidebarItemKind
}

export type SkillGroupDropTarget =
  | { type: 'assign'; groupName: string }
  | { type: 'clear' }

export type ManagedSkillEnablementSections<TSkill extends ManagedSkillLike> = {
  enabled: ManagedSkillGroup<TSkill>[]
  disabled: ManagedSkillGroup<TSkill>[]
}

export type ToolSkillsContentScrollMode = 'page' | 'pane'

const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })
export const ungroupedLabel = 'Ungrouped'
export const toolSkillsSidebarAllKey = 'all'
export const toolSkillsSidebarUngroupedKey = 'ungrouped'

export function buildToolSkillsNavItems(agents: Array<{ name: string }>): ToolSkillsNavItem[] {
  return [
    { key: 'all', label: 'All', kind: 'all' },
    ...agents.map(agent => ({ key: agent.name, label: agent.name, kind: 'agent' as const })),
  ]
}

export function buildToolSkillsGroupSidebarItems(
  configGroups: string[],
  skills: Array<{ groupName?: string }>,
): ToolSkillsGroupSidebarItem[] {
  const normalizedConfigGroups = [...new Set(
    configGroups
      .map(name => (name ?? '').trim())
      .filter(Boolean),
  )]
  const skillGroups = new Set<string>()
  for (const skill of skills) {
    const groupName = (skill.groupName ?? '').trim()
    if (groupName && groupName !== ungroupedLabel) {
      skillGroups.add(groupName)
    }
  }
  const extraSkillGroups = [...skillGroups]
    .filter(name => !normalizedConfigGroups.includes(name))
    .sort((left, right) => collator.compare(left, right))
  const sidebarGroups = [...normalizedConfigGroups, ...extraSkillGroups]
  return [
    { key: toolSkillsSidebarAllKey, label: 'All', kind: 'all' },
    { key: toolSkillsSidebarUngroupedKey, label: ungroupedLabel, kind: 'ungrouped' },
    ...sidebarGroups.map(groupName => ({ key: groupName, label: groupName, kind: 'group' as const })),
  ]
}

export function extractConfiguredAgentSkillGroups(cfg: any): string[] {
  const management = cfg?.agentSkillManagement
  const groups = Array.isArray(management?.groups)
    ? management.groups
    : Array.isArray(management?.Groups)
      ? management.Groups
      : []
  return groups
    .map((groupName: unknown) => (typeof groupName === 'string' ? groupName.trim() : ''))
    .filter(Boolean)
}

export function getToolSkillsContentScrollMode(target: string): ToolSkillsContentScrollMode {
  return target === 'all' ? 'pane' : 'page'
}

export function filterAllSkillsByGroup<T extends { groupName?: string }>(skills: T[], selectedGroupKey: string): T[] {
  const normalized = (selectedGroupKey ?? '').trim()
  if (!normalized || normalized === toolSkillsSidebarAllKey) {
    return skills
  }
  if (normalized === toolSkillsSidebarUngroupedKey || normalized === ungroupedLabel) {
    return skills.filter(skill => {
      const groupName = (skill.groupName ?? '').trim()
      return !groupName || groupName === ungroupedLabel
    })
  }
  return skills.filter(skill => (skill.groupName ?? '').trim() === normalized)
}

export function resolveSkillGroupDropTarget(targetKey: string): SkillGroupDropTarget {
  const normalized = (targetKey ?? '').trim()
  if (!normalized || normalized === toolSkillsSidebarAllKey || normalized === toolSkillsSidebarUngroupedKey || normalized === ungroupedLabel) {
    return { type: 'clear' }
  }
  return { type: 'assign', groupName: normalized }
}

function groupByGroupName<TSkill extends ManagedSkillLike>(skills: TSkill[]): ManagedSkillGroup<TSkill>[] {
  if (skills.length === 0) return []
  return groupManagedSkills(skills)
}

export function buildManagedSkillEnablementSections<TSkill extends ManagedSkillLike>(skills: TSkill[]): ManagedSkillEnablementSections<TSkill> {
  const enabled = groupByGroupName(skills.filter(skill => skill.enabled))
  const disabled = groupByGroupName(skills.filter(skill => !skill.enabled))
  return { enabled, disabled }
}

export function isMemoryPanelAvailable(target: string): boolean {
  return target !== 'all'
}

export function filterManagedSkills<TSkill extends ManagedSkillLike>(skills: TSkill[], search: string, sortOrder: 'asc' | 'desc'): TSkill[] {
  const query = search.trim().toLocaleLowerCase()
  const filtered = query
    ? skills.filter(skill => skill.name.toLocaleLowerCase().includes(query))
    : skills
  return [...filtered].sort((left, right) => {
    const result = collator.compare(left.name, right.name)
    return sortOrder === 'asc' ? result : -result
  })
}

export function groupManagedSkills<TSkill extends ManagedSkillLike>(skills: TSkill[]): ManagedSkillGroup<TSkill>[] {
  const grouped = new Map<string, TSkill[]>()
  for (const skill of skills) {
    const groupName = (skill.groupName ?? '').trim() || ungroupedLabel
    const existing = grouped.get(groupName) ?? []
    existing.push(skill)
    grouped.set(groupName, existing)
  }
  return [...grouped.entries()]
    .sort((left, right) => collator.compare(left[0], right[0]))
    .map(([groupName, groupedSkills]) => ({
      groupName,
      skills: [...groupedSkills].sort((left, right) => collator.compare(left.name, right.name)),
    }))
}

export function splitAgentSkillEntries<TSkill extends ManagedSkillLike & AgentSkillEntryLike>(skills: TSkill[]) {
  return {
    pushSkills: skills.filter(skill => skill.pushed),
    scanOnlySkills: skills.filter(skill => skill.seenInAgentScan && !skill.pushed),
  }
}

export function getSettledValue<T>(result: PromiseSettledResult<T>, fallback: T): T {
  return result.status === 'fulfilled' ? result.value : fallback
}

export function getToolSkillsSearchPlaceholder(
  target: string,
  activePanel: ToolSkillsPanel,
  t: (key: 'toolSkills.searchAllPlaceholder' | 'toolSkills.searchMemoryPlaceholder' | 'toolSkills.searchSkillsPlaceholder') => string,
): string {
  if (target === 'all') return t('toolSkills.searchAllPlaceholder')
  return activePanel === 'memory'
    ? t('toolSkills.searchMemoryPlaceholder')
    : t('toolSkills.searchSkillsPlaceholder')
}
