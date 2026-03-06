import { useState } from 'react'
import { Plus } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { CreateCategory, RenameCategory, DeleteCategory } from '../../wailsjs/go/main/App'

interface Props {
  categories: string[]
  selected: string | null
  onSelect: (cat: string | null) => void
  onDrop: (skillId: string, category: string) => void
  onRefresh: () => void
}

const defaultCategoryName = 'Default'

export default function CategoryPanel({ categories, selected, onSelect, onDrop, onRefresh }: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number; cat: string } | null>(null)
  const [renaming, setRenaming] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createName, setCreateName] = useState('')

  const handleDrop = (e: React.DragEvent, cat: string) => {
    e.preventDefault()
    const id = e.dataTransfer.getData('skillId')
    if (id) onDrop(id, cat)
  }

  return (
    <div className="w-48 flex-shrink-0 border-r border-gray-800 p-3 flex flex-col gap-0.5">
      {/* All */}
      <div
        onClick={() => onSelect(null)}
        onDragOver={e => e.preventDefault()}
        onDrop={e => handleDrop(e, '')}
        className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${selected === null ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'}`}
      >全部</div>

      {/* Categories */}
      {categories.map(cat => (
        renaming === cat
          ? <input
              key={cat} autoFocus value={newName}
              onChange={e => setNewName(e.target.value)}
              onBlur={async () => {
                if (newName && newName !== cat) { await RenameCategory(cat, newName); onRefresh() }
                setRenaming(null)
              }}
              onKeyDown={async e => {
                if (e.key === 'Enter') { await RenameCategory(cat, newName); onRefresh(); setRenaming(null) }
                if (e.key === 'Escape') setRenaming(null)
              }}
              className="px-3 py-1.5 rounded-lg text-sm bg-gray-700 text-white outline-none w-full"
            />
          : <div
              key={cat}
              onClick={() => onSelect(cat)}
              onDragOver={e => e.preventDefault()}
              onDrop={e => handleDrop(e, cat)}
              onContextMenu={e => {
                e.preventDefault()
                if (cat === defaultCategoryName) return
                setMenu({ x: e.clientX, y: e.clientY, cat })
              }}
              className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${selected === cat ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'}`}
            >{cat}</div>
      ))}

      {/* New category input */}
      {creating
        ? <input
            autoFocus value={createName}
            onChange={e => setCreateName(e.target.value)}
            onBlur={async () => {
              if (createName) { await CreateCategory(createName); onRefresh() }
              setCreating(false); setCreateName('')
            }}
            onKeyDown={async e => {
              if (e.key === 'Enter') { await CreateCategory(createName); onRefresh(); setCreating(false); setCreateName('') }
              if (e.key === 'Escape') { setCreating(false); setCreateName('') }
            }}
            className="px-3 py-1.5 rounded-lg text-sm bg-gray-700 text-white outline-none w-full"
          />
        : <button
            onClick={() => setCreating(true)}
            className="flex items-center gap-1.5 px-3 py-2 text-sm text-gray-500 hover:text-gray-300 mt-1"
          ><Plus size={14} /> 新建分类</button>
      }

      {/* Context menu */}
      {menu && (
        <ContextMenu
          x={menu.x} y={menu.y}
          items={[
            { label: '重命名', onClick: () => { setRenaming(menu.cat); setNewName(menu.cat) } },
            { label: '删除', onClick: async () => { await DeleteCategory(menu.cat); onRefresh() }, danger: true },
          ]}
          onClose={() => setMenu(null)}
        />
      )}
    </div>
  )
}
