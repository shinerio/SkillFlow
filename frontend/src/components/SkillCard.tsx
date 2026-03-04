import { useState } from 'react'
import { Github, FolderOpen, RefreshCw } from 'lucide-react'
import ContextMenu from './ContextMenu'

interface Skill { id: string; name: string; category: string; source: 'github' | 'manual'; hasUpdate: boolean }
interface Props {
  skill: Skill
  categories: string[]
  onDelete: () => void
  onUpdate?: () => void
  onMoveCategory: (category: string) => void
}

export default function SkillCard({ skill, categories, onDelete, onUpdate, onMoveCategory }: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null)

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault()
    setMenu({ x: e.clientX, y: e.clientY })
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
        draggable
        onDragStart={e => e.dataTransfer.setData('skillId', skill.id)}
        onContextMenu={handleContextMenu}
        className="relative bg-gray-800 border border-gray-700 rounded-xl p-4 cursor-grab hover:border-indigo-500 transition-colors group"
      >
        {skill.hasUpdate && (
          <span className="absolute top-2 right-2 w-2.5 h-2.5 rounded-full bg-red-500" />
        )}
        <div className="flex items-center gap-2 mb-2">
          {skill.source === 'github'
            ? <Github size={14} className="text-gray-400" />
            : <FolderOpen size={14} className="text-gray-400" />}
          <span className={`text-xs px-1.5 py-0.5 rounded ${skill.source === 'github' ? 'bg-blue-900/50 text-blue-300' : 'text-gray-400'}`}>
            {skill.source}
          </span>
        </div>
        <p className="font-medium text-sm truncate">{skill.name}</p>
        <div className="mt-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
          {skill.hasUpdate && (
            <button onClick={onUpdate} className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1">
              <RefreshCw size={12} /> 更新
            </button>
          )}
          <button onClick={onDelete} className="text-xs text-red-400 hover:text-red-300 ml-auto">删除</button>
        </div>
      </div>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menuItems} onClose={() => setMenu(null)} />
      )}
    </>
  )
}
