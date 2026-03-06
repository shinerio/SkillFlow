import { useRef, useState } from 'react'
import { FolderOpen, Github, FolderOpenDot, Copy, Check } from 'lucide-react'
import { OpenPath, GetSkillMeta, GetSkillMetaByPath, ReadSkillFileContent } from '../../wailsjs/go/main/App'
import SkillTooltip from './SkillTooltip'

interface Props {
  name: string
  subtitle?: string      // e.g. category or path hint
  source?: string        // 'github' | 'manual' | undefined
  path?: string          // filesystem path to open in file manager / fetch meta from
  id?: string            // if provided, use GetSkillMeta(id); else GetSkillMetaByPath(path)
  showSelection?: boolean  // default true; set false to hide selection checkbox (e.g. StarredRepos non-select mode)
  imported?: boolean     // show "已导入" badge
  selected: boolean
  onToggle: () => void
}

export default function SyncSkillCard({
  name, subtitle, source, path, id,
  showSelection = true, imported, selected, onToggle,
}: Props) {
  const cardRef = useRef<HTMLDivElement>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [hoveredRect, setHoveredRect] = useState<DOMRect | null>(null)
  const [meta, setMeta] = useState<any | null>(null)
  const [copied, setCopied] = useState(false)

  const handleMouseEnter = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      if (!cardRef.current) return
      setHoveredRect(cardRef.current.getBoundingClientRect())
      setMeta(null)
      try {
        let m: any
        if (id) {
          m = await GetSkillMeta(id)
        } else if (path) {
          m = await GetSkillMetaByPath(path)
        }
        setMeta(m ?? {})
      } catch {
        setMeta({})
      }
    }, 300)
  }

  const handleMouseLeave = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredRect(null)
    setMeta(null)
  }

  const handleOpen = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (path) OpenPath(path)
  }

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!path) return
    try {
      const content = await ReadSkillFileContent(path)
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  const skillInfo = { Name: name, Source: source }

  return (
    <>
      <div
        ref={cardRef}
        onClick={onToggle}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`relative flex flex-col gap-2 p-3 rounded-xl border cursor-pointer transition-colors select-none group ${
          selected
            ? 'bg-indigo-900/30 border-indigo-500'
            : 'bg-gray-800 border-gray-700 hover:border-indigo-500'
        }`}
      >
        {/* Top-right action buttons — visible on hover */}
        <div className="absolute top-2 right-2 flex items-center gap-0.5 z-10">
          {path && (
            <button
              onClick={handleCopy}
              title="复制 skill.md"
              className="p-1 rounded text-gray-500 opacity-0 group-hover:opacity-100 hover:text-gray-200 hover:bg-gray-700 transition-all"
            >
              {copied
                ? <Check size={12} className="text-green-400" />
                : <Copy size={12} />}
            </button>
          )}
          {path && (
            <button
              onClick={handleOpen}
              title="打开目录"
              className="p-1 rounded text-gray-500 opacity-0 group-hover:opacity-100 hover:text-gray-200 hover:bg-gray-700 transition-all"
            >
              <FolderOpenDot size={13} />
            </button>
          )}
        </div>

        {/* Source badge + imported badge */}
        <div className="flex items-center gap-1.5 pr-14 flex-wrap">
          {source === 'github'
            ? <Github size={12} className="text-gray-400 shrink-0" />
            : <FolderOpen size={12} className="text-gray-400 shrink-0" />}
          {source && (
            <span className={`text-xs px-1.5 py-0.5 rounded max-w-[72px] truncate ${
              source === 'github' ? 'bg-blue-900/50 text-blue-300' : 'bg-gray-700 text-gray-400'
            }`} title={source}>{source}</span>
          )}
          {imported && (
            <span className="text-xs bg-green-900/50 text-green-300 px-1.5 py-0.5 rounded">已导入</span>
          )}
        </div>

        {/* Skill name */}
        <p className="text-sm font-medium leading-snug truncate pr-5">{name}</p>

        {/* Subtitle (category or repo name) */}
        {subtitle && (
          <p className="text-xs text-gray-500 truncate">{subtitle}</p>
        )}

        {/* Selection indicator */}
        {showSelection && (
          <div className={`absolute bottom-2 right-2 w-4 h-4 rounded border-2 flex items-center justify-center transition-colors ${
            selected ? 'bg-indigo-500 border-indigo-500' : 'border-gray-600 bg-gray-700'
          }`}>
            {selected && (
              <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            )}
          </div>
        )}
      </div>

      {hoveredRect && (
        <SkillTooltip skill={skillInfo} meta={meta} anchorRect={hoveredRect} />
      )}
    </>
  )
}
