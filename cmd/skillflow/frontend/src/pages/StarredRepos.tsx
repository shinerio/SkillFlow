import { useEffect, useMemo, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import {
  ListStarredRepos, AddStarredRepo, AddStarredRepoWithCredentials, RemoveStarredRepo,
  UpdateStarredRepo, UpdateAllStarredRepos,
  ListAllStarSkills, ListRepoStarSkills,
  ImportStarSkills, ListCategories, OpenURL, GetConfig,
  GetEnabledTools, PushStarSkillsToTools, PushStarSkillsToToolsForce, CheckMissingPushDirs,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  Star, RefreshCw, Plus, Trash2, LayoutGrid, Folder,
  ChevronLeft, CheckSquare, Download, AlertCircle, X, ExternalLink, ArrowUpToLine, Lock, KeyRound, FolderPlus, CheckCircle,
} from 'lucide-react'
import SyncSkillCard from '../components/SyncSkillCard'
import { ToolIcon } from '../config/toolIcons'
import AnimatedDialog from '../components/ui/AnimatedDialog'
import SkillListControls from '../components/SkillListControls'
import { toastVariants } from '../lib/motionVariants'
import { useLanguage } from '../contexts/LanguageContext'
import { useSkillStatusVisibility } from '../contexts/SkillStatusVisibilityContext'
import { SkillSortOrder, filterAndSortSkills } from '../lib/skillList'

export default function StarredRepos() {
  const { t } = useLanguage()
  const { repoEncoded } = useParams()
  const navigate = useNavigate()
  const currentRepo = repoEncoded ? decodeURIComponent(repoEncoded) : null

  const [repos, setRepos] = useState<any[]>([])
  const [repoSkills, setRepoSkills] = useState<any[]>([])
  const [allSkills, setAllSkills] = useState<any[]>([])
  const [view, setView] = useState<'folder' | 'flat'>('folder')
  const [syncing, setSyncing] = useState(false)
  const [addUrl, setAddUrl] = useState('')
  const [showAdd, setShowAdd] = useState(false)
  const [adding, setAdding] = useState(false)
  const [addError, setAddError] = useState('')
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [importCategory, setImportCategory] = useState('')
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [importing, setImporting] = useState(false)
  const [tools, setTools] = useState<any[]>([])
  const [showPushToolDialog, setShowPushToolDialog] = useState(false)
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [pushingToTools, setPushingToTools] = useState(false)
  const [pushConflicts, setPushConflicts] = useState<any[]>([])
  const [showPushConflictDialog, setShowPushConflictDialog] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{name: string, dir: string}[]>([])
  const [showMkdirDialog, setShowMkdirDialog] = useState(false)
  const [pushSuccessMsg, setPushSuccessMsg] = useState('')
  // Auth dialogs
  const [showHttpAuthDialog, setShowHttpAuthDialog] = useState(false)
  const [showSshErrorDialog, setShowSshErrorDialog] = useState(false)
  const [authUrl, setAuthUrl] = useState('')
  const [authUsername, setAuthUsername] = useState('')
  const [authPassword, setAuthPassword] = useState('')
  const [authError, setAuthError] = useState('')
  const [authAdding, setAuthAdding] = useState(false)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<SkillSortOrder>('asc')
  const [hasGitHubPAT, setHasGitHubPAT] = useState(true)
  const [captureMsg, setCaptureMsg] = useState('')

  const loadRepos = async () => {
    const r = await ListStarredRepos()
    setRepos(r ?? [])
  }

  const loadAllSkills = async () => {
    const s = await ListAllStarSkills()
    setAllSkills(s ?? [])
  }

  const loadRepoSkills = async (url: string) => {
    const s = await ListRepoStarSkills(url)
    setRepoSkills(s ?? [])
  }

  useEffect(() => {
    loadRepos()
    loadAllSkills()
    ListCategories().then(c => {
      setCategories(c ?? [])
      if (c && c.length > 0) setImportCategory(c[0])
    })
    GetConfig().then(c => setHasGitHubPAT(Boolean(c?.github?.pat?.trim?.() || c?.github?.pat)))
    GetEnabledTools().then(t => setTools(t ?? []))
    const off1 = EventsOn('star.sync.progress', () => loadRepos())
    const off2 = EventsOn('star.sync.done', () => { loadRepos(); loadAllSkills(); setSyncing(false) })
    const off3 = EventsOn('github.star.captured', (payload: string) => {
      const repoName = String(payload || '')
      if (repoName) setCaptureMsg(`🌟 ${t('starred.captured')}: ${repoName}`)
      loadRepos(); loadAllSkills()
    })
    return () => { off1?.(); off2?.(); off3?.() }
  }, [])

  useEffect(() => {
    if (currentRepo) loadRepoSkills(currentRepo)
  }, [currentRepo])

  const handleAddRepo = async () => {
    setAdding(true); setAddError('')
    try {
      await AddStarredRepo(addUrl)
      setShowAdd(false); setAddUrl('')
      await Promise.all([loadRepos(), loadAllSkills()])
    } catch (e: any) {
      const msg = String(e?.message ?? e ?? t('starred.addFailed'))
      if (msg.startsWith('AUTH_SSH:')) {
        setShowAdd(false)
        setShowSshErrorDialog(true)
      } else if (msg.startsWith('AUTH_HTTP:')) {
        setAuthUrl(addUrl)
        setAuthUsername(''); setAuthPassword(''); setAuthError('')
        setShowHttpAuthDialog(true)
      } else {
        setAddError(msg)
      }
    } finally { setAdding(false) }
  }

  const handleAuthRetry = async () => {
    setAuthAdding(true); setAuthError('')
    try {
      await AddStarredRepoWithCredentials(authUrl, authUsername, authPassword)
      setShowHttpAuthDialog(false)
      setShowAdd(false); setAddUrl('')
      await Promise.all([loadRepos(), loadAllSkills()])
    } catch (e: any) {
      setAuthError(String(e?.message ?? e ?? t('starred.authFailed')))
    } finally { setAuthAdding(false) }
  }

  const handleUpdateAll = async () => {
    setSyncing(true)
    try {
      await UpdateAllStarredRepos()
    } finally {
      setSyncing(false)
      await Promise.all([loadRepos(), loadAllSkills()])
    }
  }

  const handleUpdateOne = async (url: string) => {
    await UpdateStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const handleRemove = async (url: string) => {
    await RemoveStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const toggleSelectPath = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev)
      next.has(path) ? next.delete(path) : next.add(path)
      return next
    })
  }

  const toggleSelectAll = (skills: any[]) => {
    if (selectedPaths.size === skills.length) setSelectedPaths(new Set())
    else setSelectedPaths(new Set(skills.map((s: any) => s.path)))
  }

  const handleBatchImport = async () => {
    setImporting(true)
    try {
      const byRepo = new Map<string, string[]>()
      for (const path of selectedPaths) {
        const sk = skills.find((s: any) => s.path === path)
        if (!sk) continue
        const arr = byRepo.get(sk.repoUrl) ?? []
        arr.push(path)
        byRepo.set(sk.repoUrl, arr)
      }
      for (const [rURL, paths] of byRepo) {
        await ImportStarSkills(paths, rURL, importCategory)
      }
      setShowImportDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      if (currentRepo) loadRepoSkills(currentRepo); else loadAllSkills()
    } catch (e: any) {
      console.error('Import failed:', e)
    } finally { setImporting(false) }
  }

  const doPushToTools = async () => {
    setPushingToTools(true)
    try {
      const paths = [...selectedPaths]
      const toolNames = [...selectedTools]
      const conflicts = await PushStarSkillsToTools(paths, toolNames)
      setShowPushToolDialog(false)
      if (conflicts && conflicts.length > 0) {
        setPushConflicts(conflicts)
        setShowPushConflictDialog(true)
      } else {
        setSelectMode(false)
        setSelectedPaths(new Set())
        const count = paths.length
        const toolCount = toolNames.length
        setPushSuccessMsg(t('starred.successMsg', { count, toolCount }))
        setTimeout(() => setPushSuccessMsg(''), 3000)
      }
    } catch (e: any) {
      console.error('Push to tools failed:', e)
    } finally {
      setPushingToTools(false)
    }
  }

  const handlePushToTools = async () => {
    const toolNames = [...selectedTools]
    const missing = await CheckMissingPushDirs(toolNames)
    if (missing && missing.length > 0) {
      setMissingDirs(missing as {name: string, dir: string}[])
      setShowMkdirDialog(true)
    } else {
      await doPushToTools()
    }
  }

  const confirmMkdirAndPush = async () => {
    setShowMkdirDialog(false)
    setMissingDirs([])
    await doPushToTools()
  }

  const handlePushToToolsForce = async () => {
    try {
      for (const conflict of pushConflicts) {
        await PushStarSkillsToToolsForce([conflict.skillPath], [conflict.toolName])
      }
      setShowPushConflictDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      setPushConflicts([])
      setPushSuccessMsg(t('starred.successMsg', { count: selectedPaths.size, toolCount: selectedTools.size }))
      setTimeout(() => setPushSuccessMsg(''), 3000)
    } catch (e: any) {
      console.error('Force push failed:', e)
    }
  }

  const toggleTool = (name: string) => {
    setSelectedTools(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  const filteredRepoSkills = useMemo(
    () => filterAndSortSkills(repoSkills, search, sortOrder, skill => skill.name ?? ''),
    [repoSkills, search, sortOrder],
  )

  const filteredAllSkills = useMemo(
    () => filterAndSortSkills(allSkills, search, sortOrder, skill => skill.name ?? ''),
    [allSkills, search, sortOrder],
  )

  const skillGridVisible = !!currentRepo || view === 'flat'
  const skills = currentRepo ? filteredRepoSkills : filteredAllSkills

  const toolBtnStyle = (active: boolean) => active ? {
    background: 'var(--accent-glow)',
    color: 'var(--accent-primary)',
    border: '1px solid var(--border-accent)',
    boxShadow: 'var(--glow-accent-sm)',
  } : {
    background: 'var(--bg-elevated)',
    color: 'var(--text-secondary)',
    border: '1px solid var(--border-base)',
  }

  useEffect(() => {
    if (!selectMode || !skillGridVisible) return
    const visiblePaths = new Set(skills.map((skill: any) => skill.path))
    setSelectedPaths(prev => {
      const next = new Set([...prev].filter(path => visiblePaths.has(path)))
      return next.size === prev.size ? prev : next
    })
  }, [selectMode, skillGridVisible, skills])

  useEffect(() => {
    if (!skillGridVisible && selectMode) {
      setSelectMode(false)
      setSelectedPaths(new Set())
    }
  }, [skillGridVisible, selectMode])

  useEffect(() => {
    if (!captureMsg) return
    const timer = window.setTimeout(() => setCaptureMsg(''), 4000)
    return () => window.clearTimeout(timer)
  }, [captureMsg])

  return (
    <div className="flex flex-col h-full">
      {/* Success toast */}
      <AnimatePresence>
        {pushSuccessMsg && (
          <motion.div
            variants={toastVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            className="fixed top-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm shadow-dialog"
            style={{
              background: 'rgba(52,211,153,0.15)',
              border: '1px solid rgba(52,211,153,0.4)',
              color: 'var(--color-success)',
              backdropFilter: 'blur(8px)',
            }}
          >
            <CheckCircle size={15} className="shrink-0" />
            {pushSuccessMsg}
          </motion.div>
        )}
      </AnimatePresence>
      <AnimatePresence>
        {captureMsg && (
          <motion.div
            variants={toastVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            className="fixed top-16 left-1/2 -translate-x-1/2 z-50 flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm shadow-dialog"
            style={{
              background: 'rgba(251,191,36,0.15)',
              border: '1px solid rgba(251,191,36,0.4)',
              color: 'var(--color-warning)',
              backdropFilter: 'blur(8px)',
            }}
          >
            <Star size={15} className="shrink-0" />
            {captureMsg}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Toolbar */}
      <div className="flex items-center gap-3 px-6 py-4 flex-wrap" style={{ borderBottom: '1px solid var(--border-base)' }}>
        {currentRepo ? (
          <button
            onClick={() => { navigate('/starred'); setSelectMode(false); setSelectedPaths(new Set()) }}
            className="flex items-center gap-1 text-sm transition-colors"
            style={{ color: 'var(--text-muted)' }}
            onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
            onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
          >
            <ChevronLeft size={14} />
            <span>{currentRepo.split('/').slice(-2).join('/')}</span>
          </button>
        ) : (
          <h2 className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Star size={14} /> {t('starred.title')}
          </h2>
        )}
        {skillGridVisible ? (
          <SkillListControls
            search={search}
            onSearchChange={setSearch}
            sortOrder={sortOrder}
            onSortOrderChange={setSortOrder}
            placeholder={currentRepo ? t('starred.searchCurrentRepo') : t('starred.searchAllRepos')}
            resultLabel={t('common.showingNSkills', { count: skills.length })}
            searchClassName="max-w-[420px]"
          />
        ) : (
          <div className="flex-1" />
        )}
        {selectMode ? (
          <>
            <button
              onClick={() => toggleSelectAll(skills)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <CheckSquare size={14} />{selectedPaths.size === skills.length ? t('common.deselectAll') : t('common.selectAll')}
            </button>
            <button
              onClick={() => { setSelectedTools(new Set()); setShowPushToolDialog(true) }}
              disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40 transition-colors"
              style={{ background: 'rgba(52,211,153,0.2)', color: 'var(--color-success)', border: '1px solid rgba(52,211,153,0.4)' }}
            >
              <ArrowUpToLine size={14} /> {t('starred.pushToToolsCount')} {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button
              onClick={() => setShowImportDialog(true)}
              disabled={selectedPaths.size === 0}
              className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg disabled:opacity-40"
            >
              <Download size={14} /> {t('starred.importToMySkills')} {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button
              onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >{t('common.cancel')}</button>
          </>
        ) : (
          <>
            {!currentRepo && (
              <>
                {[['folder', t('starred.folder'), <Folder size={14} />], ['flat', t('starred.flat'), <LayoutGrid size={14} />]].map(([v, label, icon]) => (
                  <button
                    key={v as string}
                    onClick={() => setView(v as 'folder' | 'flat')}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-all duration-150"
                    style={view === v ? {
                      background: 'var(--bg-elevated)',
                      color: 'var(--text-primary)',
                      border: '1px solid var(--border-surface)',
                    } : {
                      color: 'var(--text-muted)',
                      border: '1px solid transparent',
                    }}
                  >
                    {icon} {label}
                  </button>
                ))}
              </>
            )}
            {skillGridVisible && (
              <button
                onClick={() => setSelectMode(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <CheckSquare size={14} /> {t('starred.batchImport')}
              </button>
            )}
            <button
              onClick={handleUpdateAll}
              disabled={syncing}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => { e.currentTarget.style.backgroundColor = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; e.currentTarget.style.color = 'var(--text-muted)' }}
            >
              <RefreshCw size={14} className={syncing ? 'animate-spin' : ''} /> {t('starred.updateAll')}
            </button>
            {!currentRepo && (
              <button
                onClick={() => setShowAdd(true)}
                className="btn-primary flex items-center gap-1.5 px-4 py-1.5 text-sm rounded-lg"
              >
                <Plus size={14} /> {t('starred.addRepo')}
              </button>
            )}
          </>
        )}
      </div>

      {!hasGitHubPAT && !currentRepo && (
        <div className="mx-6 mt-3 mb-0 px-3 py-2 rounded-lg text-xs" style={{ background: 'rgba(251,191,36,0.12)', border: '1px solid rgba(251,191,36,0.4)', color: 'var(--text-secondary)' }}>
          {t('starred.githubPatRequired')}
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {currentRepo ? (
          <SkillGrid skills={filteredRepoSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        ) : view === 'folder' ? (
          <RepoGrid repos={repos}
            onEnter={url => navigate(`/starred/${encodeURIComponent(url)}`)}
            onUpdate={handleUpdateOne}
            onRemove={handleRemove} />
        ) : (
          <SkillGrid skills={filteredAllSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        )}
      </div>

      {/* Add repo dialog */}
      <AnimatedDialog open={showAdd} onClose={() => { setShowAdd(false); setAddError('') }} width="w-[460px]">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Star size={16} /> {t('starred.addRepoTitle')}
          </h3>
          <button onClick={() => { setShowAdd(false); setAddError('') }} style={{ color: 'var(--text-muted)' }}>
            <X size={16} />
          </button>
        </div>
        <div className="flex gap-2 mb-3">
          <input
            value={addUrl}
            onChange={e => setAddUrl(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !adding && addUrl && handleAddRepo()}
            placeholder={t('starred.urlPlaceholder')}
            className="input-base flex-1"
          />
          <button
            onClick={handleAddRepo}
            disabled={adding || !addUrl}
            className="btn-primary px-4 py-2 rounded-lg text-sm min-w-[72px]"
          >
            {adding ? t('starred.cloning') : t('starred.addBtn')}
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>{t('starred.addHint')}</p>
        {addError && (
          <div
            className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm"
            style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
          >
            <AlertCircle size={15} className="mt-0.5 shrink-0" />
            <span>{addError}</span>
          </div>
        )}
      </AnimatedDialog>

      {/* Import to My Skills dialog */}
      <AnimatedDialog open={showImportDialog} onClose={() => setShowImportDialog(false)} width="w-[380px]">
        <h3 className="font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>{t('starred.selectCategory')}</h3>
        <select
          value={importCategory}
          onChange={e => setImportCategory(e.target.value)}
          className="select-base mb-4"
        >
          {categories.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        <div className="flex gap-3">
          <button
            onClick={handleBatchImport}
            disabled={importing}
            className="btn-primary flex-1 py-2 rounded-lg text-sm"
          >
            {importing ? t('starred.importing') : t('starred.importN', { count: selectedPaths.size })}
          </button>
          <button onClick={() => setShowImportDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">{t('common.cancel')}</button>
        </div>
      </AnimatedDialog>

      {/* Mkdir confirmation dialog */}
      <AnimatedDialog open={showMkdirDialog} onClose={() => { setShowMkdirDialog(false); setMissingDirs([]) }} width="w-[460px]" zIndex={60}>
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <FolderPlus size={16} /> {t('starred.mkdirTitle')}
          </h3>
          <button onClick={() => { setShowMkdirDialog(false); setMissingDirs([]) }} style={{ color: 'var(--text-muted)' }}>
            <X size={16} />
          </button>
        </div>
        <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>{t('starred.mkdirDesc')}</p>
        <ul className="space-y-1.5 mb-4 max-h-40 overflow-y-auto">
          {missingDirs.map(d => (
            <li key={d.name} className="text-sm rounded-lg px-3 py-2" style={{ background: 'var(--bg-surface)' }}>
              <span className="font-medium" style={{ color: 'var(--text-primary)' }}>{d.name}</span>
              <span className="text-xs block truncate" style={{ color: 'var(--text-muted)' }} title={d.dir}>{d.dir}</span>
            </li>
          ))}
        </ul>
        <div className="flex gap-3">
          <button onClick={confirmMkdirAndPush} className="btn-primary flex-1 py-2 rounded-lg text-sm">{t('starred.createAndPush')}</button>
          <button onClick={() => { setShowMkdirDialog(false); setMissingDirs([]) }} className="btn-secondary flex-1 py-2 rounded-lg text-sm">{t('common.cancel')}</button>
        </div>
      </AnimatedDialog>

      {/* Push to tool dialog */}
      <AnimatedDialog open={showPushToolDialog} onClose={() => setShowPushToolDialog(false)} width="w-[420px]">
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <ArrowUpToLine size={16} /> {t('starred.pushToTools')}
          </h3>
          <button onClick={() => setShowPushToolDialog(false)} style={{ color: 'var(--text-muted)' }}><X size={16} /></button>
        </div>
        <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>{t('starred.pushDialogDesc')}</p>
        {tools.length === 0 ? (
          <p className="text-sm py-4 text-center" style={{ color: 'var(--text-muted)' }}>{t('starred.noTools')}</p>
        ) : (
          <div className="flex flex-wrap gap-2 mb-4">
            {tools.map((t: any) => (
              <button
                key={t.name}
                onClick={() => toggleTool(t.name)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200"
                style={toolBtnStyle(selectedTools.has(t.name))}
              >
                <ToolIcon name={t.name} size={16} />
                {t.name}
              </button>
            ))}
          </div>
        )}
        <div className="flex gap-3">
          <button
            onClick={handlePushToTools}
            disabled={pushingToTools || selectedTools.size === 0 || tools.length === 0}
            className="btn-primary flex-1 py-2 rounded-lg text-sm"
          >
            {pushingToTools ? t('starred.pushingToTools') : t('starred.pushToNTools', { count: selectedTools.size })}
          </button>
          <button onClick={() => setShowPushToolDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">{t('common.cancel')}</button>
        </div>
      </AnimatedDialog>

      {/* HTTP auth dialog */}
      <AnimatedDialog open={showHttpAuthDialog} onClose={() => setShowHttpAuthDialog(false)} width="w-[460px]">
        <div className="flex justify-between items-center mb-1">
          <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
            <Lock size={16} /> {t('starred.authTitle')}
          </h3>
          <button onClick={() => setShowHttpAuthDialog(false)} style={{ color: 'var(--text-muted)' }}><X size={16} /></button>
        </div>
        <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>{t('starred.authDesc')}</p>
        <div className="space-y-2 mb-4">
          <input
            value={authUsername}
            onChange={e => setAuthUsername(e.target.value)}
            placeholder={t('starred.username')}
            className="input-base"
          />
          <input
            type="password"
            value={authPassword}
            onChange={e => setAuthPassword(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !authAdding && handleAuthRetry()}
            placeholder={t('starred.password')}
            className="input-base"
          />
        </div>
        {authError && (
          <div
            className="flex items-start gap-2 rounded-lg px-4 py-3 text-sm mb-3"
            style={{ background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--color-error)' }}
          >
            <AlertCircle size={15} className="mt-0.5 shrink-0" />
            <span>{authError}</span>
          </div>
        )}
        <div className="flex gap-3">
          <button onClick={handleAuthRetry} disabled={authAdding} className="btn-primary flex-1 py-2 rounded-lg text-sm">
            {authAdding ? t('starred.connecting') : t('common.confirm')}
          </button>
          <button onClick={() => setShowHttpAuthDialog(false)} className="btn-secondary flex-1 py-2 rounded-lg text-sm">{t('common.cancel')}</button>
        </div>
      </AnimatedDialog>

      {/* SSH auth error dialog */}
      <AnimatedDialog open={showSshErrorDialog} onClose={() => setShowSshErrorDialog(false)} width="w-[460px]">
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <KeyRound size={16} /> {t('starred.sshTitle')}
        </h3>
        <p className="text-sm mb-3" style={{ color: 'var(--text-secondary)' }}>{t('starred.sshDesc')}</p>
        <ul className="text-sm space-y-1.5 list-disc list-inside mb-4" style={{ color: 'var(--text-muted)' }}>
          <li>{t('starred.sshCheckKeygen')}（<code style={{ color: 'var(--text-secondary)' }}>ssh-keygen</code>）</li>
          <li>{t('starred.sshCheckPubkey')}</li>
          <li>{t('starred.sshCheckAgent')}（<code style={{ color: 'var(--text-secondary)' }}>ssh-add</code>）</li>
          <li>{t('starred.sshCheckHttps')}</li>
        </ul>
        <button onClick={() => setShowSshErrorDialog(false)} className="btn-secondary w-full py-2 rounded-lg text-sm">{t('common.close')}</button>
      </AnimatedDialog>

      {/* Push conflict dialog */}
      <AnimatedDialog open={showPushConflictDialog} onClose={() => setShowPushConflictDialog(false)} width="w-[420px]">
        <h3 className="font-semibold mb-2 flex items-center gap-2" style={{ color: 'var(--color-warning)' }}>
          <AlertCircle size={16} /> {t('starred.conflictsTitle')}
        </h3>
        <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>{t('starred.conflictsDesc')}</p>
        <ul className="space-y-1 mb-4 max-h-40 overflow-y-auto">
          {pushConflicts.map(c => (
            <li
              key={`${c.skillPath}|${c.toolName}`}
              className="text-sm px-3 py-1.5 rounded"
              style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)' }}
            >
              {c.skillName} → {c.toolName}
            </li>
          ))}
        </ul>
        <div className="flex gap-3">
          <button
            onClick={handlePushToToolsForce}
            className="flex-1 py-2 rounded-lg text-sm text-white transition-colors"
            style={{ background: 'var(--color-warning)' }}
          >{t('starred.overwriteAll')}</button>
          <button
            onClick={() => { setShowPushConflictDialog(false); setSelectMode(false); setSelectedPaths(new Set()); setPushConflicts([]) }}
            className="btn-secondary flex-1 py-2 rounded-lg text-sm"
          >{t('starred.skipConflicts')}</button>
        </div>
      </AnimatedDialog>
    </div>
  )
}

function RepoGrid({ repos, onEnter, onUpdate, onRemove }: {
  repos: any[]
  onEnter: (url: string) => void
  onUpdate: (url: string) => void
  onRemove: (url: string) => void
}) {
  const { t } = useLanguage()
  if (repos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
        <Star size={32} className="mb-2 opacity-30" />
        <p className="text-sm">{t('starred.emptyTitle')}</p>
        <p className="text-xs mt-1">{t('starred.emptyHint')}</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
      {repos.map((r: any) => (
        <div
          key={r.url}
          onClick={() => onEnter(r.url)}
          className="card-base p-4 cursor-pointer"
        >
          <div className="flex justify-between items-start mb-2">
            <div className="flex-1 mr-2 min-w-0">
              <span className="font-medium text-sm truncate block" style={{ color: 'var(--text-primary)' }}>{r.name}</span>
              <div className="mt-1 flex gap-1">
                {r.manual && <span className="text-[10px] px-1.5 py-0.5 rounded" style={{ background: 'var(--bg-elevated)', color: 'var(--text-muted)' }}>{t('starred.badgeManual')}</span>}
                {r.starred && <span className="text-[10px] px-1.5 py-0.5 rounded" style={{ background: 'rgba(251,191,36,0.18)', color: 'var(--color-warning)' }}>⭐ {t('starred.badgeStarred')}</span>}
              </div>
            </div>
            <div className="flex gap-1 shrink-0" onClick={e => e.stopPropagation()}>
              <button
                onClick={() => OpenURL(r.source ? `https://${r.source}` : r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--accent-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title={t('starred.openInBrowser')}
              >
                <ExternalLink size={12} />
              </button>
              <button
                onClick={() => onUpdate(r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--text-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title={t('starred.updateBtn')}
              >
                <RefreshCw size={12} />
              </button>
              <button
                onClick={() => onRemove(r.url)}
                className="p-1 rounded transition-colors"
                style={{ color: 'var(--text-muted)' }}
                onMouseEnter={e => { e.currentTarget.style.color = 'var(--color-error)' }}
                onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)' }}
                title={t('starred.removeStarred')}
              >
                <Trash2 size={12} />
              </button>
            </div>
          </div>
          {r.syncError ? (
            <p className="text-xs truncate" style={{ color: 'var(--color-error)' }} title={r.syncError}>{r.syncError}</p>
          ) : (
            <>
              <p className="text-xs truncate" style={{ color: 'var(--text-muted)' }} title={r.source || r.url}>{r.source || r.url}</p>
              <p className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                {r.lastSync && r.lastSync !== '0001-01-01T00:00:00Z'
                  ? `${t('starred.syncAt')} ${new Date(r.lastSync).toLocaleString()}`
                  : t('starred.notSynced')}
              </p>
            </>
          )}
        </div>
      ))}
    </div>
  )
}

function SkillGrid({ skills, selectMode, selectedPaths, onToggle, showRepo = false }: {
  skills: any[]
  selectMode: boolean
  selectedPaths: Set<string>
  onToggle: (path: string) => void
  showRepo?: boolean
}) {
  const { t } = useLanguage()
  const visibility = useSkillStatusVisibility('starredRepos')
  if (skills.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48" style={{ color: 'var(--text-muted)' }}>
        <p className="text-sm">{t('starred.noSkills')}</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
      {skills.map((sk: any) => {
        const src = sk.source || sk.repoUrl || ''
        const sourceType = src.includes('github.com') ? 'github' : src ? 'git' : undefined
        return (
          <SyncSkillCard
            key={sk.path}
            name={sk.name}
            path={sk.path}
            source={sourceType}
            subtitle={showRepo ? sk.repoName : undefined}
            imported={sk.imported}
            pushedTools={sk.pushedTools}
            showImported={visibility.includes('imported')}
            showPushedTools={visibility.includes('pushedTools')}
            showSelection={selectMode}
            selected={selectedPaths.has(sk.path)}
            onToggle={() => selectMode && onToggle(sk.path)}
          />
        )
      })}
    </div>
  )
}
