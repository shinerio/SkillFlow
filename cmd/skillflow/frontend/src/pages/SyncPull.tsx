import { useEffect, useState } from 'react'
import { GetEnabledTools, ScanToolSkills, PullFromTool, PullFromToolForce, ListCategories } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowDownToLine, AlertCircle, X, CheckSquare, Square } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

export default function SyncPull() {
  const defaultCategory = 'Default'
  const [tools, setTools] = useState<any[]>([])
  const [selectedTool, setSelectedTool] = useState('')
  const [scanned, setScanned] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [targetCategory, setTargetCategory] = useState(defaultCategory)
  const [scanning, setScanning] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [conflicts, setConflicts] = useState<string[]>([])
  const [done, setDone] = useState(false)
  const [scanError, setScanError] = useState('')
  const [scannedOnce, setScannedOnce] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListCategories()]).then(([t, c]) => {
      setTools(t ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const scan = async (toolName: string) => {
    setScanning(true)
    setScanned([])
    setScanError('')
    setDone(false)
    try {
      const skills = await ScanToolSkills(toolName)
      setScanned(skills ?? [])
      setSelected(new Set((skills ?? []).map((s: any) => s.Name)))
      setScannedOnce(true)
    } catch (e: any) {
      setScanError(String(e?.message ?? e))
    } finally {
      setScanning(false)
    }
  }

  const pull = async () => {
    setPulling(true)
    const names = [...selected]
    const result = await PullFromTool(selectedTool, names, targetCategory)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPulling(false)
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  const toggleAll = () => {
    if (selected.size === scanned.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(scanned.map((s: any) => s.Name)))
    }
  }

  const allSelected = scanned.length > 0 && selected.size === scanned.length

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 border-r border-gray-800 p-3 flex flex-col gap-0.5">
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide text-gray-500 uppercase">
          导入分类
        </div>
        {categories.map(category => (
          <button
            key={category}
            onClick={() => setTargetCategory(category)}
            className={`px-3 py-2 rounded-lg text-sm text-left transition-colors ${
              targetCategory === category ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-800 flex flex-col gap-4">
          <div className="flex items-center gap-2 text-lg font-semibold">
            <ArrowDownToLine size={18} />
            从工具拉取
          </div>

          <section>
            <p className="text-sm text-gray-400 mb-3">来源工具</p>
            <div className="flex flex-wrap gap-2">
              {tools.map(t => (
                <button
                  key={t.name}
                  onClick={() => {
                    setSelectedTool(t.name)
                    setScanned([])
                    setDone(false)
                    setScanError('')
                    setScannedOnce(false)
                    scan(t.name)
                  }}
                  className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors ${
                    selectedTool === t.name
                      ? 'bg-indigo-600 border-indigo-500 text-white'
                      : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                  }`}
                >
                  <ToolIcon name={t.name} size={20} />
                  {t.name}
                </button>
              ))}
            </div>
          </section>

          {scanning && (
            <p className="text-sm text-gray-400">扫描中...</p>
          )}

          {scanError && (
            <div className="flex items-start gap-2 bg-red-950 border border-red-700 text-red-300 rounded-lg px-4 py-3 text-sm">
              <AlertCircle size={16} className="mt-0.5 shrink-0 text-red-400" />
              <span className="flex-1">{scanError}</span>
              <button onClick={() => setScanError('')} className="shrink-0 text-red-500 hover:text-red-300">
                <X size={14} />
              </button>
            </div>
          )}

          {!scanError && !scanning && scannedOnce && scanned.length === 0 && (
            <div className="flex items-center gap-2 bg-yellow-950 border border-yellow-700 text-yellow-300 rounded-lg px-4 py-3 text-sm">
              <AlertCircle size={16} className="shrink-0 text-yellow-400" />
              <span>未发现任何 Skill，请确认工具目录中包含含有 skill.md 的子目录</span>
            </div>
          )}

          {scanned.length > 0 && (
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-4">
                <p className="text-sm text-gray-400">
                  选择要导入的 Skills
                  <span className="ml-1 text-gray-500">（{selected.size}/{scanned.length}）</span>
                </p>
                <button
                  onClick={toggleAll}
                  className="flex items-center gap-1.5 text-xs text-gray-400 hover:text-white transition-colors"
                >
                  {allSelected ? <CheckSquare size={13} /> : <Square size={13} />}
                  {allSelected ? '取消全选' : '全选'}
                </button>
              </div>
              <p className="text-sm text-gray-400">
                导入到分类「<span className="text-gray-200">{targetCategory || defaultCategory}</span>」
              </p>
            </div>
          )}
        </div>

        {scanned.length > 0 && (
          <>
            <div className="flex-1 overflow-y-auto p-6">
              <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
                {scanned.map((sk: any) => (
                  <SyncSkillCard
                    key={sk.Name}
                    name={sk.Name}
                    path={sk.Path}
                    selected={selected.has(sk.Name)}
                    onToggle={() => toggle(sk.Name)}
                  />
                ))}
              </div>
            </div>

            <div className="px-6 py-4 border-t border-gray-800 flex items-center gap-4">
              <button
                onClick={pull}
                disabled={pulling || selected.size === 0}
                className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
              >
                {pulling ? '拉取中...' : `开始拉取 (${selected.size})`}
              </button>
              {done && <span className="text-sm text-green-400">拉取完成 ✓</span>}
            </div>
          </>
        )}
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            await PullFromToolForce(selectedTool, [name], targetCategory)
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
