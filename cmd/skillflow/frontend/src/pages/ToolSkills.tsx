import { useCallback, useEffect, useMemo, useRef, useState, type MouseEvent as ReactMouseEvent } from 'react'
import { OpenPath } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import AgentSkillGroupSidebar, { type AgentSkillGroupSidebarItem } from '../components/AgentSkillGroupSidebar'
import AllAgentSkillCard from '../components/AllAgentSkillCard'
import ConflictDialog from '../components/ConflictDialog'
import ManagedAgentSkillCard from '../components/ManagedAgentSkillCard'
import SkillListControls from '../components/SkillListControls'
import SkillStatusStrip from '../components/SkillStatusStrip'
import SkillTooltip from '../components/SkillTooltip'
import SyncSkillCard from '../components/SyncSkillCard'
import AnimatedDialog from '../components/ui/AnimatedDialog'
import { ToolIcon } from '../config/toolIcons'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { buildAgentMemoryEntries } from '../lib/agentMemoryPreview'
import {
  AssignAgentSkillGroup,
  ClearAgentSkillGroup,
  CreateAgentSkillGroup,
  DeleteAgentSkill,
  DeleteAgentSkillGroup,
  GetAgentMemoryPreview,
  GetConfig,
  GetEnabledAgents,
  GetSkillMetaByPath,
  ListAgentSkills,
  ListAllAgentSkills,
  ListCategories,
  ListManagedAgentSkills,
  PullFromAgent,
  PullFromAgentForce,
  ReadSkillFileContent,
  RenameAgentSkillGroup,
  ScanAgentSkills,
  SetAgentSkillEnabled,
  SetAgentSkillGroupEnabled,
} from '../lib/backend'
import { copyTextToClipboard } from '../lib/clipboard'
import { createToolSkillsEventSubscriptions } from '../lib/dashboardSkillSettings'
import { filterAndSortSkills, type SkillSortOrder } from '../lib/skillList'
import {
  createToolSkillsPullState,
  getToolSkillsVisibleNotImportedPaths,
  isToolSkillsPullReady,
  syncToolSkillsPullVisibleSelection,
  toggleToolSkillsPullAllVisible,
  toggleToolSkillsPullPath,
  toggleToolSkillsPullVisibleNotImported,
} from '../lib/toolSkillsPullState'
import {
  buildManagedSkillEnablementSections,
  buildToolSkillsGroupSidebarItems,
  buildToolSkillsNavItems,
  extractConfiguredAgentSkillGroups,
  filterAllSkillsByGroup,
  filterManagedSkills,
  getSettledValue,
  getToolSkillsContentScrollMode,
  getToolSkillsSearchPlaceholder,
  isMemoryPanelAvailable,
  resolveSkillGroupDropTarget,
  splitAgentSkillEntries,
  toolSkillsSidebarAllKey,
  ungroupedLabel,
  type ManagedSkillGroup,
} from '../lib/toolSkillsManagement'
import { getDefaultToolSkillsPanel, type ToolSkillsPanel } from '../lib/toolSkillsPanels'
import { subscribeToEvents } from '../lib/wailsEvents'
import {
  AlertCircle,
  ArrowDownToLine,
  ArrowUpToLine,
  Brain,
  Check,
  CheckSquare,
  Copy,
  FolderOpenDot,
  Layers3,
  RefreshCw,
  ScanLine,
} from 'lucide-react'

type AgentMemoryRuleItem = {
  name: string
  path: string
  content: string
  managed: boolean
}

type AgentMemoryPreviewItem = {
  agentName: string
  memoryPath: string
  rulesDir: string
  mainExists: boolean
  mainContent: string
  rulesDirExists: boolean
  rules: AgentMemoryRuleItem[]
}

type ManagedSkill = {
  name: string
  groupName?: string
  enabled: boolean
  paths: string[]
}

type AllSkill = {
  name: string
  groupName?: string
  agents: string[]
  instanceCount: number
}

type AgentSkillEntry = {
  name: string
  path: string
  source?: string
  imported?: boolean
  updatable?: boolean
  pushed?: boolean
  pushedAgents?: string[]
  seenInAgentScan?: boolean
}

const defaultCategory = 'Default'
type GroupDialogMode = 'create' | 'rename' | null
type AgentSkillViewMode = 'browse' | 'manage'

export default function ToolSkills() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('myAgents')
  const [agents, setAgents] = useState<any[]>([])
  const [groupNames, setGroupNames] = useState<string[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedTarget, setSelectedTarget] = useState<string>('all')
  const [managedSkills, setManagedSkills] = useState<ManagedSkill[]>([])
  const [allSkills, setAllSkills] = useState<AllSkill[]>([])
  const [agentSkills, setAgentSkills] = useState<AgentSkillEntry[]>([])
  const [memoryPreview, setMemoryPreview] = useState<AgentMemoryPreviewItem | null>(null)
  const [memoryLoading, setMemoryLoading] = useState(false)
  const [memoryError, setMemoryError] = useState('')
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [activePanel, setActivePanel] = useState<ToolSkillsPanel>(getDefaultToolSkillsPanel())
  const [groupDialogMode, setGroupDialogMode] = useState<GroupDialogMode>(null)
  const [groupDraftName, setGroupDraftName] = useState('')
  const [groupDraftSourceName, setGroupDraftSourceName] = useState('')
  const [groupDialogSaving, setGroupDialogSaving] = useState(false)
  const [groupDialogError, setGroupDialogError] = useState('')
  const [groupDeleteTarget, setGroupDeleteTarget] = useState<string | null>(null)
  const [selectedAllGroupKey, setSelectedAllGroupKey] = useState<string>(toolSkillsSidebarAllKey)
  const [draggingAllSkillName, setDraggingAllSkillName] = useState<string | null>(null)
  const [agentSkillViewMode, setAgentSkillViewMode] = useState<AgentSkillViewMode>('browse')
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [deleting, setDeleting] = useState(false)
  const [pullMode, setPullMode] = useState(false)
  const [scanned, setScanned] = useState<any[]>([])
  const [scanning, setScanning] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [pullState, setPullState] = useState(createToolSkillsPullState(defaultCategory))
  const [scanError, setScanError] = useState('')
  const [pullConflicts, setPullConflicts] = useState<string[]>([])
  const [pullDone, setPullDone] = useState(false)

  const resetSkillModes = useCallback(() => {
    setAgentSkillViewMode('browse')
    setSelectMode(false)
    setSelectedPaths(new Set())
    setPullMode(false)
    setScanned([])
    setScanning(false)
    setPulling(false)
    setPullState(createToolSkillsPullState(categories[0] ?? defaultCategory))
    setScanError('')
    setPullConflicts([])
    setPullDone(false)
  }, [categories])

  const loadConfig = useCallback(async () => {
    const [cfg, listedCategories] = await Promise.all([GetConfig(), ListCategories()])
    setGroupNames(extractConfiguredAgentSkillGroups(cfg))
    setCategories((listedCategories ?? []).filter(Boolean))
  }, [])

  const loadAllView = useCallback(async () => {
    setLoading(true)
    try {
      const [enabledAgentsResult, entriesResult] = await Promise.allSettled([GetEnabledAgents(), ListAllAgentSkills()])
      setAgents(getSettledValue(enabledAgentsResult, []) ?? [])
      setAllSkills(getSettledValue(entriesResult, []) ?? [])
      setManagedSkills([])
      setAgentSkills([])
    } finally {
      setLoading(false)
    }
  }, [])

  const loadAgentView = useCallback(async (agentName: string) => {
    setLoading(true)
    try {
      const [enabledAgentsResult, managedEntriesResult, detailedEntriesResult] = await Promise.allSettled([
        GetEnabledAgents(),
        ListManagedAgentSkills(agentName),
        ListAgentSkills(agentName),
      ])
      setAgents(getSettledValue(enabledAgentsResult, []) ?? [])
      setManagedSkills(getSettledValue(managedEntriesResult, []) ?? [])
      setAgentSkills(getSettledValue(detailedEntriesResult, []) ?? [])
      setAllSkills([])
    } finally {
      setLoading(false)
    }
  }, [])

  const loadMemoryPreview = useCallback(async (agentName: string) => {
    setMemoryLoading(true)
    setMemoryError('')
    try {
      const preview = await GetAgentMemoryPreview(agentName)
      setMemoryPreview((preview ?? null) as AgentMemoryPreviewItem | null)
    } catch (error) {
      console.error('Failed to load agent memory preview', error)
      setMemoryPreview(null)
      setMemoryError(t('toolSkills.memoryLoadFailed'))
    } finally {
      setMemoryLoading(false)
    }
  }, [t])

  const reloadSelected = useCallback(async () => {
    await loadConfig()
    if (selectedTarget === 'all') {
      await loadAllView()
      setMemoryPreview(null)
      setMemoryError('')
      return
    }
    await loadAgentView(selectedTarget)
    if (activePanel === 'memory') {
      await loadMemoryPreview(selectedTarget)
    }
  }, [activePanel, loadAgentView, loadAllView, loadConfig, loadMemoryPreview, selectedTarget])

  useEffect(() => {
    void (async () => {
      await loadConfig()
      await loadAllView()
    })()
  }, [loadAllView, loadConfig])

  useEffect(() => {
    if (selectedTarget === 'all') return
    return subscribeToEvents(EventsOn, createToolSkillsEventSubscriptions(() => { void reloadSelected() }))
  }, [reloadSelected, selectedTarget])

  const navItems = useMemo(() => buildToolSkillsNavItems(agents), [agents])
  const agentNavItems = useMemo(
    () => navItems.filter(item => item.kind === 'agent'),
    [navItems],
  )
  const availableCategories = useMemo(
    () => categories.length > 0 ? categories : [defaultCategory],
    [categories],
  )
  const allGroupSidebarSourceItems = useMemo(
    () => buildToolSkillsGroupSidebarItems(groupNames, allSkills),
    [allSkills, groupNames],
  )
  const mergedGroupNames = useMemo(
    () => allGroupSidebarSourceItems
      .filter(item => item.kind === 'group')
      .map(item => item.label),
    [allGroupSidebarSourceItems],
  )
  const allGroupSidebarItems = useMemo<AgentSkillGroupSidebarItem[]>(
    () => allGroupSidebarSourceItems.map(item => ({
      ...item,
      label: item.kind === 'all'
        ? t('category.all')
        : item.kind === 'ungrouped'
          ? t('toolSkills.noGroup')
          : item.label,
      count: filterAllSkillsByGroup(allSkills, item.key).length,
    })),
    [allGroupSidebarSourceItems, allSkills, t],
  )
  const selectedAllGroupItem = useMemo(
    () => allGroupSidebarItems.find(item => item.key === selectedAllGroupKey) ?? allGroupSidebarItems[0] ?? null,
    [allGroupSidebarItems, selectedAllGroupKey],
  )

  const filteredAllSkills = useMemo(
    () => filterManagedSkills(filterAllSkillsByGroup(allSkills, selectedAllGroupKey), search, sortOrder),
    [allSkills, search, selectedAllGroupKey, sortOrder],
  )
  const filteredManagedSkills = useMemo(
    () => filterManagedSkills(managedSkills, search, sortOrder),
    [managedSkills, search, sortOrder],
  )
  const managedEnablementSections = useMemo(
    () => buildManagedSkillEnablementSections(filteredManagedSkills),
    [filteredManagedSkills],
  )
  const groupedEnabledManagedSkills = useMemo(
    () => managedEnablementSections.enabled,
    [managedEnablementSections],
  )
  const groupedDisabledManagedSkills = useMemo(
    () => managedEnablementSections.disabled,
    [managedEnablementSections],
  )
  const { pushSkills, scanOnlySkills } = useMemo(
    () => splitAgentSkillEntries(agentSkills),
    [agentSkills],
  )
  const filteredPushSkills = useMemo(
    () => filterAndSortSkills(pushSkills, search, sortOrder, skill => skill.name ?? ''),
    [pushSkills, search, sortOrder],
  )
  const filteredScanOnlySkills = useMemo(
    () => filterAndSortSkills(scanOnlySkills, search, sortOrder, skill => skill.name ?? ''),
    [scanOnlySkills, search, sortOrder],
  )
  const filteredScanned = useMemo(
    () => filterAndSortSkills(scanned, search, sortOrder, skill => skill.name ?? ''),
    [scanned, search, sortOrder],
  )
  const visibleNotImportedPaths = useMemo(
    () => getToolSkillsVisibleNotImportedPaths(filteredScanned),
    [filteredScanned],
  )
  const memoryEntries = useMemo(
    () => buildAgentMemoryEntries(memoryPreview),
    [memoryPreview],
  )
  const filteredMemoryEntries = useMemo(() => {
    const query = search.trim().toLocaleLowerCase()
    if (!query) return memoryEntries
    return memoryEntries.filter(entry =>
      entry.title.toLocaleLowerCase().includes(query) ||
      entry.path.toLocaleLowerCase().includes(query) ||
      entry.content.toLocaleLowerCase().includes(query),
    )
  }, [memoryEntries, search])

  const mainMemoryEntry = useMemo(
    () => filteredMemoryEntries.find(entry => entry.kind === 'main') ?? null,
    [filteredMemoryEntries],
  )
  const ruleEntries = useMemo(
    () => filteredMemoryEntries.filter(entry => entry.kind === 'rule'),
    [filteredMemoryEntries],
  )

  const resultCount = selectedTarget === 'all'
    ? filteredAllSkills.length
    : activePanel === 'memory'
      ? filteredMemoryEntries.length
      : pullMode
        ? filteredScanned.length
        : agentSkillViewMode === 'manage'
          ? filteredManagedSkills.length
          : filteredPushSkills.length + filteredScanOnlySkills.length

  const searchPlaceholder = selectedTarget !== 'all' && pullMode
    ? t('toolSkills.searchPullPlaceholder')
    : getToolSkillsSearchPlaceholder(selectedTarget, activePanel, t)

  const currentAgent = useMemo(
    () => agents.find((agent) => agent.name === selectedTarget) ?? null,
    [agents, selectedTarget],
  )

  const allSelected = filteredPushSkills.length > 0 && filteredPushSkills.every(skill => selectedPaths.has(skill.path))
  const allPullSelected = filteredScanned.length > 0 && filteredScanned.every(skill => pullState.selectedPaths.includes(skill.path))
  const allNotImportedSelected = visibleNotImportedPaths.length > 0 && visibleNotImportedPaths.every(path => pullState.selectedPaths.includes(path))
  const pullReady = isToolSkillsPullReady(pullState)
  const contentScrollMode = getToolSkillsContentScrollMode(selectedTarget)
  const showingMemoryPanel = selectedTarget !== 'all' && activePanel === 'memory'
  const showingPullMode = selectedTarget !== 'all' && activePanel === 'skills' && pullMode
  const showingManageView = selectedTarget !== 'all' && activePanel === 'skills' && !pullMode && agentSkillViewMode === 'manage'
  const showingBrowseView = selectedTarget !== 'all' && activePanel === 'skills' && !pullMode && agentSkillViewMode === 'browse'
  const toolbarSecondaryButtonClassName = 'btn-secondary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40'

  const selectTarget = async (target: string) => {
    setSelectedTarget(target)
    setSelectedAllGroupKey(toolSkillsSidebarAllKey)
    setSearch('')
    setMemoryPreview(null)
    setMemoryError('')
    resetSkillModes()
    if (target === 'all') {
      await loadAllView()
      return
    }
    await loadAgentView(target)
    if (activePanel === 'memory') {
      await loadMemoryPreview(target)
    }
  }

  const selectPanel = async (panel: ToolSkillsPanel) => {
    setActivePanel(panel)
    if (panel !== 'skills') {
      resetSkillModes()
    }
    if (selectedTarget !== 'all' && panel === 'memory') {
      await loadMemoryPreview(selectedTarget)
    }
  }

  const openCreateGroupDialog = () => {
    setGroupDialogMode('create')
    setGroupDraftName('')
    setGroupDraftSourceName('')
    setGroupDialogError('')
  }

  const openRenameGroupDialog = (groupName: string) => {
    setGroupDialogMode('rename')
    setGroupDraftName(groupName)
    setGroupDraftSourceName(groupName)
    setGroupDialogError('')
  }

  const resetGroupDialogState = () => {
    setGroupDialogMode(null)
    setGroupDraftName('')
    setGroupDraftSourceName('')
    setGroupDialogError('')
  }

  const closeGroupDialog = () => {
    if (groupDialogSaving) return
    resetGroupDialogState()
  }

  const normalizedGroupDraftName = groupDraftName.trim()
  const isDuplicateGroupName = normalizedGroupDraftName !== ''
    && mergedGroupNames.some(groupName => groupName === normalizedGroupDraftName && groupName !== groupDraftSourceName)

  const submitGroupDialog = async () => {
    if (groupDialogSaving || normalizedGroupDraftName === '' || isDuplicateGroupName) return
    setGroupDialogSaving(true)
    setGroupDialogError('')
    try {
      if (groupDialogMode === 'create') {
        await CreateAgentSkillGroup(normalizedGroupDraftName)
      } else if (groupDialogMode === 'rename') {
        await RenameAgentSkillGroup(groupDraftSourceName, normalizedGroupDraftName)
        if (selectedAllGroupKey === groupDraftSourceName) {
          setSelectedAllGroupKey(normalizedGroupDraftName)
        }
      }
      resetGroupDialogState()
      await reloadSelected()
    } catch (error: any) {
      setGroupDialogError(String(error?.message ?? error ?? t('toolSkills.groupActionFailed')))
    } finally {
      setGroupDialogSaving(false)
    }
  }

  const openDeleteGroupDialog = (groupName: string) => {
    setGroupDeleteTarget(groupName)
    setGroupDialogError('')
  }

  const confirmDeleteGroup = async () => {
    if (!groupDeleteTarget) return
    setGroupDialogError('')
    try {
      await DeleteAgentSkillGroup(groupDeleteTarget)
      if (selectedAllGroupKey === groupDeleteTarget) {
        setSelectedAllGroupKey(toolSkillsSidebarAllKey)
      }
      setGroupDeleteTarget(null)
      await reloadSelected()
    } catch (error: any) {
      setGroupDialogError(String(error?.message ?? error ?? t('toolSkills.groupActionFailed')))
    }
  }

  const setSkillGroup = async (skillName: string, groupName: string) => {
    if (!groupName) {
      await ClearAgentSkillGroup(skillName)
    } else {
      await AssignAgentSkillGroup(skillName, groupName)
    }
    await reloadSelected()
  }

  const handleAllSkillDrop = async (skillName: string, target: AgentSkillGroupSidebarItem) => {
    setDraggingAllSkillName(null)
    const dropTarget = resolveSkillGroupDropTarget(target.key)
    if (dropTarget.type === 'clear') {
      await setSkillGroup(skillName, '')
      return
    }
    await setSkillGroup(skillName, dropTarget.groupName)
  }

  const toggleSkill = async (skillName: string, enabled: boolean) => {
    if (selectedTarget === 'all') return
    await SetAgentSkillEnabled(selectedTarget, skillName, enabled)
    await reloadSelected()
  }

  const toggleGroup = async (groupName: string, enabled: boolean) => {
    if (selectedTarget === 'all' || groupName === ungroupedLabel) return
    await SetAgentSkillGroupEnabled(selectedTarget, groupName, enabled)
    await reloadSelected()
  }

  const handleDelete = async (skillPath: string) => {
    if (selectedTarget === 'all') return
    await DeleteAgentSkill(selectedTarget, skillPath)
    await loadAgentView(selectedTarget)
  }

  const handleBatchDelete = async () => {
    if (selectedTarget === 'all' || selectedPaths.size === 0) return
    setDeleting(true)
    try {
      for (const path of selectedPaths) {
        await DeleteAgentSkill(selectedTarget, path)
      }
      setSelectedPaths(new Set())
      setSelectMode(false)
      await loadAgentView(selectedTarget)
    } finally {
      setDeleting(false)
    }
  }

  const toggleSelectPath = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }

  const toggleSelectAll = () => {
    const visiblePaths = filteredPushSkills.map(skill => skill.path)
    setSelectedPaths(prev => {
      const next = new Set(prev)
      if (visiblePaths.every(path => next.has(path))) {
        visiblePaths.forEach(path => next.delete(path))
      } else {
        visiblePaths.forEach(path => next.add(path))
      }
      return next
    })
  }

  useEffect(() => {
    if (!selectMode) return
    const visiblePaths = new Set(filteredPushSkills.map(skill => skill.path))
    setSelectedPaths(prev => {
      const next = new Set([...prev].filter(path => visiblePaths.has(path)))
      return next.size === prev.size ? prev : next
    })
  }, [filteredPushSkills, selectMode])

  useEffect(() => {
    if (!pullMode) return
    const nextPaths = syncToolSkillsPullVisibleSelection(
      pullState.selectedPaths,
      filteredScanned.map(skill => skill.path),
    )
    const same = nextPaths.length === pullState.selectedPaths.length
      && nextPaths.every((path, index) => path === pullState.selectedPaths[index])
    if (same) return
    setPullState(prev => ({ ...prev, selectedPaths: nextPaths }))
  }, [filteredScanned, pullMode, pullState.selectedPaths])

  useEffect(() => {
    if (selectedTarget !== 'all') return
    if (allGroupSidebarItems.some(item => item.key === selectedAllGroupKey)) return
    setSelectedAllGroupKey(toolSkillsSidebarAllKey)
  }, [allGroupSidebarItems, selectedAllGroupKey, selectedTarget])

  const scanForPull = useCallback(async (agentName: string) => {
    setScanning(true)
    setScanned([])
    setScanError('')
    setPullConflicts([])
    setPullDone(false)
    try {
      const scannedSkills = await ScanAgentSkills(agentName)
      setScanned(scannedSkills ?? [])
      setPullState(prev => ({ ...prev, selectedPaths: [] }))
    } catch (error: any) {
      setScanError(String(error?.message ?? error ?? t('toolSkills.pullFailed')))
    } finally {
      setScanning(false)
    }
  }, [t])

  const enterPullMode = useCallback(async () => {
    if (selectedTarget === 'all') return
    setSelectMode(false)
    setSelectedPaths(new Set())
    setPullMode(true)
    setPullState(createToolSkillsPullState(availableCategories[0] ?? defaultCategory))
    await scanForPull(selectedTarget)
  }, [availableCategories, scanForPull, selectedTarget])

  const exitPullMode = useCallback(() => {
    setPullMode(false)
    setPullState(createToolSkillsPullState(availableCategories[0] ?? defaultCategory))
    setScanned([])
    setScanError('')
    setPullConflicts([])
  }, [availableCategories])

  const handleStartPull = async () => {
    if (selectedTarget === 'all' || !pullReady) return
    setPulling(true)
    setScanError('')
    setPullDone(false)
    try {
      const conflicts = await PullFromAgent(selectedTarget, pullState.selectedPaths, pullState.targetCategory)
      if (conflicts && conflicts.length > 0) {
        setPullConflicts(conflicts)
        return
      }
      setPullDone(true)
      exitPullMode()
      await loadAgentView(selectedTarget)
    } catch (error: any) {
      setScanError(String(error?.message ?? error ?? t('toolSkills.pullFailed')))
    } finally {
      setPulling(false)
    }
  }

  const getNavStyle = (isActive: boolean) => isActive ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    color: 'var(--text-muted)',
    border: '1px solid transparent',
  }

  const toggleManageView = () => {
    if (agentSkillViewMode === 'manage') {
      setAgentSkillViewMode('browse')
      return
    }
    setSelectMode(false)
    setSelectedPaths(new Set())
    setPullMode(false)
    setScanned([])
    setScanError('')
    setPullConflicts([])
    setPullDone(false)
    setAgentSkillViewMode('manage')
  }

  const renderManagedGroupSections = (
    groups: ManagedSkillGroup<ManagedSkill>[],
    emptyKey: 'toolSkills.noEnabledSkills' | 'toolSkills.noDisabledSkills',
  ) => {
    if (groups.length === 0) {
      return (
        <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>
          {t(emptyKey)}
        </p>
      )
    }

    return groups.map(group => {
      const displayGroupName = !group.groupName || group.groupName === ungroupedLabel ? t('toolSkills.noGroup') : group.groupName
      const isNamedGroup = group.groupName && group.groupName !== ungroupedLabel
      return (
        <div key={`${emptyKey}:${group.groupName}`} className="space-y-3">
          <div className="flex items-center gap-3">
            <h3 className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{displayGroupName}</h3>
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('toolSkills.groupCount', { count: group.skills.length })}</span>
            <div className="flex-1" />
            {isNamedGroup && (
              <>
                <button onClick={() => void toggleGroup(group.groupName, true)} className="text-xs px-3 py-1.5 rounded-lg" style={{ background: 'var(--bg-elevated)' }}>{t('toolSkills.enableGroup')}</button>
                <button onClick={() => void toggleGroup(group.groupName, false)} className="text-xs px-3 py-1.5 rounded-lg" style={{ background: 'var(--bg-elevated)' }}>{t('toolSkills.disableGroup')}</button>
              </>
            )}
          </div>
          <div className="grid grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
            {group.skills.map(skill => (
              <ManagedAgentSkillCard
                key={`${group.groupName}:${skill.name}`}
                name={skill.name}
                groupName={!skill.groupName || skill.groupName === ungroupedLabel ? undefined : skill.groupName}
                pathCount={skill.paths.length}
                enabled={skill.enabled}
                primaryActionLabel={skill.enabled ? t('toolSkills.disableSkill') : t('toolSkills.enableSkill')}
                onPrimaryAction={() => void toggleSkill(skill.name, !skill.enabled)}
                onOpen={skill.paths[0] ? () => { void OpenPath(skill.paths[0]) } : undefined}
              />
            ))}
          </div>
        </div>
      )
    })
  }

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 p-3 flex flex-col gap-0.5 overflow-y-auto" style={{ borderRight: '1px solid var(--border-base)' }}>
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
          {t('toolSkills.toolList')}
        </div>
        <button
          key="all"
          onClick={() => void selectTarget('all')}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
          style={getNavStyle(selectedTarget === 'all')}
        >
          <Layers3 size={18} />
          <span className="truncate">{t('toolSkills.allAgents')}</span>
        </button>
        {agentNavItems.map(item => (
          <button
            key={item.key}
            onClick={() => void selectTarget(item.key)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-left transition-all duration-150"
            style={getNavStyle(selectedTarget === item.key)}
          >
            <ToolIcon name={item.label} size={20} />
            <span className="truncate">{item.label}</span>
          </button>
        ))}
        {!loading && agentNavItems.length === 0 && (
          <p className="px-3 text-xs mt-2" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noTools')}</p>
        )}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 flex flex-col gap-4" style={{ borderBottom: '1px solid var(--border-base)' }}>
          <div className="flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-2">
              {selectedTarget === 'all' ? <Layers3 size={18} /> : <ToolIcon name={selectedTarget} size={22} />}
              <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>
                {selectedTarget === 'all' ? t('toolSkills.allAgents') : selectedTarget}
              </span>
            </div>

            {isMemoryPanelAvailable(selectedTarget) && (
              <div className="flex items-center gap-1 rounded-xl p-1" style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}>
                {([
                  ['skills', t('toolSkills.panelSkills')],
                  ['memory', t('toolSkills.panelMemory')],
                ] as const).map(([panel, label]) => {
                  const selected = activePanel === panel
                  return (
                    <button
                      key={panel}
                      type="button"
                      onClick={() => void selectPanel(panel)}
                      className="rounded-lg px-3 py-1.5 text-sm transition-all duration-150"
                      style={selected ? {
                        background: 'var(--active-surface)',
                        color: 'var(--active-text)',
                        border: '1px solid var(--active-border)',
                        boxShadow: 'var(--active-shadow)',
                      } : {
                        color: 'var(--text-muted)',
                        border: '1px solid transparent',
                      }}
                    >
                      {label}
                    </button>
                  )
                })}
              </div>
            )}

            <div className="flex-1" />

            {selectedTarget === 'all' ? (
              <button
                type="button"
                onClick={openCreateGroupDialog}
                className="btn-primary px-3 py-1.5 text-sm rounded-lg"
              >
                {t('toolSkills.createGroup')}
              </button>
            ) : activePanel === 'skills' ? (
              <>
                <button
                  type="button"
                  onClick={toggleManageView}
                  disabled={scanning || pulling}
                  className={toolbarSecondaryButtonClassName}
                  style={agentSkillViewMode === 'manage' ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                  } : undefined}
                >
                  <Layers3 size={14} />
                  {agentSkillViewMode === 'manage' ? t('toolSkills.exitManageView') : t('toolSkills.enterManageView')}
                </button>
                {agentSkillViewMode === 'manage' ? null : selectMode ? (
                  <>
                    <button
                      onClick={toggleSelectAll}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
                      style={{ color: 'var(--text-muted)' }}
                    >
                      <CheckSquare size={14} />
                      {allSelected ? t('common.deselectAll') : t('common.selectAll')}
                    </button>
                    <button
                      onClick={() => void handleBatchDelete()}
                      disabled={selectedPaths.size === 0 || deleting}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40 text-white"
                      style={{ background: 'var(--color-error)' }}
                    >
                      {t('common.delete')} {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
                    </button>
                    <button
                      onClick={() => {
                        setSelectMode(false)
                        setSelectedPaths(new Set())
                      }}
                      className="px-3 py-1.5 text-sm rounded-lg"
                      style={{ color: 'var(--text-muted)' }}
                    >
                      {t('common.cancel')}
                    </button>
                  </>
                ) : pullMode ? (
                  <>
                    <label
                      className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg"
                      style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)', color: 'var(--text-secondary)' }}
                    >
                      <span>{t('toolSkills.pullTargetCategory')}</span>
                      <select
                        value={pullState.targetCategory}
                        onChange={event => setPullState(prev => ({ ...prev, targetCategory: event.target.value }))}
                        className="rounded-md px-2 py-1 text-sm outline-none"
                        style={{ background: 'var(--bg-surface)', color: 'var(--text-primary)' }}
                      >
                        {availableCategories.map(category => (
                          <option key={category} value={category}>{category}</option>
                        ))}
                      </select>
                    </label>
                    <button
                      onClick={() => setPullState(prev => ({
                        ...prev,
                        selectedPaths: toggleToolSkillsPullAllVisible(prev.selectedPaths, filteredScanned.map(skill => skill.path)),
                      }))}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg"
                      style={{ color: 'var(--text-muted)' }}
                    >
                      <CheckSquare size={14} />
                      {allPullSelected ? t('common.deselectAll') : t('common.selectAll')}
                    </button>
                    <button
                      onClick={() => setPullState(prev => ({
                        ...prev,
                        selectedPaths: toggleToolSkillsPullVisibleNotImported(prev.selectedPaths, visibleNotImportedPaths),
                      }))}
                      disabled={visibleNotImportedPaths.length === 0}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
                      style={{ color: 'var(--text-muted)' }}
                    >
                      <CheckSquare size={14} />
                      {allNotImportedSelected ? t('common.deselectAll') : t('syncPull.selectNotImported')}
                    </button>
                    <button
                      onClick={() => void handleStartPull()}
                      disabled={!pullReady || pulling}
                      className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-60"
                    >
                      {pulling ? <RefreshCw size={14} className="animate-spin" /> : <ArrowDownToLine size={14} />}
                      {t('toolSkills.manualPullStart', { count: pullState.selectedPaths.length })}
                    </button>
                    <button
                      onClick={exitPullMode}
                      className="px-3 py-1.5 text-sm rounded-lg"
                      style={{ color: 'var(--text-muted)' }}
                    >
                      {t('common.cancel')}
                    </button>
                  </>
                ) : (
                  <>
                    <button
                      onClick={() => void enterPullMode()}
                      className={toolbarSecondaryButtonClassName}
                    >
                      <ArrowDownToLine size={14} />
                      {t('toolSkills.manualPull')}
                    </button>
                    {filteredPushSkills.length > 0 && (
                      <button
                        onClick={() => setSelectMode(true)}
                        className={toolbarSecondaryButtonClassName}
                      >
                        <CheckSquare size={14} />
                        {t('toolSkills.batchDelete')}
                      </button>
                    )}
                  </>
                )}
              </>
            ) : null}

            {selectedTarget === 'codex' && activePanel === 'skills' && (
              <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('toolSkills.codexRestartHint')}</p>
            )}
          </div>

          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={searchPlaceholder}
            resultLabel={t('toolSkills.showingNResults', { count: resultCount })}
          />
        </div>

        <div className={`flex-1 min-h-0 p-6 ${contentScrollMode === 'pane' ? 'overflow-hidden' : 'overflow-y-auto space-y-6'}`}>
          {loading ? (
            <div className="flex items-center justify-center h-32 text-sm" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>
          ) : selectedTarget === 'all' ? (
            <section className="h-full min-h-[560px]">
              <div
                className="flex h-full min-h-0 overflow-hidden rounded-2xl"
                style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
              >
                <AgentSkillGroupSidebar
                  items={allGroupSidebarItems}
                  selectedKey={selectedAllGroupKey}
                  draggingSkillName={draggingAllSkillName}
                  onSelect={setSelectedAllGroupKey}
                  onCreateGroup={openCreateGroupDialog}
                  onRenameGroup={openRenameGroupDialog}
                  onDeleteGroup={openDeleteGroupDialog}
                  onDropSkill={(skillName, target) => { void handleAllSkillDrop(skillName, target) }}
                />

                <div className="min-w-0 flex-1 overflow-y-auto p-5 space-y-5">
                  <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1">
                      <h2 className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                        {selectedAllGroupItem?.label ?? t('category.all')}
                      </h2>
                      <p className="text-xs" style={{ color: 'var(--text-muted)' }}>
                        {t('toolSkills.groupDragHint')}
                      </p>
                    </div>
                    <span
                      className="rounded-full px-3 py-1 text-xs font-medium"
                      style={{ background: 'var(--bg-elevated)', color: 'var(--text-muted)', border: '1px solid var(--border-base)' }}
                    >
                      {t('toolSkills.groupCount', { count: selectedAllGroupItem?.count ?? filteredAllSkills.length })}
                    </span>
                  </div>

                  {allSkills.length === 0 ? (
                    <div className="space-y-2 rounded-2xl p-5" style={{ background: 'var(--bg-base)', border: '1px dashed var(--border-base)' }}>
                      <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('toolSkills.noManagedSkills')}</p>
                      <p className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noManagedSkillsHint')}</p>
                    </div>
                  ) : filteredAllSkills.length === 0 ? (
                    <div className="space-y-2 rounded-2xl p-5" style={{ background: 'var(--bg-base)', border: '1px dashed var(--border-base)' }}>
                      <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                        {search.trim() ? t('toolSkills.noMatch') : t('toolSkills.emptyGroupHint')}
                      </p>
                      {!search.trim() && (
                        <p className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.groupDragHint')}</p>
                      )}
                    </div>
                  ) : (
                    <div className="grid grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                      {filteredAllSkills.map(skill => (
                        <AllAgentSkillCard
                          key={skill.name}
                          name={skill.name}
                          groupName={!skill.groupName || skill.groupName === ungroupedLabel ? undefined : skill.groupName}
                          agents={skill.agents ?? []}
                          instanceCount={skill.instanceCount}
                          dragging={draggingAllSkillName === skill.name}
                          onDragStateChange={setDraggingAllSkillName}
                        />
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </section>
          ) : showingMemoryPanel ? (
            <section className="space-y-6">
              <div className="flex items-center gap-2">
                <Brain size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.memoryTitle')}</span>
                <div className="flex-1" />
                <button
                  onClick={() => void loadMemoryPreview(selectedTarget)}
                  disabled={memoryLoading}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
                  style={{ background: 'var(--bg-elevated)' }}
                >
                  <RefreshCw size={14} className={memoryLoading ? 'animate-spin' : ''} />
                  {t('toolSkills.memoryRefresh')}
                </button>
              </div>

              {memoryError ? (
                <p className="text-sm" style={{ color: 'var(--color-error)' }}>{memoryError}</p>
              ) : memoryLoading && !memoryPreview ? (
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</p>
              ) : memoryPreview ? (
                <>
                  {mainMemoryEntry && (
                    <div className="card-base p-4">
                      <div className="flex items-start gap-3">
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.memoryFile')}</span>
                          </div>
                          {memoryPreview.memoryPath
                            ? <p className="text-xs break-all" style={{ color: 'var(--text-muted)' }}>{memoryPreview.memoryPath}</p>
                            : <p className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryNotConfigured')}</p>
                          }
                        </div>
                        <button
                          onClick={() => { if (memoryPreview.memoryPath) void OpenPath(memoryPreview.memoryPath) }}
                          disabled={!memoryPreview.memoryPath}
                          className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
                          style={{ background: 'var(--bg-elevated)' }}
                        >
                          <FolderOpenDot size={14} />
                          {t('toolSkills.openFile')}
                        </button>
                      </div>
                      <div className="mt-4 rounded-lg border p-3 max-h-56 overflow-y-auto" style={{ borderColor: 'var(--border-base)', background: 'var(--bg-panel)' }}>
                        {!memoryPreview.memoryPath ? (
                          <p className="text-sm" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryNotConfigured')}</p>
                        ) : !memoryPreview.mainExists ? (
                          <p className="text-sm" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.memoryFileMissing')}</p>
                        ) : (
                          <pre className="whitespace-pre-wrap break-words text-sm m-0" style={{ color: 'var(--text-primary)' }}>{mainMemoryEntry.content || t('toolSkills.emptyContent')}</pre>
                        )}
                      </div>
                    </div>
                  )}

                  <div>
                    <div className="flex items-center gap-2 mb-4">
                      <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.rulesDir')}</span>
                      {memoryPreview.rulesDir
                        ? <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={memoryPreview.rulesDir}>{memoryPreview.rulesDir}</span>
                        : <span className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDir')}</span>
                      }
                      <div className="flex-1" />
                      <button
                        onClick={() => { if (memoryPreview.rulesDir) void OpenPath(memoryPreview.rulesDir) }}
                        disabled={!memoryPreview.rulesDir}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
                        style={{ background: 'var(--bg-elevated)' }}
                      >
                        <FolderOpenDot size={14} />
                        {t('toolSkills.openDir')}
                      </button>
                    </div>

                    {!memoryPreview.rulesDir ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.rulesNotConfigured')}</p>
                    ) : !memoryPreview.rulesDirExists ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.rulesDirMissing')}</p>
                    ) : ruleEntries.length === 0 ? (
                      <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>
                        {search.trim() ? t('toolSkills.noMemoryMatch') : t('toolSkills.noRuleFiles')}
                      </p>
                    ) : (
                      <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
                        {ruleEntries.map(entry => (
                          <div key={entry.key} className="card-base p-4">
                            <div className="flex items-start gap-3">
                              <div className="min-w-0 flex-1">
                                <div className="flex items-center gap-2 mb-1">
                                  <span className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>{entry.title}</span>
                                  {entry.managed && (
                                    <span className="px-2 py-0.5 rounded-full text-[11px] font-medium" style={{ background: 'var(--accent-glow)', color: 'var(--accent-primary)', border: '1px solid var(--border-accent)' }}>
                                      {t('toolSkills.managedRule')}
                                    </span>
                                  )}
                                </div>
                                <p className="text-xs break-all" style={{ color: 'var(--text-muted)' }}>{entry.path}</p>
                              </div>
                              <button
                                onClick={() => void OpenPath(entry.path)}
                                className="p-1 rounded"
                                style={{ color: 'var(--text-muted)' }}
                                title={t('toolSkills.openFile')}
                              >
                                <FolderOpenDot size={13} />
                              </button>
                            </div>
                            <div className="mt-3 rounded-lg border p-3 max-h-40 overflow-y-auto" style={{ borderColor: 'var(--border-base)', background: 'var(--bg-panel)' }}>
                              <pre className="whitespace-pre-wrap break-words text-sm m-0" style={{ color: 'var(--text-primary)' }}>{entry.content || t('toolSkills.emptyContent')}</pre>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </>
              ) : null}
            </section>
          ) : showingPullMode ? (
            <section className="space-y-4">
              <div className="flex items-center gap-2">
                <ArrowDownToLine size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.manualPullTitle')}</span>
              </div>

              {scanning ? (
                <p className="text-sm pl-5" style={{ color: 'var(--text-muted)' }}>{t('syncPull.scanning')}</p>
              ) : scanError ? (
                <div className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm" style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}>
                  <AlertCircle size={16} className="mt-0.5 shrink-0" />
                  <span className="flex-1">{scanError}</span>
                </div>
              ) : scanned.length === 0 ? (
                <div className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm" style={{ background: 'rgba(251,191,36,0.1)', border: '1px solid rgba(251,191,36,0.3)', color: 'var(--color-warning)' }}>
                  <AlertCircle size={16} className="shrink-0" />
                  <span>{t('syncPull.emptyWarning')}</span>
                </div>
              ) : filteredScanned.length === 0 ? (
                <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('syncPull.noMatch')}</p>
              ) : (
                <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                  {filteredScanned.map(sk => (
                    <SyncSkillCard
                      key={sk.path}
                      name={sk.name}
                      source={sk.source}
                      path={sk.path}
                      imported={sk.imported}
                      updatable={sk.updatable}
                      pushedAgents={(sk.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedTarget)}
                      showImported={visibility.includes('imported')}
                      showUpdatable={visibility.includes('updatable')}
                      showPushedAgents={visibility.includes('pushedAgents')}
                      selected={pullState.selectedPaths.includes(sk.path)}
                      onToggle={() => setPullState(prev => ({
                        ...prev,
                        selectedPaths: toggleToolSkillsPullPath(prev.selectedPaths, sk.path),
                      }))}
                    />
                  ))}
                </div>
              )}
            </section>
          ) : showingManageView ? (
            <>
              {pullDone && (
                <div className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm" style={{ background: 'rgba(52,211,153,0.1)', border: '1px solid rgba(52,211,153,0.3)', color: 'var(--color-success)' }}>
                  <Check size={16} className="shrink-0" />
                  <span>{t('toolSkills.pullDone')}</span>
                </div>
              )}

              <section className="space-y-8">
                {filteredManagedSkills.length === 0 ? (
                  <div className="space-y-2 rounded-2xl p-5" style={{ background: 'var(--bg-base)', border: '1px dashed var(--border-base)' }}>
                    <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                      {search.trim() ? t('toolSkills.noMatch') : t('toolSkills.noManagedSkills')}
                    </p>
                    {!search.trim() && (
                      <p className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noManagedSkillsHint')}</p>
                    )}
                  </div>
                ) : (
                  <>
                    <section className="space-y-4">
                      <div className="flex items-center gap-2">
                        <Check size={14} style={{ color: 'var(--color-success)' }} className="shrink-0" />
                        <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.enabledSectionTitle')}</span>
                      </div>
                      {renderManagedGroupSections(groupedEnabledManagedSkills, 'toolSkills.noEnabledSkills')}
                    </section>

                    <section className="space-y-4">
                      <div className="flex items-center gap-2">
                        <AlertCircle size={14} style={{ color: 'var(--text-muted)' }} className="shrink-0" />
                        <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.disabledSectionTitle')}</span>
                      </div>
                      {renderManagedGroupSections(groupedDisabledManagedSkills, 'toolSkills.noDisabledSkills')}
                    </section>
                  </>
                )}
              </section>
            </>
          ) : showingBrowseView ? (
            <>
              {pullDone && (
                <div className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm" style={{ background: 'rgba(52,211,153,0.1)', border: '1px solid rgba(52,211,153,0.3)', color: 'var(--color-success)' }}>
                  <Check size={16} className="shrink-0" />
                  <span>{t('toolSkills.pullDone')}</span>
                </div>
              )}

              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ArrowUpToLine size={14} style={{ color: 'var(--color-success)' }} className="shrink-0" />
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.pushPath')}</span>
                  {currentAgent?.pushDir
                    ? <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={currentAgent.pushDir}>{currentAgent.pushDir}</span>
                    : <span className="text-xs" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDir')}</span>
                  }
                </div>
                {!currentAgent?.pushDir ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushDirDesc')}</p>
                ) : pushSkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noPushSkills')}</p>
                ) : filteredPushSkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noMatch')}</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {filteredPushSkills.map(skill => (
                      <ToolSkillCard
                        key={skill.path}
                        name={skill.name}
                        path={skill.path}
                        source={skill.source}
                        imported={skill.imported}
                        updatable={skill.updatable}
                        pushedAgents={(skill.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedTarget)}
                        showImported={visibility.includes('imported')}
                        showUpdatable={visibility.includes('updatable')}
                        showPushedAgents={visibility.includes('pushedAgents')}
                        canDelete
                        selectMode={selectMode}
                        selected={selectedPaths.has(skill.path)}
                        onToggleSelect={() => toggleSelectPath(skill.path)}
                        onDelete={() => void handleDelete(skill.path)}
                      />
                    ))}
                  </div>
                )}
              </section>

              <section>
                <div className="flex items-center gap-2 mb-4">
                  <ScanLine size={14} style={{ color: 'var(--accent-primary)' }} className="shrink-0" />
                  <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{t('toolSkills.scanPath')}</span>
                  {currentAgent?.scanDirs?.length > 0 && (
                    <span className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={currentAgent.scanDirs.join(', ')}>
                      {t('toolSkills.nDirs', { count: currentAgent.scanDirs.length })}
                    </span>
                  )}
                </div>
                <p className="mb-4 pl-5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('toolSkills.scanPathHint')}</p>
                {scanOnlySkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noScanSkills')}</p>
                ) : filteredScanOnlySkills.length === 0 ? (
                  <p className="text-sm pl-5" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.noMatch')}</p>
                ) : (
                  <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                    {filteredScanOnlySkills.map(skill => (
                      <ToolSkillCard
                        key={skill.path}
                        name={skill.name}
                        path={skill.path}
                        source={skill.source}
                        imported={skill.imported}
                        updatable={skill.updatable}
                        pushedAgents={(skill.pushedAgents ?? []).filter((agentName: string) => agentName !== selectedTarget)}
                        showImported={visibility.includes('imported')}
                        showUpdatable={visibility.includes('updatable')}
                        showPushedAgents={visibility.includes('pushedAgents')}
                        canDelete={false}
                        selectMode={false}
                        selected={false}
                        onToggleSelect={() => {}}
                        onDelete={() => {}}
                      />
                    ))}
                  </div>
                )}
              </section>
            </>
          ) : null}
        </div>
      </div>

      <ConflictDialog
        conflicts={pullConflicts}
        labelForConflict={(path) => scanned.find((item: any) => item.path === path)?.name ?? path}
        onOverwrite={async (path) => {
          await PullFromAgentForce(selectedTarget, [path], pullState.targetCategory)
          setPullConflicts(prev => prev.filter(conflict => conflict !== path))
        }}
        onSkip={(path) => setPullConflicts(prev => prev.filter(conflict => conflict !== path))}
        onDone={() => {
          setPullDone(true)
          exitPullMode()
          void loadAgentView(selectedTarget)
        }}
      />

      <AnimatedDialog
        open={groupDialogMode !== null}
        onClose={groupDialogSaving ? undefined : closeGroupDialog}
        width="w-[420px]"
        zIndex={65}
      >
        <div className="flex items-center gap-2 mb-3">
          <Layers3 size={18} style={{ color: 'var(--accent-primary)' }} />
          <span className="text-base font-semibold" style={{ color: 'var(--text-primary)' }}>
            {groupDialogMode === 'create'
              ? t('toolSkills.groupDialogCreateTitle')
              : t('toolSkills.groupDialogRenameTitle')}
          </span>
        </div>
        <p className="text-sm leading-6" style={{ color: 'var(--text-secondary)' }}>
          {groupDialogMode === 'create'
            ? t('toolSkills.groupDialogCreateDesc')
            : t('toolSkills.groupDialogRenameDesc')}
        </p>
        <label className="mt-5 block text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
          {t('toolSkills.groupNameLabel')}
        </label>
        <input
          autoFocus
          value={groupDraftName}
          onChange={event => {
            setGroupDraftName(event.target.value)
            if (groupDialogError) setGroupDialogError('')
          }}
          onKeyDown={event => {
            if (event.key === 'Enter') {
              event.preventDefault()
              void submitGroupDialog()
            }
          }}
          placeholder={t('toolSkills.groupNamePlaceholder')}
          disabled={groupDialogSaving}
          className="mt-2 w-full rounded-xl px-3 py-2.5 text-sm outline-none"
          style={{ background: 'var(--bg-base)', color: 'var(--text-primary)', border: '1px solid var(--border-base)' }}
        />
        {normalizedGroupDraftName === '' ? (
          <p className="mt-3 text-sm" style={{ color: 'var(--color-warning)' }}>{t('toolSkills.groupNameRequired')}</p>
        ) : isDuplicateGroupName ? (
          <p className="mt-3 text-sm" style={{ color: 'var(--color-warning)' }}>{t('toolSkills.groupNameDuplicate')}</p>
        ) : groupDialogError ? (
          <p className="mt-3 text-sm" style={{ color: 'var(--color-error)' }}>{groupDialogError}</p>
        ) : null}
        <div className="mt-6 flex items-center justify-end gap-3">
          <button onClick={closeGroupDialog} disabled={groupDialogSaving} className="btn-secondary">
            {t('common.cancel')}
          </button>
          <button
            onClick={() => void submitGroupDialog()}
            disabled={groupDialogSaving || normalizedGroupDraftName === '' || isDuplicateGroupName}
            className="btn-primary disabled:opacity-40"
          >
            {groupDialogSaving ? t('common.saving') : t('common.save')}
          </button>
        </div>
      </AnimatedDialog>

      <AnimatedDialog
        open={groupDeleteTarget !== null}
        onClose={() => {
          setGroupDeleteTarget(null)
          setGroupDialogError('')
        }}
        width="w-[420px]"
        zIndex={65}
      >
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle size={18} style={{ color: 'var(--color-error)' }} />
          <span className="text-base font-semibold" style={{ color: 'var(--text-primary)' }}>
            {t('toolSkills.groupDialogDeleteTitle')}
          </span>
        </div>
        <p className="text-sm leading-6" style={{ color: 'var(--text-secondary)' }}>
          {t('toolSkills.groupDialogDeleteDesc')}
        </p>
        {groupDeleteTarget && (
          <div
            className="mt-4 rounded-xl px-3 py-3 text-sm"
            style={{ background: 'var(--bg-base)', color: 'var(--text-primary)', border: '1px solid var(--border-base)' }}
          >
            {groupDeleteTarget}
          </div>
        )}
        {groupDialogError && (
          <p className="mt-4 text-sm" style={{ color: 'var(--color-error)' }}>{groupDialogError}</p>
        )}
        <div className="mt-6 flex items-center justify-end gap-3">
          <button
            onClick={() => {
              setGroupDeleteTarget(null)
              setGroupDialogError('')
            }}
            className="btn-secondary"
          >
            {t('common.cancel')}
          </button>
          <button onClick={() => void confirmDeleteGroup()} className="btn-primary" style={{ background: 'var(--color-error)' }}>
            {t('common.confirm')}
          </button>
        </div>
      </AnimatedDialog>
    </div>
  )
}

interface ToolSkillCardProps {
  name: string
  path: string
  source?: string
  imported?: boolean
  updatable?: boolean
  pushedAgents?: string[]
  showImported: boolean
  showUpdatable: boolean
  showPushedAgents: boolean
  canDelete: boolean
  selectMode: boolean
  selected: boolean
  onToggleSelect: () => void
  onDelete: () => void
}

function ToolSkillCard({
  name,
  path,
  source,
  imported,
  updatable,
  pushedAgents = [],
  showImported,
  showUpdatable,
  showPushedAgents,
  canDelete,
  selectMode,
  selected,
  onToggleSelect,
  onDelete,
}: ToolSkillCardProps) {
  const { t } = useLanguage()
  const cardRef = useRef<HTMLDivElement>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [hoveredRect, setHoveredRect] = useState<DOMRect | null>(null)
  const [meta, setMeta] = useState<any | null>(null)
  const [copied, setCopied] = useState(false)

  const sourceLabel = source === 'github'
    ? t('common.sourceGitHub')
    : source === 'manual'
      ? t('common.sourceManual')
      : source === 'git'
        ? t('common.sourceGit')
        : source

  const handleMouseEnter = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      if (!cardRef.current) return
      setHoveredRect(cardRef.current.getBoundingClientRect())
      setMeta(null)
      try {
        const nextMeta = await GetSkillMetaByPath(path)
        setMeta(nextMeta ?? {})
      } catch {
        setMeta({})
      }
    }, 300)
  }

  const handleMouseLeave = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredRect(null)
    setMeta(null)
  }

  const handleOpen = (event: ReactMouseEvent<HTMLButtonElement>) => {
    event.stopPropagation()
    void OpenPath(path)
  }

  const handleCopy = async (event: ReactMouseEvent<HTMLButtonElement>) => {
    event.stopPropagation()
    try {
      const content = await ReadSkillFileContent(path)
      await copyTextToClipboard(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // ignore clipboard failures on hover actions
    }
  }

  const handleClick = () => {
    if (selectMode && canDelete) onToggleSelect()
  }

  return (
    <>
      <div
        ref={cardRef}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`card-base relative p-4 group ${selectMode && canDelete ? 'cursor-pointer' : 'cursor-default'}`}
        style={selected ? {
          background: 'var(--accent-glow)',
          borderColor: 'var(--border-accent)',
          boxShadow: 'var(--glow-accent-sm)',
        } : undefined}
      >
        {selectMode && canDelete && (
          <div className="absolute top-2 left-2 z-10">
            <div
              className="w-4 h-4 rounded border-2 flex items-center justify-center transition-all duration-150"
              style={selected ? {
                background: 'var(--accent-secondary)',
                borderColor: 'var(--accent-secondary)',
                boxShadow: 'var(--glow-accent-sm)',
              } : {
                borderColor: 'var(--text-muted)',
                background: 'var(--bg-elevated)',
              }}
            >
              {selected && (
                <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              )}
            </div>
          </div>
        )}

        {!selectMode && (
          <div className="absolute top-2 right-2 flex items-center gap-0.5 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={(event) => void handleCopy(event)}
              title={t('toolSkills.copySkill')}
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
            >
              {copied ? <Check size={13} style={{ color: 'var(--color-success)' }} /> : <Copy size={13} />}
            </button>
            <button
              onClick={handleOpen}
              title={t('toolSkills.openDir')}
              className="p-1 rounded transition-colors"
              style={{ color: 'var(--text-muted)' }}
            >
              <FolderOpenDot size={13} />
            </button>
          </div>
        )}

        <SkillStatusStrip
          className={`${selectMode && canDelete ? 'pl-5' : ''} pr-12`}
          badges={[
            ...(sourceLabel ? [{
              key: `source:${sourceLabel}`,
              label: sourceLabel,
              tone: source === 'github' ? ('accent' as const) : ('muted' as const),
            }] : []),
            ...(showImported && imported ? [{
              key: 'imported',
              label: t('common.imported'),
              tone: 'success' as const,
            }] : []),
            ...(showUpdatable && updatable ? [{
              key: 'updatable',
              label: t('common.updatable'),
              tone: 'warning' as const,
            }] : []),
          ]}
          pushedAgents={showPushedAgents ? pushedAgents : []}
        />

        <p
          className={`mt-1 min-h-[2.75rem] font-medium text-sm leading-snug line-clamp-2 ${selectMode && canDelete ? 'pl-5' : ''} ${!selectMode ? 'pr-5' : ''}`}
          style={{ color: 'var(--text-primary)' }}
        >
          {name}
        </p>

        {!selectMode && canDelete && (
          <div className="mt-3 flex opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={(event) => {
                event.stopPropagation()
                onDelete()
              }}
              className="text-xs ml-auto transition-colors"
              style={{ color: 'var(--color-error)' }}
            >
              {t('toolSkills.delete')}
            </button>
          </div>
        )}

        {!canDelete && (
          <div className="mt-3 flex">
            <span className="text-xs ml-auto" style={{ color: 'var(--text-disabled)' }}>{t('toolSkills.readOnly')}</span>
          </div>
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={{ Name: name, Source: source }} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
