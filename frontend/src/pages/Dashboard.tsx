import { useEffect, useRef, useState, useCallback } from 'react'
import {
  ListSkills, ListCategories, MoveSkillCategory,
  DeleteSkill, DeleteSkills, ImportLocal, UpdateSkill, CheckUpdates,
  OpenFolderDialog, GetSkillMeta,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'
import SkillTooltip from '../components/SkillTooltip'
import GitHubInstallDialog from '../components/GitHubInstallDialog'
import { Github, FolderOpen, RefreshCw, Search, Trash2, CheckSquare } from 'lucide-react'

export default function Dashboard() {
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showGitHub, setShowGitHub] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedIDs, setSelectedIDs] = useState<Set<string>>(new Set())

  // Hover tooltip state
  const [hoveredSkill, setHoveredSkill] = useState<{ skill: any; rect: DOMRect } | null>(null)
  const [hoveredMeta, setHoveredMeta] = useState<any | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const load = useCallback(async () => {
    const [s, c] = await Promise.all([ListSkills(), ListCategories()])
    setSkills(s ?? [])
    setCategories(c ?? [])
  }, [])

  useEffect(() => {
    load()
    EventsOn('update.available', load)
  }, [load])

  const filtered = skills.filter(sk => {
    const matchCat = selectedCat === null || sk.Category === selectedCat
    const matchSearch = !search || sk.Name.toLowerCase().includes(search.toLowerCase())
    return matchCat && matchSearch
  })

  const handleDrop = async (skillId: string, category: string) => {
    await MoveSkillCategory(skillId, category)
    load()
  }

  const handleWindowDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }
  const handleWindowDragLeave = () => setDragOver(false)
  const handleWindowDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const items = Array.from(e.dataTransfer.items)
    for (const item of items) {
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        const file = item.getAsFile()
        if (file) {
          // @ts-ignore — Wails provides .path on File objects
          await ImportLocal(file.path ?? file.name, selectedCat ?? '')
          load()
        }
      }
    }
  }

  const handleImportButton = async () => {
    const dir = await OpenFolderDialog()
    if (dir) { await ImportLocal(dir, selectedCat ?? ''); load() }
  }

  const toggleSelectMode = () => {
    setSelectMode(prev => !prev)
    setSelectedIDs(new Set())
    clearHover()
  }

  const toggleSelectID = (id: string) => {
    setSelectedIDs(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleSelectAll = () => {
    if (selectedIDs.size === filtered.length) {
      setSelectedIDs(new Set())
    } else {
      setSelectedIDs(new Set(filtered.map(sk => sk.ID)))
    }
  }

  const handleBatchDelete = async () => {
    if (selectedIDs.size === 0) return
    await DeleteSkills(Array.from(selectedIDs))
    setSelectedIDs(new Set())
    setSelectMode(false)
    load()
  }

  const allSelected = filtered.length > 0 && selectedIDs.size === filtered.length

  // Hover handlers
  const clearHover = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    setHoveredSkill(null)
    setHoveredMeta(null)
  }

  const handleHoverStart = (sk: any, rect: DOMRect) => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current)
    hoverTimer.current = setTimeout(async () => {
      setHoveredSkill({ skill: sk, rect })
      setHoveredMeta(null)
      const meta = await GetSkillMeta(sk.ID)
      setHoveredMeta(meta)
    }, 300)
  }

  const handleHoverEnd = () => {
    clearHover()
  }

  return (
    <div
      className={`flex h-full relative ${dragOver ? 'ring-2 ring-inset ring-indigo-500' : ''}`}
      onDragOver={handleWindowDragOver}
      onDragLeave={handleWindowDragLeave}
      onDrop={handleWindowDrop}
    >
      {dragOver && (
        <div className="absolute inset-0 bg-indigo-500/10 flex items-center justify-center z-40 pointer-events-none">
          <p className="text-indigo-300 text-lg font-medium">松开以导入 Skill</p>
        </div>
      )}

      <CategoryPanel
        categories={categories}
        selected={selectedCat}
        onSelect={setSelectedCat}
        onDrop={handleDrop}
        onRefresh={load}
      />

      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-3 px-6 py-4 border-b border-gray-800">
          <div className="relative flex-1 max-w-xs">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input
              value={search} onChange={e => setSearch(e.target.value)}
              placeholder="搜索 Skills..."
              className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-8 pr-3 py-1.5 text-sm outline-none focus:border-indigo-500"
            />
          </div>

          {selectMode ? (
            <>
              <button
                onClick={toggleSelectAll}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              >
                <CheckSquare size={14} />
                {allSelected ? '取消全选' : '全选'}
              </button>
              <button
                onClick={handleBatchDelete}
                disabled={selectedIDs.size === 0}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-red-600 hover:bg-red-500 disabled:opacity-40 disabled:cursor-not-allowed rounded-lg"
              >
                <Trash2 size={14} /> 删除 {selectedIDs.size > 0 ? `(${selectedIDs.size})` : ''}
              </button>
              <button
                onClick={toggleSelectMode}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              >
                取消
              </button>
            </>
          ) : (
            <>
              <button
                onClick={() => CheckUpdates()}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              ><RefreshCw size={14} /> 检查更新</button>
              <button
                onClick={toggleSelectMode}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              ><CheckSquare size={14} /> 批量删除</button>
              <button
                onClick={handleImportButton}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
              ><FolderOpen size={14} /> 手动导入</button>
              <button
                onClick={() => setShowGitHub(true)}
                className="flex items-center gap-1.5 px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 rounded-lg"
              ><Github size={14} /> 从 GitHub 安装</button>
            </>
          )}
        </div>

        {/* Skills grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
            {filtered.map(sk => (
              <SkillCard
                key={sk.ID}
                skill={{ id: sk.ID, name: sk.Name, category: sk.Category, source: sk.Source, hasUpdate: !!sk.LatestSHA }}
                categories={categories}
                onDelete={async () => { await DeleteSkill(sk.ID); load() }}
                onUpdate={async () => { await UpdateSkill(sk.ID); load() }}
                onMoveCategory={async cat => { await MoveSkillCategory(sk.ID, cat); load() }}
                selectMode={selectMode}
                selected={selectedIDs.has(sk.ID)}
                onToggleSelect={() => toggleSelectID(sk.ID)}
                onHoverStart={rect => handleHoverStart(sk, rect)}
                onHoverEnd={handleHoverEnd}
              />
            ))}
          </div>
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <p className="text-sm">没有找到 Skills</p>
              <p className="text-xs mt-1">从 GitHub 安装或拖拽文件夹到此处</p>
            </div>
          )}
        </div>
      </div>

      {hoveredSkill && (
        <SkillTooltip
          skill={hoveredSkill.skill}
          meta={hoveredMeta}
          anchorRect={hoveredSkill.rect}
        />
      )}

      {showGitHub && (
        <GitHubInstallDialog onClose={() => setShowGitHub(false)} onDone={() => { setShowGitHub(false); load() }} />
      )}
    </div>
  )
}
