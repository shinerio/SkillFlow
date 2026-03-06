import { useEffect, useState } from 'react'
import { BackupNow, ListCloudFiles, RestoreFromCloud, GetConfig, GetGitConflictPending, ResolveGitConflict } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Cloud, Upload, Download, RefreshCw, GitMerge, AlertTriangle } from 'lucide-react'

type GitConflictInfo = {
  message: string
  files: string[]
}

export default function Backup() {
  const [files, setFiles] = useState<Array<{ path: string; size: number }>>([])
  const [status, setStatus] = useState<'idle' | 'backing-up' | 'done' | 'error'>('idle')
  const [currentFile, setCurrentFile] = useState('')
  const [cloudEnabled, setCloudEnabled] = useState(false)
  const [isGit, setIsGit] = useState(false)

  // Git conflict dialog state
  const [conflictOpen, setConflictOpen] = useState(false)
  const [resolving, setResolving] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<GitConflictInfo>({ message: '', files: [] })

  const parseConflictPayload = (data: string): GitConflictInfo => {
    try {
      const parsed = JSON.parse(data)
      if (typeof parsed === 'string') {
        return { message: parsed, files: [] }
      }
      return {
        message: parsed?.message ?? '',
        files: Array.isArray(parsed?.files) ? parsed.files.filter((f: any) => typeof f === 'string' && f.trim() !== '') : [],
      }
    } catch {
      return { message: data, files: [] }
    }
  }

  useEffect(() => {
    GetConfig().then(cfg => {
      setCloudEnabled(cfg?.cloud?.enabled ?? false)
      setIsGit(cfg?.cloud?.provider === 'git')
    })
    // Check for a conflict that happened before the UI was ready (startup pull)
    GetGitConflictPending().then(pending => { if (pending) setConflictOpen(true) })

    EventsOn('backup.started', () => setStatus('backing-up'))
    EventsOn('backup.progress', (data: string) => {
      try { setCurrentFile(JSON.parse(data).currentFile ?? '') } catch {}
    })
    EventsOn('backup.completed', () => { setStatus('done'); loadFiles() })
    EventsOn('backup.failed', () => setStatus('error'))
    EventsOn('git.sync.started', () => setStatus('backing-up'))
    EventsOn('git.sync.completed', () => { setStatus('done'); loadFiles() })
    EventsOn('git.sync.failed', () => setStatus('error'))
    EventsOn('git.conflict', (data: string) => {
      setConflictInfo(parseConflictPayload(data))
      setConflictOpen(true)
    })
  }, [])

  const loadFiles = async () => {
    const f = await ListCloudFiles()
    const normalized = (f ?? [])
      .map((item: any) => {
        const path = item?.path ?? item?.Path ?? ''
        const rawSize = item?.size ?? item?.Size ?? 0
        const size = typeof rawSize === 'number' ? rawSize : Number(rawSize) || 0
        return { path, size }
      })
      .filter((item: { path: string }) => item.path !== '')
    setFiles(normalized)
  }

  const handleResolve = async (useLocal: boolean) => {
    setResolving(true)
    try {
      await ResolveGitConflict(useLocal)
      setConflictOpen(false)
      setStatus('done')
      loadFiles()
    } catch {
      setStatus('error')
    } finally {
      setResolving(false)
    }
  }

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><Cloud size={18} /> 云备份</h2>

      {!cloudEnabled && (
        <div className="bg-yellow-900/30 border border-yellow-700/50 rounded-xl p-4 mb-6 text-sm text-yellow-300">
          云备份未启用。请前往设置 → 云存储完成配置。
        </div>
      )}

      <div className="flex gap-3 mb-8">
        <button
          onClick={async () => {
            try {
              setStatus('backing-up')
              await BackupNow()
            } catch {
              setStatus('error')
            }
          }}
          disabled={!cloudEnabled || status === 'backing-up'}
          className="flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
        >
          {status === 'backing-up' ? <RefreshCw size={14} className="animate-spin" /> : <Upload size={14} />}
          {status === 'backing-up' ? `备份中 ${currentFile}` : '立即备份'}
        </button>
        <button
          onClick={async () => {
            try {
              setStatus('backing-up')
              await RestoreFromCloud()
              loadFiles()
              if (!isGit) setStatus('done')
            } catch {
              setStatus('error')
            }
          }}
          disabled={!cloudEnabled}
          className="flex items-center gap-2 px-5 py-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm disabled:opacity-50"
        ><Download size={14} /> {isGit ? '拉取远端' : '从云端恢复'}</button>
        <button onClick={loadFiles} className="flex items-center gap-2 px-4 py-2.5 text-gray-400 hover:text-white rounded-lg hover:bg-gray-800 text-sm">
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      {status === 'done' && <p className="mb-4 text-sm text-green-400">{isGit ? 'Git 同步完成' : '备份完成'}</p>}
      {status === 'error' && <p className="mb-4 text-sm text-red-400">{isGit ? 'Git 同步失败，请检查仓库配置' : '备份失败，请检查云存储配置'}</p>}

      {files.length > 0 && (
        <div>
          <p className="text-sm text-gray-400 mb-3">{isGit ? 'Git 跟踪文件' : '云端文件'}（{files.length} 个）</p>
          <div className="max-h-96 overflow-y-auto border border-gray-800 rounded-xl divide-y divide-gray-800">
            {files.map((f, i) => (
              <div key={i} className="flex items-center justify-between px-4 py-2.5 text-sm">
                <span className="text-gray-300 font-mono text-xs">{f.path}</span>
                <span className="text-gray-500 text-xs">{(f.size / 1024).toFixed(1)} KB</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Git conflict resolution dialog */}
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
    </div>
  )
}
