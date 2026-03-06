import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ListStarredRepos, AddStarredRepo, RemoveStarredRepo,
  UpdateStarredRepo, UpdateAllStarredRepos,
  ListAllStarSkills, ListRepoStarSkills,
  ImportStarSkills, ListCategories, OpenURL,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  Star, RefreshCw, Plus, Trash2, LayoutGrid, Folder,
  ChevronLeft, CheckSquare, Download, AlertCircle, X, ExternalLink,
} from 'lucide-react'
import SyncSkillCard from '../components/SyncSkillCard'

export default function StarredRepos() {
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
    const off1 = EventsOn('star.sync.progress', () => loadRepos())
    const off2 = EventsOn('star.sync.done', () => { loadRepos(); loadAllSkills(); setSyncing(false) })
    return () => { off1?.(); off2?.() }
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
      setAddError(String(e?.message ?? e ?? '添加失败'))
    } finally { setAdding(false) }
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
      const skills = currentRepo ? repoSkills : allSkills
      // group by repoUrl for multi-repo flat import
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

  const skills = currentRepo ? repoSkills : allSkills

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-3 px-6 py-4 border-b border-gray-800 flex-wrap">
        {currentRepo ? (
          <button onClick={() => { navigate('/starred'); setSelectMode(false); setSelectedPaths(new Set()) }}
            className="flex items-center gap-1 text-sm text-gray-400 hover:text-white">
            <ChevronLeft size={14} />
            <span>{currentRepo.split('/').slice(-2).join('/')}</span>
          </button>
        ) : (
          <h2 className="text-sm font-medium flex items-center gap-2">
            <Star size={14} /> 仓库收藏
          </h2>
        )}
        <div className="flex-1" />
        {selectMode ? (
          <>
            <button onClick={() => toggleSelectAll(skills)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <CheckSquare size={14} />{selectedPaths.size === skills.length ? '取消全选' : '全选'}
            </button>
            <button onClick={() => setShowImportDialog(true)} disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 rounded-lg">
              <Download size={14} /> 导入 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">取消</button>
          </>
        ) : (
          <>
            {!currentRepo && (
              <>
                <button onClick={() => setView('folder')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg ${view === 'folder' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:bg-gray-800'}`}>
                  <Folder size={14} /> 文件夹
                </button>
                <button onClick={() => setView('flat')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg ${view === 'flat' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:bg-gray-800'}`}>
                  <LayoutGrid size={14} /> 平铺
                </button>
              </>
            )}
            <button onClick={() => setSelectMode(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <CheckSquare size={14} /> 批量导入
            </button>
            <button onClick={handleUpdateAll} disabled={syncing}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <RefreshCw size={14} className={syncing ? 'animate-spin' : ''} /> 全部更新
            </button>
            {!currentRepo && (
              <button onClick={() => setShowAdd(true)}
                className="flex items-center gap-1.5 px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 rounded-lg">
                <Plus size={14} /> 添加仓库
              </button>
            )}
          </>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {currentRepo ? (
          <SkillGrid skills={repoSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        ) : view === 'folder' ? (
          <RepoGrid repos={repos}
            onEnter={url => navigate(`/starred/${encodeURIComponent(url)}`)}
            onUpdate={handleUpdateOne}
            onRemove={handleRemove} />
        ) : (
          <SkillGrid skills={allSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        )}
      </div>

      {/* Add repo dialog */}
      {showAdd && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[460px] border border-gray-700">
            <div className="flex justify-between items-center mb-4">
              <h3 className="font-semibold flex items-center gap-2"><Star size={16} /> 添加远程仓库</h3>
              <button onClick={() => { setShowAdd(false); setAddError('') }}><X size={16} className="text-gray-400" /></button>
            </div>
            <div className="flex gap-2 mb-3">
              <input value={addUrl} onChange={e => setAddUrl(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !adding && addUrl && handleAddRepo()}
                placeholder="https://host/owner/repo.git 或 git@host:owner/repo.git"
                className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
              <button onClick={handleAddRepo} disabled={adding || !addUrl}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50 min-w-[72px]">
                {adding ? '克隆中...' : '添加'}
              </button>
            </div>
            <p className="text-xs text-gray-500 mb-3">首次添加会 git clone 仓库，可能需要一些时间</p>
            {addError && (
              <div className="flex items-start gap-2 bg-red-950 border border-red-700 text-red-300 rounded-lg px-4 py-3 text-sm">
                <AlertCircle size={15} className="mt-0.5 shrink-0" />
                <span>{addError}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Import category dialog */}
      {showImportDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[380px] border border-gray-700">
            <h3 className="font-semibold mb-4">选择导入分类</h3>
            <select value={importCategory} onChange={e => setImportCategory(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm mb-4">
              {categories.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
            <div className="flex gap-3">
              <button onClick={handleBatchImport} disabled={importing}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
                {importing ? '导入中...' : `导入 ${selectedPaths.size} 个`}
              </button>
              <button onClick={() => setShowImportDialog(false)}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function RepoGrid({ repos, onEnter, onUpdate, onRemove }: {
  repos: any[]
  onEnter: (url: string) => void
  onUpdate: (url: string) => void
  onRemove: (url: string) => void
}) {
  if (repos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 text-gray-500">
        <Star size={32} className="mb-2 opacity-30" />
        <p className="text-sm">还没有收藏的仓库</p>
        <p className="text-xs mt-1">点击「添加仓库」开始收藏</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
      {repos.map((r: any) => (
        <div key={r.url} onClick={() => onEnter(r.url)}
          className="bg-gray-800 rounded-xl p-4 border border-gray-700 hover:border-indigo-500 cursor-pointer transition-colors">
          <div className="flex justify-between items-start mb-2">
            <span className="font-medium text-sm truncate flex-1 mr-2">{r.name}</span>
            <div className="flex gap-1 shrink-0" onClick={e => e.stopPropagation()}>
              <button onClick={() => OpenURL(r.url)}
                className="p-1 text-gray-400 hover:text-indigo-400 rounded" title="在浏览器中打开">
                <ExternalLink size={12} />
              </button>
              <button onClick={() => onUpdate(r.url)}
                className="p-1 text-gray-400 hover:text-white rounded" title="更新">
                <RefreshCw size={12} />
              </button>
              <button onClick={() => onRemove(r.url)}
                className="p-1 text-gray-400 hover:text-red-400 rounded" title="删除收藏">
                <Trash2 size={12} />
              </button>
            </div>
          </div>
          {r.syncError ? (
            <p className="text-xs text-red-400 truncate" title={r.syncError}>{r.syncError}</p>
          ) : (
            <>
              <p className="text-xs text-gray-500 truncate" title={r.source || r.url}>{r.source || r.url}</p>
              <p className="text-xs text-gray-500 mt-1">
                {r.lastSync && r.lastSync !== '0001-01-01T00:00:00Z'
                  ? `同步于 ${new Date(r.lastSync).toLocaleDateString()}`
                  : '未同步'}
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
  if (skills.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 text-gray-500">
        <p className="text-sm">没有找到 Skills</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
      {skills.map((sk: any) => (
        <SyncSkillCard
          key={sk.path}
          name={sk.name}
          path={sk.path}
          source={sk.source || undefined}
          subtitle={showRepo ? sk.repoName : undefined}
          imported={sk.imported}
          showSelection={selectMode}
          selected={selectedPaths.has(sk.path)}
          onToggle={() => selectMode && onToggle(sk.path)}
        />
      ))}
    </div>
  )
}
