import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes, NavLink, useLocation } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star, X, Download, RefreshCw, AlertTriangle, GitMerge, MessageSquareWarning, ExternalLink, Wrench, Palette, Languages } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import StarredRepos from './pages/StarredRepos'
import ToolSkills from './pages/ToolSkills'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DownloadAppUpdate, ApplyAppUpdate, GetGitConflictPending, ResolveGitConflict, OpenURL, SetSkippedUpdateVersion } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import { ThemeProvider, useThemeContext } from './contexts/ThemeContext'
import { LanguageProvider, useLanguage } from './contexts/LanguageContext'
import type { Translations } from './i18n'
import { THEME_LABELS, getNextTheme } from './hooks/useTheme'
import AnimatedDialog from './components/ui/AnimatedDialog'
import { pageVariants } from './lib/motionVariants'

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

function parseAppUpdatePayload(data: unknown): main.AppUpdateInfo {
  return main.AppUpdateInfo.createFrom(data)
}

function AppContent() {
  const { t, lang, setLang } = useLanguage()
  const { theme, cycleTheme } = useThemeContext()
  const [dialogState, setDialogState] = useState<UpdateDialogState>('idle')
  const [updateInfo, setUpdateInfo] = useState<main.AppUpdateInfo | null>(null)

  const [conflictOpen, setConflictOpen] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<GitConflictInfo>({ message: '', files: [] })
  const [resolving, setResolving] = useState(false)
  const [resolveError, setResolveError] = useState('')
  const nextTheme = getNextTheme(theme)

  const handleResolve = async (useLocal: boolean) => {
    setResolving(true)
    setResolveError('')
    try {
      await ResolveGitConflict(useLocal)
      setConflictOpen(false)
    } catch (e: any) {
      setResolveError(String(e?.message ?? e ?? t('common.confirm')))
    } finally {
      setResolving(false)
    }
  }

  useEffect(() => {
    EventsOn('app.update.available', (data: unknown) => {
      setUpdateInfo(parseAppUpdatePayload(data))
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
    const releaseURL = updateInfo?.releaseUrl || 'https://github.com/shinerio/SkillFlow/releases/latest'
    if (releaseURL) {
      OpenURL(releaseURL)
    }
    setDialogState('idle')
  }

  return (
    <div
      className="flex h-screen flex-col relative"
      style={{ background: 'var(--app-shell)', color: 'var(--text-primary)' }}
    >
      {/* Git conflict dialog */}
      <AnimatedDialog open={conflictOpen} width="w-[420px]" zIndex={50}>
        <div className="flex items-center gap-2 mb-3">
          <AlertTriangle size={18} style={{ color: 'var(--color-warning)' }} />
          <span className="font-semibold text-base">{t('conflict.title')}</span>
        </div>
        <p className="text-sm mb-2" style={{ color: 'var(--text-secondary)' }}>
          {t('conflict.desc')}
        </p>
        {conflictInfo.files.length > 0 && (
          <div className="mb-3">
            <p className="text-xs mb-1.5" style={{ color: 'var(--text-muted)' }}>{t('conflict.filesLabel', { count: conflictInfo.files.length })}</p>
            <div
              className="max-h-28 overflow-y-auto rounded-lg px-2 py-1.5"
              style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
            >
              {conflictInfo.files.slice(0, 30).map((f, i) => (
                <div key={`${f}-${i}`} className="font-mono text-[11px] truncate" style={{ color: 'var(--text-secondary)' }}>{f}</div>
              ))}
              {conflictInfo.files.length > 30 && (
                <div className="text-[11px]" style={{ color: 'var(--text-muted)' }}>{t('conflict.moreFiles', { count: conflictInfo.files.length - 30 })}</div>
              )}
            </div>
          </div>
        )}
        {conflictInfo.message && (
          <div
            className="mb-3 rounded-lg px-2 py-1.5"
            style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
          >
            <p className="text-[11px] mb-1" style={{ color: 'var(--text-muted)' }}>{t('conflict.output')}</p>
            <pre className="text-[11px] whitespace-pre-wrap break-all max-h-20 overflow-y-auto" style={{ color: 'var(--text-secondary)' }}>{conflictInfo.message}</pre>
          </div>
        )}
        <ul className="text-xs list-disc list-inside mb-6 space-y-1" style={{ color: 'var(--text-muted)' }}>
          <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>{t('conflict.keepLocal')}</span> — {t('conflict.keepLocalDesc')}</li>
          <li><span className="font-medium" style={{ color: 'var(--text-primary)' }}>{t('conflict.keepRemote')}</span> — {t('conflict.keepRemoteDesc')}</li>
        </ul>
        {resolveError && (
          <p
            className="mb-3 text-xs rounded-lg px-3 py-2 break-all"
            style={{ color: 'var(--color-error)', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)' }}
          >{resolveError}</p>
        )}
        <div className="flex gap-3 justify-end">
          <button
            onClick={() => handleResolve(false)}
            disabled={resolving}
            className="btn-secondary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg"
          >
            {resolving ? <RefreshCw size={13} className="animate-spin" /> : <Download size={13} />}
            {t('conflict.keepRemote')}
          </button>
          <button
            onClick={() => handleResolve(true)}
            disabled={resolving}
            className="btn-primary flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg"
          >
            {resolving ? <RefreshCw size={13} className="animate-spin" /> : <GitMerge size={13} />}
            {t('conflict.keepLocal')}
          </button>
        </div>
      </AnimatedDialog>

      {/* Update dialog */}
      <AnimatedDialog open={dialogState !== 'idle'} width="w-[440px]" zIndex={50}>
        <UpdateDialogContent
          state={dialogState}
          info={updateInfo}
          onDownload={handleDownload}
          onRestart={handleRestart}
          onOpenRelease={handleOpenRelease}
          onSkip={handleSkip}
          onClose={() => setDialogState('idle')}
          t={t}
        />
      </AnimatedDialog>

      <div className="flex flex-1 overflow-hidden relative">
        {/* Sidebar */}
        <aside
          className="w-56 flex flex-col p-4 gap-1 relative"
          style={{
            background: 'var(--bg-surface)',
            borderRight: '1px solid var(--sidebar-border)',
          }}
        >
          {/* Top glow divider */}
          <div
            className="absolute top-0 left-0 right-0 h-px"
            style={{ background: 'linear-gradient(90deg, transparent, var(--shell-divider), transparent)', opacity: 0.9 }}
          />
          <div className="flex items-center justify-between mb-6 px-2">
            <h1
              className="text-[17px] font-semibold tracking-[0.08em]"
              style={{ color: 'var(--brand-color)', textShadow: 'var(--brand-shadow)' }}
            >
              SkillFlow
            </h1>
            <div className="flex items-center gap-1">
              <button
                onClick={() => setLang(lang === 'zh' ? 'en' : 'zh')}
                className="p-1.5 rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                title={t('nav.switchLangTitle')}
              >
                <Languages size={14} />
              </button>
              <button
                onClick={cycleTheme}
                className="p-1.5 rounded-lg transition-colors"
                style={{ color: 'var(--text-muted)' }}
                title={t('nav.switchThemeTitle', { theme: THEME_LABELS[nextTheme] })}
              >
                <Palette size={14} />
              </button>
            </div>
          </div>
          <NavItem to="/" icon={<Package size={16} />} label={t('nav.mySkills')} />
          <NavItem to="/tools" icon={<Wrench size={16} />} label={t('nav.myTools')} end={false} />
          <p className="text-xs px-2 mt-3 mb-1" style={{ color: 'var(--text-muted)' }}>{t('nav.syncSection')}</p>
          <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label={t('nav.pushToTool')} />
          <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label={t('nav.pullFromTool')} />
          <NavItem to="/starred" icon={<Star size={16} />} label={t('nav.starred')} end={false} />
          <div className="flex-1" />
          <div className="flex flex-col gap-1">
            <NavItem to="/backup" icon={<Cloud size={16} />} label={t('nav.backup')} />
            <NavItem to="/settings" icon={<Settings size={16} />} label={t('nav.settings')} />
            <button
              onClick={() => OpenURL(feedbackIssueURL)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={e => {
                e.currentTarget.style.backgroundColor = 'var(--bg-hover)'
                e.currentTarget.style.color = 'var(--text-primary)'
              }}
              onMouseLeave={e => {
                e.currentTarget.style.backgroundColor = ''
                e.currentTarget.style.color = 'var(--text-muted)'
              }}
            >
              <MessageSquareWarning size={16} />
              {t('nav.feedback')}
            </button>
          </div>
        </aside>

        {/* Main content with page transitions */}
        <main className="flex-1 overflow-auto relative">
          <AnimatedRoutes />
        </main>
      </div>
    </div>
  )
}

function AnimatedRoutes() {
  const location = useLocation()
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        variants={pageVariants}
        initial="initial"
        animate="animate"
        exit="exit"
        className="h-full"
      >
        <Routes location={location}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/sync/push" element={<SyncPush />} />
          <Route path="/sync/pull" element={<SyncPull />} />
          <Route path="/backup" element={<Backup />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/starred" element={<StarredRepos />} />
          <Route path="/starred/:repoEncoded" element={<StarredRepos />} />
          <Route path="/tools" element={<ToolSkills />} />
        </Routes>
      </motion.div>
    </AnimatePresence>
  )
}

export default function App() {
  return (
    <LanguageProvider>
      <ThemeProvider>
        <BrowserRouter>
          <AppContent />
        </BrowserRouter>
      </ThemeProvider>
    </LanguageProvider>
  )
}

interface UpdateDialogContentProps {
  state: UpdateDialogState
  info: main.AppUpdateInfo | null
  onDownload: () => void
  onRestart: () => void
  onOpenRelease: () => void
  onSkip: () => void
  onClose: () => void
  t: (key: keyof Translations, vars?: Record<string, string | number>) => string
}

function UpdateDialogContent({ state, info, onDownload, onRestart, onOpenRelease, onSkip, onClose, t }: UpdateDialogContentProps) {
  const isDownloading = state === 'downloading'

  return (
    <>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Download size={18} style={{ color: 'var(--accent-primary)' }} />
          <span className="font-semibold text-base">
            {state === 'ready_to_restart' ? t('update.ready') : state === 'download_failed' ? t('update.failed') : t('update.newVersion')}
          </span>
        </div>
        {!isDownloading && (
          <button
            onClick={onClose}
            style={{ color: 'var(--text-muted)' }}
            className="hover:opacity-80 transition-opacity"
          >
            <X size={16} />
          </button>
        )}
      </div>

      {(state === 'available' || state === 'downloading') && (
        <>
          <p className="text-sm mb-1" style={{ color: 'var(--text-secondary)' }}>
            {t('update.latestLabel')}<span className="font-mono font-medium" style={{ color: 'var(--accent-primary)' }}>{info?.latestVersion}</span>
          </p>
          <p className="text-sm mb-4" style={{ color: 'var(--text-muted)' }}>
            {t('update.currentLabel')}<span className="font-mono">{info?.currentVersion}</span>
          </p>
          {info?.releaseNotes && (
            <div
              className="mb-4 rounded-lg px-3 py-2 max-h-32 overflow-y-auto"
              style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-base)' }}
            >
              <p className="text-[11px] mb-1" style={{ color: 'var(--text-muted)' }}>{t('update.notes')}</p>
              <pre className="text-xs whitespace-pre-wrap break-all" style={{ color: 'var(--text-secondary)' }}>{info.releaseNotes}</pre>
            </div>
          )}
        </>
      )}

      {state === 'downloading' && (
        <div className="flex items-center gap-2 mb-4 text-sm" style={{ color: 'var(--text-secondary)' }}>
          <RefreshCw size={14} className="animate-spin" style={{ color: 'var(--accent-primary)' }} />
          <span>{t('update.downloading', { version: info?.latestVersion ?? '' })}</span>
        </div>
      )}

      {state === 'ready_to_restart' && (
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          {t('update.restartDesc')}
        </p>
      )}

      {state === 'download_failed' && (
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          {t('update.downloadFailDesc')}
        </p>
      )}

      {state === 'available' && (
        <div className="flex flex-col gap-2">
          {info?.canAutoUpdate && (
            <button
              onClick={onDownload}
              className="btn-primary flex items-center justify-center gap-2 w-full px-4 py-2.5 rounded-xl text-sm font-medium"
            >
              <Download size={14} />
              {t('update.download')}
            </button>
          )}
          <button
            onClick={onOpenRelease}
            className="btn-secondary flex items-center justify-center gap-2 w-full px-4 py-2.5 rounded-xl text-sm"
          >
            <ExternalLink size={14} />
            {t('update.openRelease')}
          </button>
          <button
            onClick={onSkip}
            className="flex items-center justify-center gap-2 w-full px-4 py-2 text-sm transition-colors"
            style={{ color: 'var(--text-muted)' }}
          >
            {t('update.skip')}
          </button>
        </div>
      )}

      {state === 'downloading' && (
        <p className="text-xs text-center" style={{ color: 'var(--text-muted)' }}>{t('update.downloadHint')}</p>
      )}

      {state === 'ready_to_restart' && (
        <div className="flex gap-3 justify-end">
          <button onClick={onClose} className="btn-secondary px-4 py-2 text-sm rounded-xl">
            {t('update.restartLater')}
          </button>
          <button onClick={onRestart} className="btn-primary flex items-center gap-2 px-4 py-2 text-sm rounded-xl">
            <RefreshCw size={13} />
            {t('update.restartNow')}
          </button>
        </div>
      )}

      {state === 'download_failed' && (
        <div className="flex gap-3 justify-end">
          <button onClick={onClose} className="btn-secondary px-4 py-2 text-sm rounded-xl">
            {t('common.close')}
          </button>
          <button onClick={onOpenRelease} className="btn-primary flex items-center gap-2 px-4 py-2 text-sm rounded-xl">
            <ExternalLink size={13} />
            {t('update.goDownload')}
          </button>
        </div>
      )}
    </>
  )
}

function NavItem({ to, icon, label, end = true }: { to: string; icon: React.ReactNode; label: string; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200 ${
          isActive ? 'nav-active' : 'nav-inactive'
        }`
      }
      style={({ isActive }) => isActive ? {
        backgroundColor: 'var(--active-surface)',
        color: 'var(--active-text)',
        border: '1px solid var(--active-border)',
        boxShadow: 'var(--active-shadow)',
      } : {
        color: 'var(--text-muted)',
        border: '1px solid transparent',
      }}
      onMouseEnter={e => {
        const el = e.currentTarget
        if (!el.classList.contains('nav-active')) {
          el.style.backgroundColor = 'var(--bg-hover)'
          el.style.color = 'var(--text-primary)'
        }
      }}
      onMouseLeave={e => {
        const el = e.currentTarget
        if (!el.classList.contains('nav-active')) {
          el.style.backgroundColor = ''
          el.style.color = 'var(--text-muted)'
        }
      }}
    >
      {icon}
      {label}
    </NavLink>
  )
}
