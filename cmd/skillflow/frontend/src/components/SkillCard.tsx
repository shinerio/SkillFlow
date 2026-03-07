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
  dragging?: boolean
  dropTargetActive?: boolean
  onDragStateChange?: (dragging: boolean) => void
  selectMode?: boolean
  selected?: boolean
  onToggleSelect?: () => void
  onHoverStart?: (rect: DOMRect) => void
  onHoverEnd?: () => void
}

export default function SkillCard({
  skill, categories, onDelete, onUpdate, onMoveCategory,
  dragging = false, dropTargetActive = false, onDragStateChange,
  selectMode, selected, onToggleSelect,
  onHoverStart, onHoverEnd,
}: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null)
  const [copied, setCopied] = useState(false)
  const cardRef = useRef<HTMLDivElement>(null)
  const dragGhostRef = useRef<HTMLDivElement | null>(null)

  const setCardDragImage = (e: React.DragEvent) => {
    if (!cardRef.current) return
    const clone = cardRef.current.cloneNode(true) as HTMLDivElement
    const rect = cardRef.current.getBoundingClientRect()
    clone.style.width = `${Math.max(rect.width * 0.82, 180)}px`
    clone.style.transform = 'scale(0.82)'
    clone.style.transformOrigin = 'top left'
    clone.style.opacity = '0.96'
    clone.style.pointerEvents = 'none'
    clone.style.position = 'fixed'
    clone.style.top = '-1000px'
    clone.style.left = '-1000px'
    clone.style.zIndex = '9999'
    document.body.appendChild(clone)
    dragGhostRef.current = clone
    e.dataTransfer.setDragImage(clone, 24, 18)
  }

  const cleanupDragGhost = () => {
    if (dragGhostRef.current?.parentNode) {
      dragGhostRef.current.parentNode.removeChild(dragGhostRef.current)
    }
    dragGhostRef.current = null
  }

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
        onDragStart={e => {
          if (selectMode) return
          e.dataTransfer.setData('text/plain', skill.id)
          e.dataTransfer.setData('application/x-skillflow-skill-id', skill.id)
          e.dataTransfer.effectAllowed = 'move'
          setCardDragImage(e)
          onDragStateChange?.(true)
        }}
        onDragEnd={() => {
          cleanupDragGhost()
          onDragStateChange?.(false)
        }}
        onContextMenu={handleContextMenu}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`relative border rounded-xl transition-all duration-150 group ${
          selectMode ? 'cursor-pointer' : 'cursor-grab'
        } ${
          dragging && dropTargetActive ? 'bg-transparent border-transparent min-h-[88px]' :
          dragging ? 'bg-gray-800/50 border-indigo-400/50 p-4 scale-[0.96] opacity-55' :
          selected
            ? 'bg-gray-800 border-indigo-500 bg-indigo-900/20 p-4'
            : 'bg-gray-800 border-gray-700 hover:border-indigo-500 p-4'
        }`}
      >
        {dragging && dropTargetActive && (
          <div className="absolute inset-x-4 top-1/2 -translate-y-1/2 h-[2px] rounded-full bg-indigo-400 shadow-[0_0_0_1px_rgba(99,102,241,0.3)]" />
        )}

        <div className={`${dragging && dropTargetActive ? 'opacity-0 pointer-events-none' : 'opacity-100'} transition-opacity`}>
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
      </div>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menuItems} onClose={() => setMenu(null)} />
      )}
    </>
  )
}
