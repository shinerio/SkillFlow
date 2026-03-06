import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ListStarredRepos, AddStarredRepo, AddStarredRepoWithCredentials, RemoveStarredRepo,
  UpdateStarredRepo, UpdateAllStarredRepos,
  ListAllStarSkills, ListRepoStarSkills,
  ImportStarSkills, ListCategories, OpenURL,
  GetEnabledTools, PushStarSkillsToTools, PushStarSkillsToToolsForce,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  Star, RefreshCw, Plus, Trash2, LayoutGrid, Folder,
  ChevronLeft, CheckSquare, Download, AlertCircle, X, ExternalLink, ArrowUpToLine, Lock, KeyRound,
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
  const [tools, setTools] = useState<any[]>([])
  const [showPushToolDialog, setShowPushToolDialog] = useState(false)
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [pushingToTools, setPushingToTools] = useState(false)
  const [pushConflicts, setPushConflicts] = useState<string[]>([])
  const [showPushConflictDialog, setShowPushConflictDialog] = useState(false)
  // Auth dialogs
  const [showHttpAuthDialog, setShowHttpAuthDialog] = useState(false)
  const [showSshErrorDialog, setShowSshErrorDialog] = useState(false)
  const [authUrl, setAuthUrl] = useState('')
  const [authUsername, setAuthUsername] = useState('')
  const [authPassword, setAuthPassword] = useState('')
  const [authError, setAuthError] = useState('')
  const [authAdding, setAuthAdding] = useState(false)

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
    GetEnabledTools().then(t => setTools(t ?? []))
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
      const msg = String(e?.message ?? e ?? '添加失败')
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
      setAuthError(String(e?.message ?? e ?? '认证失败'))
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

  const handlePushToTools = async () => {
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
      }
    } catch (e: any) {
      console.error('Push to tools failed:', e)
    } finally {
      setPushingToTools(false)
    }
  }

  const handlePushToToolsForce = async () => {
    try {
      await PushStarSkillsToToolsForce([...selectedPaths], [...selectedTools])
      setShowPushConflictDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      setPushConflicts([])
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
            <button onClick={() => { setSelectedTools(new Set()); setShowPushToolDialog(true) }} disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white disabled:opacity-40 rounded-lg hover:bg-gray-800">
              <ArrowUpToLine size={14} /> 推送到工具 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button onClick={() => setShowImportDialog(true)} disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 rounded-lg">
              <Download size={14} /> 导入到我的Skills {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
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

      {/* Import to My Skills dialog */}
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

      {/* Push to tool dialog */}
      {showPushToolDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[420px] border border-gray-700">
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-semibold flex items-center gap-2"><ArrowUpToLine size={16} /> 推送到工具</h3>
              <button onClick={() => setShowPushToolDialog(false)}><X size={16} className="text-gray-400" /></button>
            </div>
            <p className="text-xs text-gray-500 mb-4">将 Skill 直接复制到工具目录，无需导入到「我的Skills」</p>
            {tools.length === 0 ? (
              <p className="text-sm text-gray-500 py-4 text-center">没有可用的工具，请在设置中启用工具</p>
            ) : (
              <div className="space-y-1 mb-4 max-h-52 overflow-y-auto">
                {tools.map((t: any) => (
                  <label key={t.Name} className="flex items-center gap-3 p-2.5 rounded-lg hover:bg-gray-700 cursor-pointer">
                    <input type="checkbox" className="accent-indigo-500 shrink-0"
                      checked={selectedTools.has(t.Name)}
                      onChange={() => toggleTool(t.Name)} />
                    <span className="text-sm font-medium">{t.Name}</span>
                    {t.PushDir && (
                      <span className="text-xs text-gray-500 truncate flex-1 text-right" title={t.PushDir}>{t.PushDir}</span>
                    )}
                  </label>
                ))}
              </div>
            )}
            <div className="flex gap-3">
              <button onClick={handlePushToTools} disabled={pushingToTools || selectedTools.size === 0 || tools.length === 0}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
                {pushingToTools ? '推送中...' : `推送到 ${selectedTools.size} 个工具`}
              </button>
              <button onClick={() => setShowPushToolDialog(false)}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">取消</button>
            </div>
          </div>
        </div>
      )}

      {/* HTTP auth dialog */}
      {showHttpAuthDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[460px] border border-gray-700">
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-semibold flex items-center gap-2"><Lock size={16} /> 需要认证</h3>
              <button onClick={() => setShowHttpAuthDialog(false)}><X size={16} className="text-gray-400" /></button>
            </div>
            <p className="text-xs text-gray-500 mb-4">仓库需要用户名和密码（或 Access Token）才能访问</p>
            <div className="space-y-2 mb-4">
              <input
                value={authUsername}
                onChange={e => setAuthUsername(e.target.value)}
                placeholder="用户名"
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500"
              />
              <input
                type="password"
                value={authPassword}
                onChange={e => setAuthPassword(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !authAdding && handleAuthRetry()}
                placeholder="密码 / Access Token"
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500"
              />
            </div>
            {authError && (
              <div className="flex items-start gap-2 bg-red-950 border border-red-700 text-red-300 rounded-lg px-4 py-3 text-sm mb-3">
                <AlertCircle size={15} className="mt-0.5 shrink-0" />
                <span>{authError}</span>
              </div>
            )}
            <div className="flex gap-3">
              <button onClick={handleAuthRetry} disabled={authAdding}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
                {authAdding ? '连接中...' : '确认'}
              </button>
              <button onClick={() => setShowHttpAuthDialog(false)}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">取消</button>
            </div>
          </div>
        </div>
      )}

      {/* SSH auth error dialog */}
      {showSshErrorDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[460px] border border-gray-700">
            <h3 className="font-semibold mb-2 flex items-center gap-2 text-amber-400">
              <KeyRound size={16} /> SSH 认证失败
            </h3>
            <p className="text-sm text-gray-300 mb-3">无法使用 SSH 访问远程仓库，请检查以下配置：</p>
            <ul className="text-sm text-gray-400 space-y-1.5 list-disc list-inside mb-4">
              <li>SSH 密钥是否已生成（<code className="text-gray-300">ssh-keygen</code>）</li>
              <li>公钥是否已添加到 GitHub / GitLab 等远程仓库</li>
              <li>SSH Agent 是否正在运行（<code className="text-gray-300">ssh-add</code>）</li>
              <li>可尝试改用 HTTPS 协议克隆</li>
            </ul>
            <button onClick={() => setShowSshErrorDialog(false)}
              className="w-full py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">关闭</button>
          </div>
        </div>
      )}

      {/* Push conflict dialog */}
      {showPushConflictDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[420px] border border-gray-700">
            <h3 className="font-semibold mb-2 flex items-center gap-2 text-amber-400">
              <AlertCircle size={16} /> 发现冲突
            </h3>
            <p className="text-sm text-gray-400 mb-3">以下 Skill 在目标工具目录中已存在：</p>
            <ul className="space-y-1 mb-4 max-h-40 overflow-y-auto">
              {pushConflicts.map(c => (
                <li key={c} className="text-sm text-gray-300 bg-gray-900 px-3 py-1.5 rounded">{c}</li>
              ))}
            </ul>
            <div className="flex gap-3">
              <button onClick={handlePushToToolsForce}
                className="flex-1 py-2 bg-amber-600 hover:bg-amber-500 rounded-lg text-sm">覆盖全部</button>
              <button onClick={() => { setShowPushConflictDialog(false); setSelectMode(false); setSelectedPaths(new Set()); setPushConflicts([]) }}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">跳过冲突</button>
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
              <button onClick={() => OpenURL(r.source ? `https://${r.source}` : r.url)}
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
          showSelection={selectMode}
          selected={selectedPaths.has(sk.path)}
          onToggle={() => selectMode && onToggle(sk.path)}
        />
        )
      })}
    </div>
  )
}
