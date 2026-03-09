import { useEffect, useRef, useState, useCallback, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  ListSkills, ListCategories, MoveSkillCategory,
  DeleteSkill, DeleteSkills, ImportLocal, UpdateSkill, CheckUpdates,
  OpenFolderDialog, GetSkillMeta, GetConfig, SaveConfig,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'
import SkillTooltip from '../components/SkillTooltip'
import GitHubInstallDialog from '../components/GitHubInstallDialog'
import { Github, FolderOpen, RefreshCw, Trash2, CheckSquare, ArrowUpFromLine } from 'lucide-react'
import { gridContainerVariants, cardVariants } from '../lib/motionVariants'
import SkillListControls from '../components/SkillListControls'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { ToolIcon } from '../config/toolIcons'

export default function Dashboard() {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('mySkills')
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [showGitHub, setShowGitHub] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const [draggingSkillID, setDraggingSkillID] = useState<string | null>(null)
  const [categoryDragActive, setCategoryDragActive] = useState(false)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedIDs, setSelectedIDs] = useState<Set<string>>(new Set())
  const [toolOptions, setToolOptions] = useState<any[]>([])
  const [autoPushTools, setAutoPushTools] = useState<Set<string>>(new Set())
  const [dashboardCfg, setDashboardCfg] = useState<any | null>(null)
  const [savingAutoPush, setSavingAutoPush] = useState(false)

  // Hover tooltip state
  const [hoveredSkill, setHoveredSkill] = useState<{ skill: any; rect: DOMRect } | null>(null)
  const [hoveredMeta, setHoveredMeta] = useState<any | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const load = useCallback(async () => {
    const [s, c, cfg] = await Promise.all([ListSkills(), ListCategories(), GetConfig()])
    setSkills(s ?? [])
    setCategories(c ?? [])
    setDashboardCfg(cfg)
    setToolOptions((cfg?.tools ?? []).filter((tool: any) => tool.enabled))
    setAutoPushTools(new Set(cfg?.autoPushTools ?? []))
  }, [])

  useEffect(() => {
    load()
    EventsOn('update.available', load)
  }, [load])

  const filtered = useMemo(
    () => filterAndSortSkills(
      skills.filter(sk => selectedCat === null || sk.category === selectedCat),
      search,
      sortOrder,
      sk => sk.name ?? '',
    ),
    [skills, selectedCat, search, sortOrder],
  )

  const skillCounts = skills.reduce((acc, sk) => {
    const category = sk.category || 'Default'
    acc[category] = (acc[category] ?? 0) + 1
    return acc
  }, {} as Record<string, number>)

  const handleDrop = async (skillId: string, category: string) => {
    await MoveSkillCategory(skillId, category)
    setDraggingSkillID(null)
    setCategoryDragActive(false)
    load()
  }

  const isFileDrag = (e: React.DragEvent) => e.dataTransfer.types.includes('Files')

  const handleWindowDragOver = (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    e.preventDefault()
    setDragOver(true)
  }
  const handleWindowDragLeave = (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    setDragOver(false)
  }
  const handleWindowDrop = async (e: React.DragEvent) => {
    if (!isFileDrag(e)) return
    e.preventDefault()
    setDragOver(false)
    const items = Array.from(e.dataTransfer.items)
    for (const item of items) {
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        const file = item.getAsFile()
        if (file) {
          // @ts-ignore — Wails provides .path on File objects
          await ImportLocal(file.path ?? file.name, selectedCat ?? '')
          load()
        }
      }
    }
  }

  const handleImportButton = async () => {
    const dir = await OpenFolderDialog('')
    if (dir) { await ImportLocal(dir, selectedCat ?? ''); load() }
  }

  const toggleSelectMode = () => {
    setSelectMode(prev => !prev)
    setSelectedIDs(new Set())
    clearHover()
  }

  const toggleSelectID = (id: string) => {
    setSelectedIDs(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleSelectAll = () => {
    const filteredIDs = filtered.map(sk => sk.id)
    setSelectedIDs(prev => {
      const next = new Set(prev)
      if (filteredIDs.every(id => next.has(id))) {
        filteredIDs.forEach(id => next.delete(id))
      } else {
        filteredIDs.forEach(id => next.add(id))
      }
      return next
    })
  }

  const handleBatchDelete = async () => {
    if (selectedIDs.size === 0) return
    await DeleteSkills(Array.from(selectedIDs))
    setSelectedIDs(new Set())
    setSelectMode(false)
    load()
  }

  useEffect(() => {
    if (!selectMode) return
    const visibleIDs = new Set(filtered.map(sk => sk.id))
    setSelectedIDs(prev => {
      const next = new Set([...prev].filter(id => visibleIDs.has(id)))
      return next.size === prev.size ? prev : next
    })
  }, [filtered, selectMode])

  const allSelected = filtered.length > 0 && filtered.every(sk => selectedIDs.has(sk.id))

  const toggleAutoPushTool = async (name: string) => {
    if (!dashboardCfg || savingAutoPush) return

    const nextSet = new Set(autoPushTools)
    if (nextSet.has(name)) nextSet.delete(name)
    else nextSet.add(name)

    const nextCfg = {
      ...dashboardCfg,
      autoPushTools: Array.from(nextSet),
    }

    setAutoPushTools(new Set(nextSet))
    setDashboardCfg(nextCfg)
    setSavingAutoPush(true)

    try {
      await SaveConfig(nextCfg)
      await load()
    } catch (error) {
      console.error('Save auto push tools failed:', error)
      const latestCfg = await GetConfig()
      setDashboardCfg(latestCfg)
      setToolOptions((latestCfg?.tools ?? []).filter((tool: any) => tool.enabled))
      setAutoPushTools(new Set(latestCfg?.autoPushTools ?? []))
    } finally {
      setSavingAutoPush(false)
    }
  }

  const clearHover = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredSkill(null)
    setHoveredMeta(null)
  }

  const handleHoverStart = (sk: any, rect: DOMRect) => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      setHoveredSkill({
        skill: {
          Name: sk.name,
          Category: sk.category,
          Source: sk.source,
          SourceSHA: sk.sourceSha,
          LatestSHA: sk.latestSha,
        },
        rect,
      })
      setHoveredMeta(null)
      const meta = await GetSkillMeta(sk.id)
      setHoveredMeta(meta)
    }, 300)
  }

  const handleHoverEnd = () => {
    clearHover()
  }

  const containerVariants = gridContainerVariants(filtered.length)

  return (
    <div
      className={`flex h-full relative ${dragOver ? 'ring-2 ring-inset' : ''}`}
      style={dragOver ? { '--tw-ring-color': 'var(--accent-primary)' } as any : {}}
      onDragOver={handleWindowDragOver}
      onDragLeave={handleWindowDragLeave}
      onDrop={handleWindowDrop}
    >
      {dragOver && (
        <div className="absolute inset-0 flex items-center justify-center z-40 pointer-events-none"
          style={{ background: 'var(--active-surface)', backdropFilter: 'blur(6px)' }}>
          <p className="text-lg font-medium" style={{ color: 'var(--active-text)' }}>{t('dashboard.dropToImport')}</p>
        </div>
      )}

      <CategoryPanel
        categories={categories}
        skillCounts={skillCounts}
        selected={selectedCat}
        draggingSkillId={draggingSkillID}
        onSelect={setSelectedCat}
        onCategoryDragStateChange={setCategoryDragActive}
        onDrop={handleDrop}
        onRefresh={load}
      />

      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div
          className="flex flex-wrap items-center gap-3 px-6 py-4"
          style={{ borderBottom: '1px solid var(--border-base)' }}
        >
          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={t('dashboard.searchPlaceholder')}
            resultLabel={t('common.showingNSkills', { count: filtered.length })}
          />

          {selectMode ? (
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              <button
                onClick={toggleSelectAll}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                <CheckSquare size={14} />
                {allSelected ? t('common.deselectAll') : t('common.selectAll')}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedIDs.size === 0}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                style={{ backgroundColor: 'var(--color-error)', color: 'white' }}
              >
                <Trash2 size={14} /> {t('common.delete')} {selectedIDs.size > 0 ? `(${selectedIDs.size})` : ''}
              </button>
              <button
                onClick={toggleSelectMode}
                className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
              >
                {t('common.cancel')}
              </button>
            </div>
          ) : (
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              {[
                { icon: <RefreshCw size={14} />, label: t('dashboard.update'), onClick: async () => { await CheckUpdates(); load() } },
                { icon: <CheckSquare size={14} />, label: t('dashboard.batchDelete'), onClick: toggleSelectMode },
                { icon: <FolderOpen size={14} />, label: t('dashboard.import'), onClick: handleImportButton },
              ].map(btn => (
                <button
                  key={btn.label}
                  onClick={btn.onClick}
                  className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
                  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
                >
                  {btn.icon} {btn.label}
                </button>
              ))}
              <button
                onClick={() => setShowGitHub(true)}
                className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg whitespace-nowrap"
              >
                <Github size={14} /> {t('dashboard.remoteInstall')}
              </button>
            </div>
          )}
        </div>

        <div
          className="px-6 py-3 flex flex-wrap items-center gap-3"
          style={{
            borderBottom: '1px solid var(--border-base)',
            background: 'linear-gradient(135deg, color-mix(in srgb, var(--accent-glow) 42%, transparent) 0%, color-mix(in srgb, var(--bg-elevated) 94%, transparent) 100%)',
          }}
        >
          <div className="flex items-center gap-2 shrink-0 text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
            <ArrowUpFromLine size={15} />
            {t('dashboard.autoPushTitle')}
          </div>
          {toolOptions.length > 0 ? (
            <div className="flex flex-wrap items-center gap-2 min-w-0 flex-1">
              {toolOptions.map(tool => {
                const active = autoPushTools.has(tool.name)
                return (
                  <button
                    key={tool.name}
                    onClick={() => void toggleAutoPushTool(tool.name)}
                    disabled={savingAutoPush}
                    className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed ${active ? 'font-semibold -translate-y-px' : ''}`}
                    style={active ? {
                      background: 'var(--active-surface)',
                      color: 'var(--active-text)',
                      border: '1px solid var(--active-border)',
                      boxShadow: 'var(--active-shadow)',
                    } : {
                      background: 'var(--bg-elevated)',
                      color: 'var(--text-secondary)',
                      border: '1px solid var(--border-base)',
                    }}
                  >
                    <ToolIcon name={tool.name} size={20} />
                    <span>{tool.name}</span>
                  </button>
                )
              })}
              {savingAutoPush && (
                <span className="text-xs whitespace-nowrap ml-1" style={{ color: 'var(--text-muted)' }}>
                  {t('common.saving')}
                </span>
              )}
            </div>
          ) : (
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
              {t('dashboard.autoPushEmpty')}
            </p>
          )}
        </div>

        {/* Skills grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <motion.div
            className="grid grid-cols-3 xl:grid-cols-4 gap-4"
            variants={containerVariants}
            initial="initial"
            animate="animate"
          >
            {filtered.map(sk => (
              <motion.div key={sk.id} variants={filtered.length <= 30 ? cardVariants : undefined}>
                <SkillCard
                  skill={{
                    id: sk.id,
                    name: sk.name,
                    category: sk.category,
                    source: sk.source,
                    hasUpdate: !!sk.updatable,
                    path: sk.path,
                    pushedTools: sk.pushedTools,
                  }}
                  showUpdatable={visibility.includes('updatable')}
                  showPushedTools={visibility.includes('pushedTools')}
                  categories={categories}
                  onDelete={async () => { await DeleteSkill(sk.id); load() }}
                  onUpdate={async () => { await UpdateSkill(sk.id); load() }}
                  onMoveCategory={async cat => { await MoveSkillCategory(sk.id, cat); load() }}
                  dragging={draggingSkillID === sk.id}
                  dropTargetActive={draggingSkillID === sk.id && categoryDragActive}
                  onDragStateChange={(dragging) => {
                    setDraggingSkillID(dragging ? sk.id : null)
                    if (!dragging) setCategoryDragActive(false)
                  }}
                  selectMode={selectMode}
                  selected={selectedIDs.has(sk.id)}
                  onToggleSelect={() => toggleSelectID(sk.id)}
                  onHoverStart={rect => handleHoverStart(sk, rect)}
                  onHoverEnd={handleHoverEnd}
                />
              </motion.div>
            ))}
          </motion.div>
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
              <p className="text-sm">{t('dashboard.empty')}</p>
              <p className="text-xs mt-1">{t('dashboard.emptyHint')}</p>
            </div>
          )}
        </div>
      </div>

      {hoveredSkill && (
        <SkillTooltip
          skill={hoveredSkill.skill}
          meta={hoveredMeta}
          anchorRect={hoveredSkill.rect}
        />
      )}

      <AnimatePresence>
        {showGitHub && (
          <GitHubInstallDialog onClose={() => setShowGitHub(false)} onDone={() => { setShowGitHub(false); load() }} />
        )}
      </AnimatePresence>
    </div>
  )
}
