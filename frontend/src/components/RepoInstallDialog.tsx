import { useState, useEffect } from 'react'
import { X, Search, Github, Check } from 'lucide-react'
import {
  ScanGitHub,
  InstallFromGitHub,
  InstallFromGitHubToTool,
  GetEnabledTools,
  ListCategories,
} from '../../wailsjs/go/main/App'
import { install, config } from '../../wailsjs/go/models'

interface Props {
  repoURL: string
  repoName: string
  onClose: () => void
}

export default function RepoInstallDialog({ repoURL, repoName, onClose }: Props) {
  const [scanning, setScanning] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [candidates, setCandidates] = useState<install.SkillCandidate[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [scanError, setScanError] = useState('')
  const [installError, setInstallError] = useState('')
  const [done, setDone] = useState(false)

  // Target: 'skillflow' installs to SkillFlow storage; 'tool' installs directly to a tool dir.
  const [target, setTarget] = useState<'skillflow' | 'tool'>('skillflow')
  const [category, setCategory] = useState('')
  const [categories, setCategories] = useState<string[]>([])
  const [tools, setTools] = useState<config.ToolConfig[]>([])
  const [selectedTool, setSelectedTool] = useState('')

  useEffect(() => {
    ListCategories().then(cats => {
      const list = cats || []
      setCategories(list)
      if (list.length > 0) setCategory(list[0])
    })
    GetEnabledTools().then(ts => {
      const list = ts || []
      setTools(list)
      if (list.length > 0) setSelectedTool(list[0].name)
    })
    handleScan()
  }, [])

  async function handleScan() {
    setScanning(true)
    setScanError('')
    setCandidates([])
    setDone(false)
    try {
      const result = await ScanGitHub(repoURL)
      const list = result || []
      setCandidates(list)
      // Pre-select uninstalled skills.
      const next = new Set<string>()
      for (const c of list) {
        if (!c.Installed) next.add(c.Name)
      }
      setSelected(next)
    } catch (e: any) {
      setScanError(e.message || String(e))
    } finally {
      setScanning(false)
    }
  }

  function toggleAll() {
    const uninstalled = candidates.filter(c => !c.Installed).map(c => c.Name)
    if (uninstalled.every(n => selected.has(n))) {
      setSelected(new Set())
    } else {
      setSelected(new Set(uninstalled))
    }
  }

  async function handleInstall() {
    const toInstall = candidates.filter(c => selected.has(c.Name))
    if (toInstall.length === 0) return
    setInstalling(true)
    setInstallError('')
    try {
      if (target === 'skillflow') {
        await InstallFromGitHub(repoURL, toInstall, category)
      } else {
        await InstallFromGitHubToTool(repoURL, toInstall, selectedTool)
      }
      setDone(true)
    } catch (e: any) {
      setInstallError(e.message || String(e))
    } finally {
      setInstalling(false)
    }
  }

  const uninstalledCount = candidates.filter(c => !c.Installed).length
  const selectedCount = selected.size

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-gray-700 rounded-2xl w-full max-w-lg flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-gray-800">
          <div className="flex items-center gap-2 min-w-0">
            <Github size={17} className="text-gray-400 shrink-0" />
            <div className="min-w-0">
              <h2 className="text-sm font-semibold">{repoName}</h2>
              <p className="text-xs text-gray-500 truncate max-w-xs">{repoURL}</p>
            </div>
          </div>
          <button onClick={onClose} className="text-gray-500 hover:text-white transition-colors ml-2 shrink-0">
            <X size={18} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-5 flex flex-col gap-5">
          {/* Scan results */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                技能列表{candidates.length > 0 && ` (${candidates.length})`}
              </span>
              <button
                onClick={handleScan}
                disabled={scanning}
                className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300 transition-colors disabled:opacity-50"
              >
                <Search size={12} />
                {scanning ? '扫描中...' : '重新扫描'}
              </button>
            </div>

            {scanning && (
              <div className="flex items-center justify-center py-8 text-gray-500 text-sm gap-2">
                <div className="animate-spin w-4 h-4 border-2 border-indigo-500 border-t-transparent rounded-full" />
                正在扫描仓库...
              </div>
            )}

            {scanError && (
              <div className="bg-red-900/30 border border-red-700 rounded-lg px-3 py-2 text-xs text-red-300">
                {scanError}
              </div>
            )}

            {!scanning && candidates.length > 0 && (
              <>
                <div className="flex items-center justify-between mb-1.5">
                  <button
                    onClick={toggleAll}
                    className="text-xs text-gray-400 hover:text-white transition-colors"
                  >
                    {uninstalledCount > 0 && selected.size === uninstalledCount ? '取消全选' : '全选'}
                  </button>
                  <span className="text-xs text-gray-500">已选 {selectedCount} 个</span>
                </div>
                <div className="space-y-1 max-h-52 overflow-auto">
                  {candidates.map(c => (
                    <label
                      key={c.Name}
                      className={`flex items-center gap-2.5 px-3 py-2 rounded-lg transition-colors ${
                        c.Installed ? 'opacity-50 cursor-default' : 'cursor-pointer hover:bg-gray-800'
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={selected.has(c.Name)}
                        disabled={c.Installed}
                        onChange={e => {
                          const next = new Set(selected)
                          if (e.target.checked) next.add(c.Name)
                          else next.delete(c.Name)
                          setSelected(next)
                        }}
                        className="accent-indigo-500"
                      />
                      <span className="text-sm flex-1">{c.Name}</span>
                      {c.Installed && (
                        <span className="text-xs bg-gray-700 text-gray-400 px-1.5 py-0.5 rounded">已安装</span>
                      )}
                    </label>
                  ))}
                </div>
              </>
            )}

            {!scanning && !scanError && candidates.length === 0 && (
              <p className="text-sm text-gray-500 text-center py-6">未找到技能，请检查仓库地址</p>
            )}
          </div>

          {/* Target selection */}
          {candidates.length > 0 && !done && (
            <div>
              <p className="text-xs font-medium text-gray-400 uppercase tracking-wider mb-2">安装目标</p>
              <div className="space-y-2">
                {/* SkillFlow storage option */}
                <label className="flex items-center gap-2.5 cursor-pointer">
                  <input
                    type="radio"
                    name="install-target"
                    checked={target === 'skillflow'}
                    onChange={() => setTarget('skillflow')}
                    className="accent-indigo-500"
                  />
                  <span className="text-sm flex-1">SkillFlow 仓库</span>
                  {target === 'skillflow' && categories.length > 0 && (
                    <select
                      value={category}
                      onChange={e => setCategory(e.target.value)}
                      className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
                    >
                      {categories.map(c => <option key={c} value={c}>{c}</option>)}
                    </select>
                  )}
                </label>

                {/* Direct-to-tool option */}
                <label className={`flex items-center gap-2.5 ${tools.length === 0 ? 'cursor-default' : 'cursor-pointer'}`}>
                  <input
                    type="radio"
                    name="install-target"
                    checked={target === 'tool'}
                    onChange={() => setTarget('tool')}
                    className="accent-indigo-500"
                    disabled={tools.length === 0}
                  />
                  <span className={`text-sm flex-1 ${tools.length === 0 ? 'text-gray-500' : ''}`}>
                    直接安装到工具
                  </span>
                  {target === 'tool' && tools.length > 0 && (
                    <select
                      value={selectedTool}
                      onChange={e => setSelectedTool(e.target.value)}
                      className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
                    >
                      {tools.map(t => <option key={t.name} value={t.name}>{t.name}</option>)}
                    </select>
                  )}
                  {tools.length === 0 && (
                    <span className="text-xs text-gray-600">(无已启用工具)</span>
                  )}
                </label>
              </div>
            </div>
          )}

          {installError && (
            <div className="bg-red-900/30 border border-red-700 rounded-lg px-3 py-2 text-xs text-red-300">
              {installError}
            </div>
          )}

          {done && (
            <div className="flex items-center gap-2 bg-green-900/30 border border-green-700 rounded-lg px-3 py-2 text-xs text-green-300">
              <Check size={14} />
              安装完成！
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-gray-800 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg text-sm transition-colors"
          >
            {done ? '关闭' : '取消'}
          </button>
          {!done && (
            <button
              onClick={handleInstall}
              disabled={selectedCount === 0 || installing || scanning}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-sm transition-colors"
            >
              {installing ? '安装中...' : `安装${selectedCount > 0 ? ` ${selectedCount} 个` : ''}`}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
