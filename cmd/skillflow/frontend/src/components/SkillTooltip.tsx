import { createPortal } from 'react-dom'
import { Github, FolderOpen, Tag, Wrench, GitBranch, Calendar, Clock, Hash, ExternalLink } from 'lucide-react'

export interface SkillInfo {
  Name: string
  Category?: string
  Source?: string
  SourceURL?: string
  SourceSubPath?: string
  SourceSHA?: string
  LatestSHA?: string
  InstalledAt?: string
  UpdatedAt?: string
}

interface SkillMeta {
  Name: string
  Description: string
  ArgumentHint: string
  AllowedTools: string
  Context: string
  DisableModelInvocation: boolean
}

interface Props {
  skill: SkillInfo
  meta: SkillMeta | null
  anchorRect: DOMRect
}

function fmt(dateStr: string | undefined): string {
  if (!dateStr) return '—'
  try {
    return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
  } catch {
    return '—'
  }
}

function shortSHA(sha: string): string {
  return sha ? sha.slice(0, 7) : '—'
}

export default function SkillTooltip({ skill, meta, anchorRect }: Props) {
  const TOOLTIP_WIDTH = 300
  const TOOLTIP_MAX_HEIGHT = 400
  const GAP = 8

  // Position: prefer right side, fall back to left
  let left = anchorRect.right + GAP
  if (left + TOOLTIP_WIDTH > window.innerWidth - 8) {
    left = anchorRect.left - TOOLTIP_WIDTH - GAP
  }

  // Align top with the card, shift up if it would overflow
  let top = anchorRect.top
  if (top + TOOLTIP_MAX_HEIGHT > window.innerHeight - 8) {
    top = window.innerHeight - TOOLTIP_MAX_HEIGHT - 8
  }

  const displayName = meta?.Name || skill.Name
  const isGitHub = skill.Source === 'github'

  const tooltip = (
    <div
      style={{ left, top, width: TOOLTIP_WIDTH, maxHeight: TOOLTIP_MAX_HEIGHT }}
      className="fixed z-50 overflow-y-auto bg-gray-900 border border-gray-700 rounded-xl shadow-2xl text-sm pointer-events-none"
    >
      {/* Header */}
      <div className="px-4 pt-4 pb-3 border-b border-gray-800">
        <div className="flex items-start gap-2">
          <div className="mt-0.5 shrink-0 text-gray-400">
            {isGitHub ? <Github size={14} /> : <FolderOpen size={14} />}
          </div>
          <div className="min-w-0 flex-1">
            <p className="font-semibold text-white leading-snug truncate">{displayName}</p>
            <div className="flex items-center gap-1.5 mt-1">
              {skill.Source && (
                <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${
                  isGitHub ? 'bg-blue-900/60 text-blue-300' : 'bg-gray-800 text-gray-400'
                }`}>
                  {skill.Source}
                </span>
              )}
              {skill.Category && (
                <span className="text-xs text-gray-500 truncate">{skill.Category}</span>
              )}
            </div>
          </div>
        </div>

        {/* Description */}
        {meta === null ? (
          <p className="mt-3 text-xs text-gray-500 italic">加载中…</p>
        ) : meta.Description ? (
          <p className="mt-3 text-xs text-gray-300 leading-relaxed">{meta.Description}</p>
        ) : (
          <p className="mt-3 text-xs text-gray-600 italic">暂无描述</p>
        )}
      </div>

      {/* Frontmatter fields */}
      {meta && (meta.ArgumentHint || meta.AllowedTools || meta.Context) && (
        <div className="px-4 py-3 border-b border-gray-800 space-y-2">
          {meta.ArgumentHint && (
            <Row icon={<Tag size={12} />} label="参数提示">
              <code className="text-xs bg-gray-800 px-1.5 py-0.5 rounded text-indigo-300 font-mono">
                {meta.ArgumentHint}
              </code>
            </Row>
          )}
          {meta.AllowedTools && (
            <Row icon={<Wrench size={12} />} label="允许工具">
              <span className="text-xs text-gray-300">{meta.AllowedTools}</span>
            </Row>
          )}
          {meta.Context && (
            <Row icon={<GitBranch size={12} />} label="运行上下文">
              <span className="text-xs text-indigo-300 font-mono">{meta.Context}</span>
            </Row>
          )}
        </div>
      )}

      {/* Skill metadata */}
      {(isGitHub && skill.SourceURL || skill.SourceSHA || skill.InstalledAt) && (
        <div className="px-4 py-3 space-y-2">
          {isGitHub && skill.SourceURL && (
            <Row icon={<ExternalLink size={12} />} label="仓库">
              <span className="text-xs text-indigo-400 truncate max-w-[160px]">
                {skill.SourceURL.replace('https://github.com/', '')}
                {skill.SourceSubPath ? `/${skill.SourceSubPath}` : ''}
              </span>
            </Row>
          )}
          {skill.SourceSHA && (
            <Row icon={<Hash size={12} />} label="版本">
              <code className="text-xs font-mono text-gray-300">{shortSHA(skill.SourceSHA)}</code>
              {skill.LatestSHA && skill.LatestSHA !== skill.SourceSHA && (
                <span className="ml-2 text-xs text-amber-400">可更新 → {shortSHA(skill.LatestSHA)}</span>
              )}
            </Row>
          )}
          {skill.InstalledAt && (
            <Row icon={<Calendar size={12} />} label="安装时间">
              <span className="text-xs text-gray-400">{fmt(skill.InstalledAt)}</span>
            </Row>
          )}
          {skill.UpdatedAt && skill.UpdatedAt !== skill.InstalledAt && (
            <Row icon={<Clock size={12} />} label="更新时间">
              <span className="text-xs text-gray-400">{fmt(skill.UpdatedAt)}</span>
            </Row>
          )}
        </div>
      )}
    </div>
  )

  return createPortal(tooltip, document.body)
}

function Row({ icon, label, children }: { icon: React.ReactNode; label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2">
      <span className="text-gray-500 mt-0.5 shrink-0">{icon}</span>
      <span className="text-gray-500 shrink-0 w-16 text-xs leading-relaxed">{label}</span>
      <div className="flex items-center gap-1 min-w-0 flex-wrap">{children}</div>
    </div>
  )
}
