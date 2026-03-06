import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star, X, Download, RefreshCw, AlertTriangle, GitMerge, MessageSquareWarning } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import StarredRepos from './pages/StarredRepos'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DownloadAppUpdate, ApplyAppUpdate, GetGitConflictPending, ResolveGitConflict, OpenURL } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'

type BannerState = 'idle' | 'available' | 'downloading' | 'ready_to_restart' | 'download_failed'

type GitConflictInfo = {
  message: string
  files: string[]
}

const feedbackIssueURL = 'https://github.com/shinerio/skillflow/issues/new/choose'

function parseConflictPayload(data: string): GitConflictInfo {
  try {
    const parsed = JSON.parse(data)
    if (typeof parsed === 'string') return { message: parsed, files: [] }
    return {
      message: parsed?.message ?? '',
      files: Array.isArray(parsed?.files) ? parsed.files.filter((f: any) => typeof f === 'string' && f.trim() !== '') : [],
    }
  } catch {
    return { message: data, files: [] }
  }
}

export default function App() {
  const [bannerState, setBannerState] = useState<BannerState>('idle')
  const [updateInfo, setUpdateInfo] = useState<main.AppUpdateInfo | null>(null)
  const [dismissed, setDismissed] = useState(false)

  const [conflictOpen, setConflictOpen] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<GitConflictInfo>({ message: '', files: [] })
  const [resolving, setResolving] = useState(false)
  const [resolveError, setResolveError] = useState('')

  const handleResolve = async (useLocal: boolean) => {
    setResolving(true)
    setResolveError('')
    try {
      await ResolveGitConflict(useLocal)
      setConflictOpen(false)
    } catch (e: any) {
      setResolveError(String(e?.message ?? e ?? '操作失败，请重试'))
    } finally {
      setResolving(false)
    }
  }

  useEffect(() => {
    EventsOn('app.update.available', (data: main.AppUpdateInfo) => {
      setUpdateInfo(data)
      setBannerState('available')
    })
    EventsOn('app.update.download.done', () => {
      setBannerState('ready_to_restart')
    })
    EventsOn('app.update.download.fail', () => {
      setBannerState('download_failed')
    })
    EventsOn('git.conflict', (data: string) => {
      setConflictInfo(parseConflictPayload(data))
      setResolveError('')
      setConflictOpen(true)
    })
    // Check for a conflict that happened before the UI was ready (e.g. startup pull)
    GetGitConflictPending().then(pending => { if (pending) setConflictOpen(true) })
  }, [])

  const handleDownload = () => {
    if (!updateInfo?.downloadUrl) return
    setBannerState('downloading')
    DownloadAppUpdate(updateInfo.downloadUrl)
  }

  const handleRestart = () => {
    ApplyAppUpdate()
  }

  const showBanner = !dismissed && bannerState !== 'idle'

  return (
    <BrowserRouter>
      <div className="flex h-screen bg-gray-950 text-gray-100 flex-col">
        {conflictOpen && (
          <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gray-800 rounded-2xl p-6 w-[420px] border border-gray-700 shadow-2xl">
              <div className="flex items-center gap-2 mb-3">
                <AlertTriangle size={18} className="text-amber-400" />
                <span className="font-semibold text-base">Git 同步冲突</span>
              </div>
              <p className="text-sm text-gray-300 mb-2">
                本地 Skills 与远端仓库存在冲突，请选择以哪一方为准：
              </p>
              {conflictInfo.files.length > 0 && (
                <div className="mb-3">
                  <p className="text-xs text-gray-400 mb-1.5">冲突相关文件（{conflictInfo.files.length}）</p>
                  <div className="max-h-28 overflow-y-auto rounded-lg border border-gray-700 bg-gray-900/70 px-2 py-1.5">
                    {conflictInfo.files.slice(0, 30).map((f, i) => (
                      <div key={`${f}-${i}`} className="font-mono text-[11px] text-gray-300 truncate">{f}</div>
                    ))}
                    {conflictInfo.files.length > 30 && (
                      <div className="text-[11px] text-gray-500">... 还有 {conflictInfo.files.length - 30} 个文件</div>
                    )}
                  </div>
                </div>
              )}
              {conflictInfo.message && (
                <div className="mb-3 rounded-lg border border-gray-700 bg-gray-900/70 px-2 py-1.5">
                  <p className="text-[11px] text-gray-500 mb-1">Git 输出</p>
                  <pre className="text-[11px] text-gray-300 whitespace-pre-wrap break-all max-h-20 overflow-y-auto">{conflictInfo.message}</pre>
                </div>
              )}
              <ul className="text-xs text-gray-400 list-disc list-inside mb-6 space-y-1">
                <li><span className="text-white font-medium">以本地为准</span> — 保留本地内容，强制推送到远端</li>
                <li><span className="text-white font-medium">以远端为准</span> — 丢弃本地冲突部分，恢复为远端内容</li>
              </ul>
              {resolveError && (
                <p className="mb-3 text-xs text-red-400 bg-red-950/50 border border-red-800 rounded-lg px-3 py-2 break-all">{resolveError}</p>
              )}
              <div className="flex gap-3 justify-end">
                <button
                  onClick={() => handleResolve(false)}
                  disabled={resolving}
                  className="flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg bg-gray-700 hover:bg-gray-600 disabled:opacity-50"
                >
                  {resolving ? <RefreshCw size={13} className="animate-spin" /> : <Download size={13} />}
                  以远端为准
                </button>
                <button
                  onClick={() => handleResolve(true)}
                  disabled={resolving}
                  className="flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50"
                >
                  {resolving ? <RefreshCw size={13} className="animate-spin" /> : <GitMerge size={13} />}
                  以本地为准
                </button>
              </div>
            </div>
          </div>
        )}
        {showBanner && (
          <UpdateBanner
            state={bannerState}
            info={updateInfo}
            onDownload={handleDownload}
            onRestart={handleRestart}
            onDismiss={() => setDismissed(true)}
          />
        )}
        <div className="flex flex-1 overflow-hidden">
          <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col p-4 gap-1">
            <h1 className="text-lg font-bold mb-6 px-2">SkillFlow</h1>
            <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
            <p className="text-xs text-gray-500 px-2 mt-3 mb-1">同步管理</p>
            <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
            <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
            <NavItem to="/starred" icon={<Star size={16} />} label="仓库收藏" end={false} />
            <div className="flex-1" />
            <div className="flex flex-col gap-1">
              <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
              <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
              <button
                onClick={() => OpenURL(feedbackIssueURL)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-gray-400 hover:bg-gray-800 hover:text-white transition-colors"
              >
                <MessageSquareWarning size={16} />
                意见反馈
              </button>
            </div>
          </aside>
          <main className="flex-1 overflow-auto">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/sync/push" element={<SyncPush />} />
              <Route path="/sync/pull" element={<SyncPull />} />
              <Route path="/backup" element={<Backup />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/starred" element={<StarredRepos />} />
              <Route path="/starred/:repoEncoded" element={<StarredRepos />} />
            </Routes>
          </main>
        </div>
      </div>
    </BrowserRouter>
  )
}

interface UpdateBannerProps {
  state: BannerState
  info: main.AppUpdateInfo | null
  onDownload: () => void
  onRestart: () => void
  onDismiss: () => void
}

function UpdateBanner({ state, info, onDownload, onRestart, onDismiss }: UpdateBannerProps) {
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-indigo-700 text-white text-sm shrink-0">
      <div className="flex items-center gap-3">
        {state === 'available' && (
          <>
            <span>新版本可用: {info?.latestVersion}</span>
            {info?.canAutoUpdate ? (
              <button
                onClick={onDownload}
                className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
              >
                <Download size={12} />
                立即下载
              </button>
            ) : (
              <a
                href={info?.releaseUrl}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
              >
                查看详情
              </a>
            )}
          </>
        )}
        {state === 'downloading' && (
          <>
            <RefreshCw size={14} className="animate-spin" />
            <span>正在下载 {info?.latestVersion}...</span>
          </>
        )}
        {state === 'ready_to_restart' && (
          <>
            <span>已下载完成，点击重启以完成更新</span>
            <button
              onClick={onRestart}
              className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
            >
              立即重启
            </button>
          </>
        )}
        {state === 'download_failed' && (
          <>
            <span>下载失败，请手动下载</span>
            <a
              href={info?.releaseUrl}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1 px-2 py-0.5 bg-white text-indigo-700 rounded text-xs font-medium hover:bg-indigo-100"
            >
              前往下载页
            </a>
          </>
        )}
      </div>
      {state !== 'downloading' && (
        <button onClick={onDismiss} className="text-white hover:text-indigo-200 ml-4">
          <X size={14} />
        </button>
      )}
    </div>
  )
}

function NavItem({ to, icon, label, end = true }: { to: string; icon: React.ReactNode; label: string; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
          isActive ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'
        }`
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}
