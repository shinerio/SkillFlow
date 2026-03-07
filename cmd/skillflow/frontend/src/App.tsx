import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star, X, Download, RefreshCw, AlertTriangle, GitMerge, MessageSquareWarning, ExternalLink } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import StarredRepos from './pages/StarredRepos'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DownloadAppUpdate, ApplyAppUpdate, GetGitConflictPending, ResolveGitConflict, OpenURL, SetSkippedUpdateVersion } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'

type UpdateDialogState = 'idle' | 'available' | 'downloading' | 'ready_to_restart' | 'download_failed'

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
  const [dialogState, setDialogState] = useState<UpdateDialogState>('idle')
  const [updateInfo, setUpdateInfo] = useState<main.AppUpdateInfo | null>(null)

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
      setDialogState('available')
    })
    EventsOn('app.update.download.done', () => {
      setDialogState('ready_to_restart')
    })
    EventsOn('app.update.download.fail', () => {
      setDialogState('download_failed')
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
    setDialogState('downloading')
    DownloadAppUpdate(updateInfo.downloadUrl)
  }

  const handleRestart = () => {
    ApplyAppUpdate()
  }

  const handleSkip = async () => {
    if (updateInfo?.latestVersion) {
      await SetSkippedUpdateVersion(updateInfo.latestVersion)
    }
    setDialogState('idle')
  }

  const handleOpenRelease = () => {
    if (updateInfo?.releaseUrl) {
      OpenURL(updateInfo.releaseUrl)
    }
    setDialogState('idle')
  }

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
        {dialogState !== 'idle' && (
          <UpdateDialog
            state={dialogState}
            info={updateInfo}
            onDownload={handleDownload}
            onRestart={handleRestart}
            onOpenRelease={handleOpenRelease}
            onSkip={handleSkip}
            onClose={() => setDialogState('idle')}
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

interface UpdateDialogProps {
  state: UpdateDialogState
  info: main.AppUpdateInfo | null
  onDownload: () => void
  onRestart: () => void
  onOpenRelease: () => void
  onSkip: () => void
  onClose: () => void
}

function UpdateDialog({ state, info, onDownload, onRestart, onOpenRelease, onSkip, onClose }: UpdateDialogProps) {
  const isDownloading = state === 'downloading'

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-2xl p-6 w-[440px] border border-gray-700 shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Download size={18} className="text-indigo-400" />
            <span className="font-semibold text-base">
              {state === 'ready_to_restart' ? '更新已就绪' : state === 'download_failed' ? '下载失败' : '发现新版本'}
            </span>
          </div>
          {!isDownloading && (
            <button onClick={onClose} className="text-gray-400 hover:text-gray-200">
              <X size={16} />
            </button>
          )}
        </div>

        {/* Body */}
        {(state === 'available' || state === 'downloading') && (
          <>
            <p className="text-sm text-gray-300 mb-1">
              最新版本：<span className="font-mono text-indigo-300 font-medium">{info?.latestVersion}</span>
            </p>
            <p className="text-sm text-gray-400 mb-4">
              当前版本：<span className="font-mono text-gray-500">{info?.currentVersion}</span>
            </p>
            {info?.releaseNotes && (
              <div className="mb-4 rounded-lg border border-gray-700 bg-gray-900/60 px-3 py-2 max-h-32 overflow-y-auto">
                <p className="text-[11px] text-gray-500 mb-1">更新说明</p>
                <pre className="text-xs text-gray-300 whitespace-pre-wrap break-all">{info.releaseNotes}</pre>
              </div>
            )}
          </>
        )}

        {state === 'downloading' && (
          <div className="flex items-center gap-2 mb-4 text-sm text-gray-300">
            <RefreshCw size={14} className="animate-spin text-indigo-400" />
            <span>正在下载 {info?.latestVersion}，请稍候...</span>
          </div>
        )}

        {state === 'ready_to_restart' && (
          <p className="text-sm text-gray-300 mb-4">
            新版本已下载完成，点击下方按钮重启应用以完成更新。
          </p>
        )}

        {state === 'download_failed' && (
          <p className="text-sm text-gray-300 mb-4">
            自动下载失败，请前往 Release 页面手动下载最新版本。
          </p>
        )}

        {/* Actions */}
        {state === 'available' && (
          <div className="flex flex-col gap-2">
            {info?.canAutoUpdate && (
              <button
                onClick={onDownload}
                className="flex items-center justify-center gap-2 w-full px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-xl text-sm font-medium transition-colors"
              >
                <Download size={14} />
                下载并自动重启更新
              </button>
            )}
            <button
              onClick={onOpenRelease}
              className="flex items-center justify-center gap-2 w-full px-4 py-2.5 bg-gray-700 hover:bg-gray-600 rounded-xl text-sm transition-colors"
            >
              <ExternalLink size={14} />
              前往 Release 页面手动下载
            </button>
            <button
              onClick={onSkip}
              className="flex items-center justify-center gap-2 w-full px-4 py-2 text-gray-500 hover:text-gray-300 text-sm transition-colors"
            >
              跳过此版本（下次启动不再提示）
            </button>
          </div>
        )}

        {state === 'downloading' && (
          <p className="text-xs text-gray-500 text-center">下载完成后将自动提示重启</p>
        )}

        {state === 'ready_to_restart' && (
          <div className="flex gap-3 justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm rounded-xl bg-gray-700 hover:bg-gray-600 transition-colors"
            >
              稍后重启
            </button>
            <button
              onClick={onRestart}
              className="flex items-center gap-2 px-4 py-2 text-sm rounded-xl bg-indigo-600 hover:bg-indigo-500 transition-colors"
            >
              <RefreshCw size={13} />
              立即重启
            </button>
          </div>
        )}

        {state === 'download_failed' && (
          <div className="flex gap-3 justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm rounded-xl bg-gray-700 hover:bg-gray-600 transition-colors"
            >
              关闭
            </button>
            <button
              onClick={onOpenRelease}
              className="flex items-center gap-2 px-4 py-2 text-sm rounded-xl bg-indigo-600 hover:bg-indigo-500 transition-colors"
            >
              <ExternalLink size={13} />
              前往下载页
            </button>
          </div>
        )}
      </div>
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
