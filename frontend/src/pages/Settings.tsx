import { useEffect, useState } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomTool, RemoveCustomTool, OpenFolderDialog, CheckAppUpdate, GetAppVersion } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings, Globe, FolderOpen, RefreshCw } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

type Tab = 'tools' | 'cloud' | 'general' | 'network'
type ProxyMode = 'none' | 'system' | 'manual'

function Toggle({ enabled, onToggle }: { enabled: boolean; onToggle: () => void }) {
  return (
    <div
      onClick={onToggle}
      className={`w-9 h-5 rounded-full transition-colors relative cursor-pointer ${enabled ? 'bg-indigo-600' : 'bg-gray-600'}`}
    >
      <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
    </div>
  )
}

export default function SettingsPage() {
  const [tab, setTab] = useState<Tab>('tools')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newTool, setNewTool] = useState({ name: '', pushDir: '' })
  const [newScanDirs, setNewScanDirs] = useState<Record<string, string>>({})
  const [appVersion, setAppVersion] = useState('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updateResult, setUpdateResult] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders(), GetAppVersion()]).then(([c, p, v]) => {
      setCfg(c)
      setProviders(p ?? [])
      setAppVersion(v as string)
    })
  }, [])

  const checkUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateResult(null)
    try {
      const info = await CheckAppUpdate()
      if (info.hasUpdate) {
        setUpdateResult(`发现新版本 ${info.latestVersion}，请查看顶部横幅`)
      } else {
        setUpdateResult(`已是最新版本 (${info.currentVersion})`)
      }
    } catch {
      setUpdateResult('检测失败，请检查网络')
    } finally {
      setCheckingUpdate(false)
    }
  }

  const save = async () => {
    setSaving(true)
    await SaveConfig(cfg)
    setSaving(false)
  }

  const updateTool = (name: string, field: string, value: any) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => t.name === name ? { ...t, [field]: value } : t)
    }))
  }

  const addScanDir = (name: string) => {
    const path = (newScanDirs[name] ?? '').trim()
    if (!path) return
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        const current = t.scanDirs ?? []
        if (current.includes(path)) return t
        return { ...t, scanDirs: [...current, path] }
      })
    }))
    setNewScanDirs((prev) => ({ ...prev, [name]: '' }))
  }

  const updateScanDir = (name: string, index: number, value: string) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        const next = [...(t.scanDirs ?? [])]
        next[index] = value
        return { ...t, scanDirs: next }
      })
    }))
  }

  const removeScanDir = (name: string, index: number) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => {
        if (t.name !== name) return t
        return { ...t, scanDirs: (t.scanDirs ?? []).filter((_: string, i: number) => i !== index) }
      })
    }))
  }

  const setProxyMode = (mode: ProxyMode) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, Mode: mode } }))
  }

  const setProxyURL = (url: string) => {
    setCfg((prev: any) => ({ ...prev, proxy: { ...prev.proxy, URL: url } }))
  }

  const pickDir = async (onPick: (path: string) => void) => {
    const dir = await OpenFolderDialog()
    if (dir) onPick(dir)
  }

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)
  const proxyMode: ProxyMode = (cfg?.proxy?.Mode as ProxyMode) || 'none'

  if (!cfg) return <div className="p-8 text-gray-400">加载中...</div>

  return (
    <div className="p-8 max-w-2xl">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold flex items-center gap-2"><Settings size={18} /> 设置</h2>
        <div className="flex items-center gap-3">
          {updateResult && (
            <span className="text-xs text-gray-400">{updateResult}</span>
          )}
          {appVersion && (
            <span className="text-xs text-gray-500">v{appVersion.replace(/^v/, '')}</span>
          )}
          <button
            onClick={checkUpdate}
            disabled={checkingUpdate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw size={12} className={checkingUpdate ? 'animate-spin' : ''} />
            检测更新
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-gray-800 rounded-xl p-1 w-fit">
        {([['tools', '工具路径'], ['cloud', '云存储'], ['general', '通用'], ['network', '网络']] as [Tab, string][]).map(([v, label]) => (
          <button key={v} onClick={() => setTab(v)}
            className={`px-4 py-1.5 rounded-lg text-sm transition-colors ${tab === v ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'}`}
          >{label}</button>
        ))}
      </div>

      {/* Tools tab */}
      {tab === 'tools' && (
        <div className="space-y-4">
          {(cfg.tools ?? []).map((t: any) => (
            <div key={t.name} className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <ToolIcon name={t.name} size={28} />
                  <span className="font-medium text-sm">{t.name}</span>
                </div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <span className="text-xs text-gray-400">启用</span>
                  <Toggle enabled={t.enabled} onToggle={() => updateTool(t.name, 'enabled', !t.enabled)} />
                </label>
              </div>

              <div className="mb-3">
                <p className="text-xs text-gray-400 mb-1.5">推送路径（仅 1 个）</p>
                <div className="flex gap-2">
                  <input
                    value={t.pushDir ?? ''}
                    onChange={e => updateTool(t.name, 'pushDir', e.target.value)}
                    className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none focus:border-indigo-500"
                  />
                  <button onClick={() => pickDir(dir => updateTool(t.name, 'pushDir', dir))}
                    className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300" title="选择目录">
                    <FolderOpen size={14} />
                  </button>
                </div>
              </div>

              <div>
                <p className="text-xs text-gray-400 mb-1.5">扫描路径（可多个）</p>
                <div className="space-y-2">
                  {(t.scanDirs ?? []).map((dir: string, idx: number) => (
                    <div key={`${t.name}-scan-${idx}`} className="flex gap-2">
                      <input
                        value={dir}
                        onChange={e => updateScanDir(t.name, idx, e.target.value)}
                        className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none focus:border-indigo-500"
                      />
                      <button onClick={() => pickDir(d => updateScanDir(t.name, idx, d))}
                        className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300" title="选择目录">
                        <FolderOpen size={14} />
                      </button>
                      <button
                        onClick={() => removeScanDir(t.name, idx)}
                        className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300"
                        title="删除扫描路径"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
                <div className="mt-2 flex gap-2">
                  <input
                    value={newScanDirs[t.name] ?? ''}
                    onChange={e => setNewScanDirs(prev => ({ ...prev, [t.name]: e.target.value }))}
                    placeholder="/path/to/scan"
                    className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none focus:border-indigo-500"
                  />
                  <button onClick={() => pickDir(d => setNewScanDirs(prev => ({ ...prev, [t.name]: d })))}
                    className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300" title="选择目录">
                    <FolderOpen size={14} />
                  </button>
                  <button
                    onClick={() => addScanDir(t.name)}
                    className="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm flex items-center gap-1"
                  >
                    <Plus size={14} /> 添加
                  </button>
                </div>
              </div>

              {t.custom && (
                <button
                  onClick={async () => { await RemoveCustomTool(t.name); const c = await GetConfig(); setCfg(c) }}
                  className="mt-2 text-xs text-red-400 hover:text-red-300 flex items-center gap-1"
                ><Trash2 size={12} /> 删除</button>
              )}
            </div>
          ))}

          {/* Add custom tool */}
          <div className="bg-gray-800 rounded-xl p-4 border border-dashed border-gray-600">
            <p className="text-sm text-gray-400 mb-3">添加自定义工具</p>
            <div className="flex gap-2 mb-2">
              <input value={newTool.name} onChange={e => setNewTool(p => ({ ...p, name: e.target.value }))}
                placeholder="工具名称" className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm outline-none" />
            </div>
            <div className="flex gap-2">
              <input value={newTool.pushDir} onChange={e => setNewTool(p => ({ ...p, pushDir: e.target.value }))}
                placeholder="/path/to/push" className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none" />
              <button onClick={() => pickDir(d => setNewTool(p => ({ ...p, pushDir: d })))}
                className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300" title="选择目录">
                <FolderOpen size={14} />
              </button>
              <button
                onClick={async () => {
                  if (newTool.name && newTool.pushDir) {
                    await AddCustomTool(newTool.name, newTool.pushDir)
                    const c = await GetConfig(); setCfg(c)
                    setNewTool({ name: '', pushDir: '' })
                  }
                }}
                className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm flex items-center gap-1"
              ><Plus size={14} /> 添加</button>
            </div>
          </div>
        </div>
      )}

      {/* Cloud tab */}
      {tab === 'cloud' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-400 mb-2">云厂商</p>
            <div className="flex gap-2">
              {providers.map((p: any) => (
                <button key={p.name}
                  onClick={() => setCfg((prev: any) => ({ ...prev, cloud: { ...prev.cloud, provider: p.name } }))}
                  className={`px-4 py-2 rounded-lg text-sm border transition-colors ${cfg.cloud?.provider === p.name ? 'bg-indigo-600 border-indigo-500' : 'bg-gray-800 border-gray-700 hover:border-gray-500'}`}
                >{p.name}</button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              {/* Bucket / remote-path fields are not applicable for the git provider */}
              {cfg.cloud?.provider !== 'git' && (
                <div>
                  <p className="text-sm text-gray-400 mb-2">存储桶</p>
                  <input value={cfg.cloud?.bucketName ?? ''} onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, bucketName: e.target.value } }))}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
                </div>
              )}
              {selectedProvider.fields.map((f: any) => (
                <div key={f.key}>
                  <p className="text-sm text-gray-400 mb-2">{f.label}</p>
                  <input
                    type={f.secret ? 'password' : 'text'}
                    placeholder={f.placeholder ?? ''}
                    value={cfg.cloud?.credentials?.[f.key] ?? ''}
                    onChange={e => setCfg((p: any) => ({
                      ...p, cloud: { ...p.cloud, credentials: { ...p.cloud?.credentials, [f.key]: e.target.value } }
                    }))}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500 font-mono"
                  />
                </div>
              ))}
              <div>
                <p className="text-sm text-gray-400 mb-2">定时自动同步间隔（分钟，0 表示仅在变更后同步）</p>
                <input
                  type="number"
                  min={0}
                  value={cfg.cloud?.syncIntervalMinutes ?? 0}
                  onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, syncIntervalMinutes: parseInt(e.target.value) || 0 } }))}
                  className="w-32 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500"
                />
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <Toggle
                  enabled={!!cfg.cloud?.enabled}
                  onToggle={() => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, enabled: !p.cloud?.enabled } }))}
                />
                <span className="text-sm text-gray-300">启用自动云备份</span>
              </label>
            </>
          )}
        </div>
      )}

      {/* General tab */}
      {tab === 'general' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-400 mb-2">本地 Skills 存储目录</p>
            <div className="flex gap-2">
              <input value={cfg.skillsStorageDir ?? ''} onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
                className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono outline-none focus:border-indigo-500" />
              <button onClick={() => pickDir(d => setCfg((p: any) => ({ ...p, skillsStorageDir: d })))}
                className="px-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-gray-300" title="选择目录">
                <FolderOpen size={16} />
              </button>
            </div>
          </div>
          <div>
            <p className="text-sm text-gray-400 mb-2">从工具拉取时的默认分类</p>
            <input value={cfg.defaultCategory ?? ''} onChange={e => setCfg((p: any) => ({ ...p, defaultCategory: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
          </div>
        </div>
      )}

      {/* Network tab */}
      {tab === 'network' && (
        <div className="space-y-6">
          <div>
            <p className="text-sm text-gray-400 mb-1 flex items-center gap-1.5">
              <Globe size={14} /> 代理设置
            </p>
            <p className="text-xs text-gray-500 mb-4">
              代理用于远程仓库相关操作（扫描仓库、安装 Skill、检查更新）
            </p>

            <div className="space-y-2">
              {([
                ['none',   '不使用代理',   '直连，不通过任何代理'],
                ['system', '使用系统代理', '读取 HTTP_PROXY / HTTPS_PROXY 环境变量'],
                ['manual', '手动配置',     '自定义代理地址'],
              ] as [ProxyMode, string, string][]).map(([mode, label, desc]) => (
                <div
                  key={mode}
                  onClick={() => setProxyMode(mode)}
                  className={`flex items-start gap-3 p-3 rounded-xl border cursor-pointer transition-colors select-none ${
                    proxyMode === mode
                      ? 'bg-indigo-900/30 border-indigo-500'
                      : 'bg-gray-800 border-gray-700 hover:border-gray-500'
                  }`}
                >
                  <div className={`mt-0.5 w-4 h-4 rounded-full border-2 flex items-center justify-center shrink-0 transition-colors ${
                    proxyMode === mode ? 'border-indigo-400 bg-indigo-500' : 'border-gray-500'
                  }`}>
                    {proxyMode === mode && <div className="w-1.5 h-1.5 bg-white rounded-full" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium leading-snug">{label}</p>
                    <p className="text-xs text-gray-500 mt-0.5">{desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {proxyMode === 'manual' && (
            <div>
              <p className="text-sm text-gray-400 mb-2">代理地址</p>
              <input
                value={cfg.proxy?.URL ?? ''}
                onChange={e => setProxyURL(e.target.value)}
                placeholder="http://127.0.0.1:7890"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono outline-none focus:border-indigo-500"
              />
              <p className="mt-1.5 text-xs text-gray-500">
                支持 <code className="bg-gray-800 px-1 rounded">http://</code>、
                <code className="bg-gray-800 px-1 rounded">https://</code>、
                <code className="bg-gray-800 px-1 rounded">socks5://</code> 格式
              </p>
            </div>
          )}
        </div>
      )}

      <button onClick={save} disabled={saving}
        className="mt-8 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
        {saving ? '保存中...' : '保存设置'}
      </button>
    </div>
  )
}
