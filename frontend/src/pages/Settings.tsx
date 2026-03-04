import { useEffect, useState } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomTool, RemoveCustomTool } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

type Tab = 'tools' | 'cloud' | 'general'

export default function SettingsPage() {
  const [tab, setTab] = useState<Tab>('tools')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newTool, setNewTool] = useState({ name: '', skillsDir: '' })

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders()]).then(([c, p]) => {
      setCfg(c)
      setProviders(p ?? [])
    })
  }, [])

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

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)

  if (!cfg) return <div className="p-8 text-gray-400">加载中...</div>

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><Settings size={18} /> 设置</h2>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-gray-800 rounded-xl p-1 w-fit">
        {([['tools', '工具路径'], ['cloud', '云存储'], ['general', '通用']] as [Tab, string][]).map(([v, label]) => (
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
                  <div
                    onClick={() => updateTool(t.name, 'enabled', !t.enabled)}
                    className={`w-9 h-5 rounded-full transition-colors relative ${t.enabled ? 'bg-indigo-600' : 'bg-gray-600'}`}
                  >
                    <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${t.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
                  </div>
                </label>
              </div>
              <input
                value={t.skillsDir}
                onChange={e => updateTool(t.name, 'skillsDir', e.target.value)}
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none focus:border-indigo-500"
              />
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
              <input value={newTool.skillsDir} onChange={e => setNewTool(p => ({ ...p, skillsDir: e.target.value }))}
                placeholder="/path/to/skills" className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none" />
              <button
                onClick={async () => {
                  if (newTool.name && newTool.skillsDir) {
                    await AddCustomTool(newTool.name, newTool.skillsDir)
                    const c = await GetConfig(); setCfg(c)
                    setNewTool({ name: '', skillsDir: '' })
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
              <div>
                <p className="text-sm text-gray-400 mb-2">存储桶</p>
                <input value={cfg.cloud?.bucketName ?? ''} onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, bucketName: e.target.value } }))}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
              </div>
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
              <label className="flex items-center gap-3 cursor-pointer">
                <div
                  onClick={() => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, enabled: !p.cloud?.enabled } }))}
                  className={`w-9 h-5 rounded-full transition-colors relative ${cfg.cloud?.enabled ? 'bg-indigo-600' : 'bg-gray-600'}`}
                >
                  <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${cfg.cloud?.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
                </div>
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
            <input value={cfg.skillsStorageDir ?? ''} onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono outline-none focus:border-indigo-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400 mb-2">从工具拉取时的默认分类</p>
            <input value={cfg.defaultCategory ?? ''} onChange={e => setCfg((p: any) => ({ ...p, defaultCategory: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
          </div>
        </div>
      )}

      <button onClick={save} disabled={saving}
        className="mt-8 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
        {saving ? '保存中...' : '保存设置'}
      </button>
    </div>
  )
}
