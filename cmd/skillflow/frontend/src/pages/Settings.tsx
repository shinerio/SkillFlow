import { useEffect, useState, type ReactNode } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomTool, RemoveCustomTool, OpenFolderDialog, CheckAppUpdateAndNotify, GetAppVersion, GetLogDir, OpenLogDir } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings, Globe, FolderOpen, RefreshCw, Sun, Moon, Sparkles, Check } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'
import { useThemeContext } from '../contexts/ThemeContext'
import { type Theme } from '../hooks/useTheme'
import { useLanguage } from '../contexts/LanguageContext'

type Tab = 'tools' | 'cloud' | 'general' | 'network'
type ProxyMode = 'none' | 'system' | 'manual'

type ThemePreviewPalette = {
  shell: string
  sidebar: string
  sidebarSelection: string
  search: string
  panel: string
  accent: string
  accentGlow: string
  text: string
  textMuted: string
  divider: string
}

type ThemeOption = {
  id: Theme
  label: string
  tone: string
  description: string
  icon: ReactNode
  preview: ThemePreviewPalette
}

function ThemeOptionCard({ option, active, onSelect }: { option: ThemeOption; active: boolean; onSelect: (theme: Theme) => void }) {
  return (
    <button
      onClick={() => onSelect(option.id)}
      className="group relative overflow-hidden rounded-2xl p-3 text-left transition-all duration-300"
      style={{
        background: active ? 'var(--bg-elevated)' : 'var(--bg-surface)',
        border: active ? '1px solid var(--border-accent)' : '1px solid var(--border-base)',
        boxShadow: active ? 'var(--shadow-card), var(--glow-accent-sm)' : 'var(--shadow-card)',
        transform: active ? 'translateY(-1px)' : 'none',
      }}
    >
      <div
        className="relative mb-3 h-28 overflow-hidden rounded-[18px]"
        style={{
          background: option.preview.shell,
          border: `1px solid ${option.preview.divider}`,
        }}
      >
        <div
          className="absolute inset-y-0 left-0"
          style={{
            width: '34%',
            background: option.preview.sidebar,
            borderRight: `1px solid ${option.preview.divider}`,
          }}
        />
        <div
          className="absolute left-3 top-3 h-6 rounded-xl"
          style={{
            width: 'calc(34% - 24px)',
            background: option.preview.sidebarSelection,
            boxShadow: `0 10px 22px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute right-4 top-4 h-4 rounded-full"
          style={{
            left: '40%',
            background: option.preview.search,
          }}
        />
        <div
          className="absolute right-10 top-11 h-9 rounded-2xl"
          style={{
            left: '40%',
            background: option.preview.panel,
            boxShadow: `0 14px 28px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute top-[53px] h-2 rounded-full"
          style={{
            left: '44%',
            width: '4rem',
            background: option.preview.text,
            opacity: 0.78,
          }}
        />
        <div
          className="absolute top-[66px] h-2 rounded-full"
          style={{
            left: '44%',
            width: '2.75rem',
            background: option.preview.textMuted,
            opacity: 0.55,
          }}
        />
        <div
          className="absolute bottom-4 right-4 h-9 w-9 rounded-2xl"
          style={{
            background: option.preview.accent,
            boxShadow: `0 12px 26px ${option.preview.accentGlow}`,
          }}
        />
        <div
          className="absolute bottom-6 h-2 rounded-full"
          style={{
            left: '52%',
            width: '3rem',
            background: option.preview.textMuted,
            opacity: 0.4,
          }}
        />
      </div>

      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
              style={{
                background: active ? 'var(--accent-glow)' : 'var(--bg-overlay)',
                color: active ? 'var(--accent-primary)' : 'var(--text-secondary)',
                border: '1px solid var(--border-base)',
              }}
            >
              {option.icon}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>{option.label}</p>
              <p className="text-[11px] uppercase tracking-[0.18em]" style={{ color: 'var(--text-muted)' }}>{option.tone}</p>
            </div>
          </div>
          <p className="mt-3 text-xs leading-5" style={{ color: 'var(--text-secondary)' }}>{option.description}</p>
        </div>

        <span
          className="mt-0.5 flex h-6 w-6 items-center justify-center rounded-full"
          style={{
            background: active ? 'var(--accent-primary)' : 'transparent',
            color: active ? '#ffffff' : 'var(--text-disabled)',
            border: active ? 'none' : '1px solid var(--border-base)',
          }}
        >
          {active ? <Check size={14} /> : <div className="h-2.5 w-2.5 rounded-full" style={{ background: 'var(--bg-overlay)' }} />}
        </span>
      </div>
    </button>
  )
}

function Toggle({ enabled, onToggle }: { enabled: boolean; onToggle: () => void }) {
  return (
    <div
      onClick={onToggle}
      className="w-9 h-5 rounded-full relative cursor-pointer transition-all duration-200"
      style={{
        background: enabled ? 'var(--accent-secondary)' : 'var(--bg-overlay)',
        boxShadow: enabled ? 'var(--glow-accent-sm)' : 'none',
        border: '1px solid var(--border-base)',
      }}
    >
      <div
        className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform duration-200 ${enabled ? 'translate-x-4' : 'translate-x-0.5'}`}
        style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.3)' }}
      />
    </div>
  )
}

export default function SettingsPage() {
  const { theme, setTheme } = useThemeContext()
  const { t, lang, setLang } = useLanguage()
  const [tab, setTab] = useState<Tab>('tools')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newTool, setNewTool] = useState({ name: '', pushDir: '' })
  const [newScanDirs, setNewScanDirs] = useState<Record<string, string>>({})
  const [appVersion, setAppVersion] = useState('')
  const [logDir, setLogDir] = useState('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updateResult, setUpdateResult] = useState<string | null>(null)

  const themeOptions: ThemeOption[] = [
    {
      id: 'dark',
      label: 'Dark',
      tone: 'Ink Slate',
      description: t('settings.themeDark'),
      icon: <Moon size={15} />,
      preview: {
        shell: 'radial-gradient(circle at top right, rgba(154,168,193,0.12), transparent 28%), linear-gradient(180deg, #13171d 0%, #0f1318 100%)',
        sidebar: 'rgba(20, 24, 31, 0.94)',
        sidebarSelection: 'rgba(167, 183, 207, 0.12)',
        search: 'rgba(255,255,255,0.06)',
        panel: 'rgba(29, 35, 44, 0.94)',
        accent: '#a7b7cf',
        accentGlow: 'rgba(116, 132, 159, 0.22)',
        text: '#edf1f7',
        textMuted: '#7e8a9c',
        divider: 'rgba(255,255,255,0.07)',
      },
    },
    {
      id: 'young',
      label: 'Young',
      tone: 'Breeze Paper',
      description: t('settings.themeYoung'),
      icon: <Sparkles size={15} />,
      preview: {
        shell: 'radial-gradient(circle at top right, rgba(147,197,253,0.18), transparent 28%), radial-gradient(circle at bottom left, rgba(251,191,36,0.08), transparent 30%), linear-gradient(180deg, #f7fbff 0%, #eef5fd 52%, #fffdf8 100%)',
        sidebar: 'rgba(236, 243, 251, 0.97)',
        sidebarSelection: 'rgba(93, 143, 214, 0.14)',
        search: 'rgba(104, 135, 178, 0.14)',
        panel: 'rgba(255, 255, 255, 0.99)',
        accent: '#5d8fd6',
        accentGlow: 'rgba(93, 143, 214, 0.18)',
        text: '#28415d',
        textMuted: '#8195aa',
        divider: 'rgba(96, 126, 171, 0.14)',
      },
    },
    {
      id: 'light',
      label: 'Light',
      tone: 'Messor Calm',
      description: t('settings.themeLight'),
      icon: <Sun size={15} />,
      preview: {
        shell: 'linear-gradient(180deg, #f7f8fb 0%, #eef1f6 100%)',
        sidebar: 'rgba(237, 239, 243, 0.96)',
        sidebarSelection: 'rgba(177, 193, 217, 0.34)',
        search: 'rgba(55, 65, 81, 0.08)',
        panel: 'rgba(255, 255, 255, 0.98)',
        accent: '#2d6df6',
        accentGlow: 'rgba(45, 109, 246, 0.20)',
        text: '#1f2937',
        textMuted: '#97a2b3',
        divider: 'rgba(15, 23, 42, 0.08)',
      },
    },
  ]

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders(), GetAppVersion(), GetLogDir()]).then(([c, p, v, logPath]) => {
      setCfg(c)
      setProviders(p ?? [])
      setAppVersion(v as string)
      setLogDir(logPath as string)
    })
  }, [])

  const checkUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateResult(null)
    try {
      const info = await CheckAppUpdateAndNotify()
      if (info.hasUpdate) {
        setUpdateResult(t('settings.updateFound', { version: info.latestVersion }))
      } else {
        setUpdateResult(t('settings.updateLatest', { version: info.currentVersion }))
      }
    } catch (e: any) {
      setUpdateResult(t('settings.updateFailed', { msg: e?.message ?? String(e) }))
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

  const pickDir = async (onPick: (path: string) => void, currentPath = '') => {
    const dir = await OpenFolderDialog(currentPath)
    if (dir) onPick(dir)
  }

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)
  const proxyMode: ProxyMode = (cfg?.proxy?.Mode as ProxyMode) || 'none'

  if (!cfg) return <div className="p-8" style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</div>

  return (
    <div className="p-8 max-w-2xl">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Settings size={18} /> {t('settings.title')}
        </h2>
        <div className="flex items-center gap-3">
          {updateResult && (
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{updateResult}</span>
          )}
          {appVersion && (
            <span className="text-xs font-mono" style={{ color: 'var(--text-muted)' }}>
              {appVersion === 'dev' ? 'dev' : appVersion.startsWith('v') ? appVersion : `v${appVersion}`}
            </span>
          )}
          <button
            onClick={checkUpdate}
            disabled={checkingUpdate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg transition-colors disabled:opacity-50"
            style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)', border: '1px solid var(--border-base)' }}
          >
            <RefreshCw size={12} className={checkingUpdate ? 'animate-spin' : ''} />
            {checkingUpdate ? t('settings.checkingUpdate') : t('settings.checkUpdate')}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div
        className="flex gap-1 mb-6 rounded-xl p-1 w-fit"
        style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
      >
        {(['tools', 'cloud', 'general', 'network'] as Tab[]).map(tabKey => {
          const labels: Record<Tab, string> = {
            tools: t('settings.tabTools'),
            cloud: t('settings.tabCloud'),
            general: t('settings.tabGeneral'),
            network: t('settings.tabNetwork'),
          }
          return (
            <button
              key={tabKey}
              onClick={() => setTab(tabKey)}
              className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
              style={tab === tabKey ? {
                background: 'var(--bg-overlay)',
                color: 'var(--text-primary)',
                boxShadow: 'var(--glow-accent-sm)',
                border: '1px solid var(--border-accent)',
              } : {
                color: 'var(--text-muted)',
                border: '1px solid transparent',
              }}
            >{labels[tabKey]}</button>
          )
        })}
      </div>

      {/* Tools tab */}
      {tab === 'tools' && (
        <div className="space-y-4">
          {(cfg.tools ?? []).map((tool: any) => (
            <div
              key={tool.name}
              className="rounded-xl p-4"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)' }}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <ToolIcon name={tool.name} size={28} />
                  <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>{tool.name}</span>
                </div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.toolEnabled')}</span>
                  <Toggle enabled={tool.enabled} onToggle={() => updateTool(tool.name, 'enabled', !tool.enabled)} />
                </label>
              </div>

              <div className="mb-3">
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>{t('settings.pushPath')}</p>
                <div className="flex gap-2">
                  <input
                    value={tool.pushDir ?? ''}
                    onChange={e => updateTool(tool.name, 'pushDir', e.target.value)}
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(dir => updateTool(tool.name, 'pushDir', dir), tool.pushDir ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title={t('settings.selectDir')}
                  >
                    <FolderOpen size={14} />
                  </button>
                </div>
              </div>

              <div>
                <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>{t('settings.scanPaths')}</p>
                <div className="space-y-2">
                  {(tool.scanDirs ?? []).map((dir: string, idx: number) => (
                    <div key={`${tool.name}-scan-${idx}`} className="flex gap-2">
                      <input
                        value={dir}
                        onChange={e => updateScanDir(tool.name, idx, e.target.value)}
                        className="input-base flex-1 font-mono"
                      />
                      <button
                        onClick={() => pickDir(d => updateScanDir(tool.name, idx, d), dir ?? '')}
                        className="btn-secondary px-2.5 rounded-lg"
                        title={t('settings.selectDir')}
                      >
                        <FolderOpen size={14} />
                      </button>
                      <button
                        onClick={() => removeScanDir(tool.name, idx)}
                        className="btn-secondary px-2.5 rounded-lg"
                        title={t('settings.deleteScanPath')}
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
                <div className="mt-2 flex gap-2">
                  <input
                    value={newScanDirs[tool.name] ?? ''}
                    onChange={e => setNewScanDirs(prev => ({ ...prev, [tool.name]: e.target.value }))}
                    placeholder="/path/to/scan"
                    className="input-base flex-1 font-mono"
                  />
                  <button
                    onClick={() => pickDir(d => setNewScanDirs(prev => ({ ...prev, [tool.name]: d })), newScanDirs[tool.name] ?? '')}
                    className="btn-secondary px-2.5 rounded-lg"
                    title={t('settings.selectDir')}
                  >
                    <FolderOpen size={14} />
                  </button>
                  <button
                    onClick={() => addScanDir(tool.name)}
                    className="btn-secondary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
                  >
                    <Plus size={14} /> {t('settings.addPath')}
                  </button>
                </div>
              </div>

              {tool.custom && (
                <button
                  onClick={async () => { await RemoveCustomTool(tool.name); const c = await GetConfig(); setCfg(c) }}
                  className="mt-2 text-xs flex items-center gap-1 transition-colors"
                  style={{ color: 'var(--color-error)' }}
                >
                  <Trash2 size={12} /> {t('settings.deleteTool')}
                </button>
              )}
            </div>
          ))}

          {/* Add custom tool */}
          <div
            className="rounded-xl p-4"
            style={{ border: '1px dashed var(--border-surface)', background: 'var(--bg-surface)' }}
          >
            <p className="text-sm mb-3" style={{ color: 'var(--text-muted)' }}>{t('settings.addCustomTool')}</p>
            <div className="flex gap-2 mb-2">
              <input
                value={newTool.name}
                onChange={e => setNewTool(p => ({ ...p, name: e.target.value }))}
                placeholder={t('settings.toolName')}
                className="input-base flex-1"
              />
            </div>
            <div className="flex gap-2">
              <input
                value={newTool.pushDir}
                onChange={e => setNewTool(p => ({ ...p, pushDir: e.target.value }))}
                placeholder="/path/to/push"
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setNewTool(p => ({ ...p, pushDir: d })), newTool.pushDir)}
                className="btn-secondary px-2.5 rounded-lg"
                title={t('settings.selectDir')}
              >
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
                className="btn-primary px-3 py-1.5 rounded-lg text-sm flex items-center gap-1"
              >
                <Plus size={14} /> {t('settings.addPath')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Cloud tab */}
      {tab === 'cloud' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.cloudProvider')}</p>
            <div className="flex gap-2">
              {providers.map((p: any) => (
                <button
                  key={p.name}
                  onClick={() => setCfg((prev: any) => ({ ...prev, cloud: { ...prev.cloud, provider: p.name } }))}
                  className="px-4 py-2 rounded-lg text-sm transition-all duration-200"
                  style={cfg.cloud?.provider === p.name ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {p.name}
                </button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              {cfg.cloud?.provider !== 'git' && (
                <div>
                  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.bucket')}</p>
                  <input
                    value={cfg.cloud?.bucketName ?? ''}
                    onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, bucketName: e.target.value } }))}
                    className="input-base"
                  />
                </div>
              )}
              {selectedProvider.fields.map((f: any) => (
                <div key={f.key}>
                  <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{f.label}</p>
                  <input
                    type={f.secret ? 'password' : 'text'}
                    placeholder={f.placeholder ?? ''}
                    value={cfg.cloud?.credentials?.[f.key] ?? ''}
                    onChange={e => setCfg((p: any) => ({
                      ...p, cloud: { ...p.cloud, credentials: { ...p.cloud?.credentials, [f.key]: e.target.value } }
                    }))}
                    className="input-base font-mono"
                  />
                </div>
              ))}
              <div>
                <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.syncInterval')}</p>
                <input
                  type="number"
                  min={0}
                  value={cfg.cloud?.syncIntervalMinutes ?? 0}
                  onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, syncIntervalMinutes: parseInt(e.target.value) || 0 } }))}
                  className="input-base w-32"
                />
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <Toggle
                  enabled={!!cfg.cloud?.enabled}
                  onToggle={() => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, enabled: !p.cloud?.enabled } }))}
                />
                <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{t('settings.enableAutoBackup')}</span>
              </label>
            </>
          )}
        </div>
      )}

      {/* General tab */}
      {tab === 'general' && (
        <div className="space-y-4">
          {/* Language */}
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.language')}</p>
            <div className="flex gap-2">
              {(['zh', 'en'] as const).map(l => (
                <button
                  key={l}
                  onClick={() => setLang(l)}
                  className="px-4 py-1.5 rounded-lg text-sm transition-all duration-200"
                  style={lang === l ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {l === 'zh' ? '中文' : 'English'}
                </button>
              ))}
            </div>
          </div>

          {/* Theme */}
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.theme')}</p>
            <div className="grid gap-3 md:grid-cols-3">
              {themeOptions.map((option) => (
                <ThemeOptionCard
                  key={option.id}
                  option={option}
                  active={theme === option.id}
                  onSelect={setTheme}
                />
              ))}
            </div>
            <p className="mt-2 text-xs leading-5" style={{ color: 'var(--text-muted)' }}>
              {t('settings.themeHint')}
            </p>
          </div>

          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.logLevel')}</p>
            <div className="flex gap-2 mb-2">
              {([
                ['debug', 'Debug'],
                ['info', 'Info'],
                ['error', 'Error'],
              ] as [string, string][]).map(([level, label]) => (
                <button
                  key={level}
                  onClick={() => setCfg((p: any) => ({ ...p, logLevel: level }))}
                  className="px-3 py-1.5 rounded-lg text-sm transition-all duration-200"
                  style={(cfg.logLevel ?? 'error') === level ? {
                    background: 'var(--accent-glow)',
                    color: 'var(--accent-primary)',
                    border: '1px solid var(--border-accent)',
                    boxShadow: 'var(--glow-accent-sm)',
                  } : {
                    background: 'var(--bg-elevated)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  {label}
                </button>
              ))}
            </div>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.logLevelHint')}</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.logDir')}</p>
            <div className="flex items-center gap-2">
              <button
                onClick={async () => { await OpenLogDir() }}
                className="btn-secondary px-3 py-1.5 rounded-lg text-sm"
              >
                {t('settings.openLogDir')}
              </button>
              <span className="text-xs font-mono break-all" style={{ color: 'var(--text-muted)' }}>{logDir}</span>
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.logDirHint')}</p>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.skillsDir')}</p>
            <div className="flex gap-2">
              <input
                value={cfg.skillsStorageDir ?? ''}
                onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
                className="input-base flex-1 font-mono"
              />
              <button
                onClick={() => pickDir(d => setCfg((p: any) => ({ ...p, skillsStorageDir: d })), cfg.skillsStorageDir ?? '')}
                className="btn-secondary px-2.5 rounded-lg"
                title={t('settings.selectDir')}
              >
                <FolderOpen size={16} />
              </button>
            </div>
          </div>
          <div>
            <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.defaultCategory')}</p>
            <div
              className="rounded-lg px-3 py-2 text-sm"
              style={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-base)', color: 'var(--text-secondary)' }}
            >
              Default
            </div>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>{t('settings.defaultCategoryHint')}</p>
          </div>
        </div>
      )}

      {/* Network tab */}
      {tab === 'network' && (
        <div className="space-y-6">
          <div>
            <p className="text-sm mb-1 flex items-center gap-1.5" style={{ color: 'var(--text-muted)' }}>
              <Globe size={14} /> {t('settings.proxy')}
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--text-muted)' }}>
              {t('settings.proxyHint')}
            </p>

            <div className="space-y-2">
              {([
                ['none',   t('settings.proxyNone'),   t('settings.proxyNoneDesc')],
                ['system', t('settings.proxySystem'), t('settings.proxySystemDesc')],
                ['manual', t('settings.proxyManual'), t('settings.proxyManualDesc')],
              ] as [ProxyMode, string, string][]).map(([mode, label, desc]) => (
                <div
                  key={mode}
                  onClick={() => setProxyMode(mode)}
                  className="flex items-start gap-3 p-3 rounded-xl cursor-pointer transition-all duration-200 select-none"
                  style={proxyMode === mode ? {
                    background: 'var(--accent-glow)',
                    border: '1px solid var(--border-accent)',
                  } : {
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-base)',
                  }}
                >
                  <div
                    className="mt-0.5 w-4 h-4 rounded-full border-2 flex items-center justify-center shrink-0 transition-all duration-200"
                    style={proxyMode === mode ? {
                      borderColor: 'var(--accent-secondary)',
                      background: 'var(--accent-secondary)',
                    } : {
                      borderColor: 'var(--text-muted)',
                    }}
                  >
                    {proxyMode === mode && <div className="w-1.5 h-1.5 bg-white rounded-full" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium leading-snug" style={{ color: 'var(--text-primary)' }}>{label}</p>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>{desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {proxyMode === 'manual' && (
            <div>
              <p className="text-sm mb-2" style={{ color: 'var(--text-muted)' }}>{t('settings.proxyUrl')}</p>
              <input
                value={cfg.proxy?.URL ?? ''}
                onChange={e => setProxyURL(e.target.value)}
                placeholder="http://127.0.0.1:7890"
                className="input-base font-mono"
              />
              <p className="mt-1.5 text-xs" style={{ color: 'var(--text-muted)' }}>
                {t('settings.proxyUrlHint')}
              </p>
            </div>
          )}
        </div>
      )}

      <button
        onClick={save}
        disabled={saving}
        className="btn-primary mt-8 px-6 py-2.5 rounded-lg text-sm"
      >
        {saving ? t('common.saving') : t('settings.saveSettings')}
      </button>
    </div>
  )
}
