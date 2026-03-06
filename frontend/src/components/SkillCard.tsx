import { useRef, useState } from 'react'
import { Github, FolderOpen, RefreshCw, FolderOpenDot, Copy, Check } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { OpenPath, ReadSkillFileContent } from '../../wailsjs/go/main/App'

interface Skill { id: string; name: string; category: string; source: 'github' | 'manual'; hasUpdate: boolean; path?: string }
interface Props {
  skill: Skill
  categories: string[]
  onDelete: () => void
  onUpdate?: () => void
  onMoveCategory: (category: string) => void
  selectMode?: boolean
  selected?: boolean
  onToggleSelect?: () => void
  onHoverStart?: (rect: DOMRect) => void
  onHoverEnd?: () => void
}

export default function SkillCard({
  skill, categories, onDelete, onUpdate, onMoveCategory,
  selectMode, selected, onToggleSelect,
  onHoverStart, onHoverEnd,
}: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null)
  const [copied, setCopied] = useState(false)
  const cardRef = useRef<HTMLDivElement>(null)

  const handleContextMenu = (e: React.MouseEvent) => {
    if (selectMode) return
    e.preventDefault()
    setMenu({ x: e.clientX, y: e.clientY })
  }

  const handleClick = () => {
    if (selectMode) onToggleSelect?.()
  }

  const handleMouseEnter = () => {
    if (selectMode) return
    if (cardRef.current) onHoverStart?.(cardRef.current.getBoundingClientRect())
  }

  const handleMouseLeave = () => {
    onHoverEnd?.()
  }

  const handleOpenFolder = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (skill.path) OpenPath(skill.path)
  }

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!skill.path) return
    try {
      const content = await ReadSkillFileContent(skill.path)
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  const menuItems = [
    ...(skill.hasUpdate ? [{ label: '更新', onClick: () => onUpdate?.() }] : []),
    ...categories.filter(c => c !== skill.category).map(c => ({
      label: `移动到 ${c}`,
      onClick: () => onMoveCategory(c),
    })),
    { label: '删除', onClick: onDelete, danger: true },
  ]

  return (
    <>
      <div
        ref={cardRef}
        draggable={!selectMode}
        onDragStart={e => !selectMode && e.dataTransfer.setData('skillId', skill.id)}
        onContextMenu={handleContextMenu}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`relative bg-gray-800 border rounded-xl p-4 transition-colors group ${
          selectMode ? 'cursor-pointer' : 'cursor-grab'
        } ${
          selected
            ? 'border-indigo-500 bg-indigo-900/20'
            : 'border-gray-700 hover:border-indigo-500'
        }`}
      >
        {selectMode && (
          <div className="absolute top-2 left-2 z-10">
            <div className={`w-4 h-4 rounded border-2 flex items-center justify-center ${
              selected ? 'bg-indigo-500 border-indigo-500' : 'border-gray-500 bg-gray-700'
            }`}>
              {selected && (
                <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              )}
            </div>
          </div>
        )}

        {/* Open folder button — top-right, visible on hover */}
        {!selectMode && skill.path && (
          <button
            onClick={handleOpenFolder}
            title="打开目录"
            className="absolute top-2 right-2 z-10 p-1 rounded text-gray-600 opacity-0 group-hover:opacity-100 hover:text-gray-200 hover:bg-gray-700 transition-all"
          >
            <FolderOpenDot size={14} />
          </button>
        )}

        {skill.hasUpdate && !selectMode && (
          <span className="absolute top-2 right-8 w-2.5 h-2.5 rounded-full bg-red-500" />
        )}
        {skill.hasUpdate && selectMode && (
          <span className="absolute top-2 right-2 w-2.5 h-2.5 rounded-full bg-red-500" />
        )}

        <div className={`flex items-center gap-2 mb-2 ${selectMode ? 'pl-5' : ''}`}>
          {skill.source === 'github'
            ? <Github size={14} className="text-gray-400" />
            : <FolderOpen size={14} className="text-gray-400" />}
          <span className={`text-xs px-1.5 py-0.5 rounded ${skill.source === 'github' ? 'bg-blue-900/50 text-blue-300' : 'text-gray-400'}`}>
            {skill.source}
          </span>
        </div>
        <p className={`font-medium text-sm truncate ${selectMode ? 'pl-5' : 'pr-5'}`}>{skill.name}</p>
        {!selectMode && (
          <div className="mt-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            {skill.hasUpdate && (
              <button onClick={e => { e.stopPropagation(); onUpdate?.() }} className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1">
                <RefreshCw size={12} /> 更新
              </button>
            )}
            {skill.path && (
              <button onClick={handleCopy} className="text-xs text-gray-400 hover:text-gray-200 flex items-center gap-1">
                {copied ? <><Check size={12} className="text-green-400" /> 已复制</> : <><Copy size={12} /> 复制</>}
              </button>
            )}
            <button onClick={e => { e.stopPropagation(); onDelete() }} className="text-xs text-red-400 hover:text-red-300 ml-auto">删除</button>
          </div>
        )}
      </div>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menuItems} onClose={() => setMenu(null)} />
      )}
    </>
  )
}
